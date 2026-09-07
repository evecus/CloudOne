package files

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudone/cloudone/internal/auth"
)

var dangerPrefixes = []string{"/proc", "/sys", "/dev"}

func isDangerPath(absPath string) bool {
	clean := filepath.Clean(absPath)
	for _, p := range dangerPrefixes {
		if clean == p || strings.HasPrefix(clean, p+"/") {
			return true
		}
	}
	return false
}

type FileInfo struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"is_dir"`
	Size     int64       `json:"size"`
	ModTime  time.Time   `json:"mod_time"`
	IsPublic bool        `json:"is_public"`
	Mode     os.FileMode `json:"mode"`
}

type Manager struct {
	root string
	db   *auth.DB
}

func NewManager(root string, db *auth.DB) *Manager {
	return &Manager{root: root, db: db}
}

func (m *Manager) SetRoot(root string) { m.root = root }
func (m *Manager) Root() string        { return m.root }

func normalizePath(p string) string {
	trimmed := strings.TrimLeft(p, "/")
	if trimmed == "" {
		return "/"
	}
	return filepath.Clean("/" + trimmed)
}

func (m *Manager) AbsPath(rel string) (string, error) {
	norm := normalizePath(rel)
	abs := filepath.Join(m.root, norm)
	rootClean := filepath.Clean(m.root)
	if rootClean != string(os.PathSeparator) {
		if abs != rootClean && !strings.HasPrefix(abs, rootClean+string(os.PathSeparator)) {
			return "", errors.New("invalid path: path traversal detected")
		}
	}
	return abs, nil
}

func (m *Manager) SafeAbsPath(rel string) (string, error) {
	abs, err := m.AbsPath(rel)
	if err != nil {
		return "", err
	}
	if isDangerPath(abs) {
		return "", errors.New("access denied: virtual filesystem path")
	}
	return abs, nil
}

// ── 可见性核心逻辑 ────────────────────────────────────────────────────────────

func (m *Manager) IsPublic(rel string) bool {
	norm := normalizePath(rel)
	vis, err := m.db.GetVisibility(norm)
	if err != nil {
		return false
	}
	return vis.IsPublic
}

func (m *Manager) SetVisibility(rel string, isPublic bool) error {
	norm := normalizePath(rel)
	if !isPublic {
		return m.db.DeleteVisibilityPrefix(norm)
	}
	return m.markPublicRecursive(norm)
}

func (m *Manager) markPublicRecursive(norm string) error {
	if err := m.db.SetVisibility(norm, true); err != nil {
		return err
	}
	abs, err := m.SafeAbsPath(norm)
	if err != nil {
		return nil
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return nil
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		child := normalizePath(norm + "/" + e.Name())
		if err := m.markPublicRecursive(child); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) deleteVisibilityTree(norm string) {
	_ = m.db.DeleteVisibilityPrefix(norm)
}

func (m *Manager) migrateVisibility(srcNorm, dstNorm string) {
	_ = m.db.RenameVisibilityPrefix(srcNorm, dstNorm)
}

func (m *Manager) copyVisibility(srcNorm, dstNorm string) {
	records, err := m.db.ListPublicVisibilityByPrefix(srcNorm)
	if err != nil {
		return
	}
	for _, r := range records {
		newPath := dstNorm + r.FilePath[len(srcNorm):]
		_ = m.db.SetVisibility(newPath, r.IsPublic)
	}
}

// ── 文件操作 ──────────────────────────────────────────────────────────────────

func (m *Manager) ListDir(rel string) ([]FileInfo, error) {
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	norm := normalizePath(rel)

	childPaths := make([]string, 0, len(entries))
	for _, e := range entries {
		childPaths = append(childPaths, normalizePath(norm+"/"+e.Name()))
	}
	visMap := make(map[string]bool)
	if len(childPaths) > 0 {
		visList, _ := m.db.ListVisibilityByPaths(childPaths)
		for _, v := range visList {
			visMap[v.FilePath] = true
		}
	}

	result := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		childAbs := filepath.Join(abs, e.Name())
		childRel := normalizePath(norm + "/" + e.Name())

		realInfo, statErr := os.Stat(childAbs)
		isDir := e.IsDir()
		if statErr == nil {
			isDir = realInfo.IsDir()
		}
		if isDangerPath(childAbs) {
			continue
		}

		fi := FileInfo{
			Name:     e.Name(),
			Path:     childRel,
			IsDir:    isDir,
			IsPublic: visMap[childRel],
		}
		if realInfo != nil {
			fi.Size = realInfo.Size()
			fi.ModTime = realInfo.ModTime()
			fi.Mode = realInfo.Mode()
		} else if info, _ := e.Info(); info != nil {
			fi.Size = info.Size()
			fi.ModTime = info.ModTime()
			fi.Mode = info.Mode()
		}
		result = append(result, fi)
	}
	return result, nil
}

func (m *Manager) MkDir(rel string) error {
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return err
	}
	return os.MkdirAll(abs, 0755)
}

func (m *Manager) Delete(rel string) error {
	if rel == "" || rel == "/" {
		return errors.New("cannot delete root directory")
	}
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(abs); err != nil {
		return err
	}
	m.deleteVisibilityTree(normalizePath(rel))
	return nil
}

func (m *Manager) Move(src, dst string) error {
	srcAbs, err := m.SafeAbsPath(src)
	if err != nil {
		return err
	}
	dstAbs, err := m.SafeAbsPath(dst)
	if err != nil {
		return err
	}
	if err := os.Rename(srcAbs, dstAbs); err != nil {
		return err
	}
	m.migrateVisibility(normalizePath(src), normalizePath(dst))
	return nil
}

func (m *Manager) Copy(src, dst string) error {
	srcAbs, err := m.SafeAbsPath(src)
	if err != nil {
		return err
	}
	dstAbs, err := m.SafeAbsPath(dst)
	if err != nil {
		return err
	}
	info, err := os.Stat(srcAbs)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := copyDir(srcAbs, dstAbs); err != nil {
			return err
		}
	} else {
		if err := copyFile(srcAbs, dstAbs); err != nil {
			return err
		}
	}
	m.copyVisibility(normalizePath(src), normalizePath(dst))
	return nil
}

func (m *Manager) Write(rel string, r io.Reader) error {
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		return err
	}
	f, err := os.Create(abs)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (m *Manager) ReadContent(rel string) (string, error) {
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return "", err
	}
	f, err := os.Open(abs)
	if err != nil {
		return "", err
	}
	defer f.Close()
	const maxSize = 2 * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(f, maxSize))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *Manager) GetPermission(rel string) (os.FileMode, error) {
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
}

func (m *Manager) Chmod(rel string, mode os.FileMode) error {
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return err
	}
	return os.Chmod(abs, mode)
}

func (m *Manager) Open(rel string) (*os.File, error) {
	abs, err := m.SafeAbsPath(rel)
	if err != nil {
		return nil, err
	}
	return os.Open(abs)
}

func (m *Manager) ListPublic(rel string) ([]FileInfo, error) {
	all, err := m.ListDir(rel)
	if err != nil {
		return nil, err
	}
	result := make([]FileInfo, 0)
	for _, f := range all {
		if f.IsPublic {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *Manager) GetAllPublicFlat() ([]FileInfo, error) {
	records, err := m.db.ListPublicVisibility()
	if err != nil {
		return nil, err
	}

	publicSet := make(map[string]struct{}, len(records))
	for _, r := range records {
		publicSet[r.FilePath] = struct{}{}
	}

	result := make([]FileInfo, 0)
	for _, r := range records {
		parentPath := filepath.Dir(r.FilePath)
		if parentPath == "." {
			parentPath = "/"
		}
		if _, parentPublic := publicSet[parentPath]; parentPublic {
			continue
		}

		abs, err := m.SafeAbsPath(r.FilePath)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			// 文件已不存在，清理脏记录
			_ = m.db.DeleteVisibilityPrefix(r.FilePath)
			continue
		}

		fi := FileInfo{
			Name:     filepath.Base(r.FilePath),
			Path:     r.FilePath,
			IsDir:    info.IsDir(),
			IsPublic: true,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Mode:     info.Mode(),
		}
		if info.IsDir() {
			fi.Size = 0
		}
		result = append(result, fi)
	}
	return result, nil
}

// ── 搜索 ──────────────────────────────────────────────────────────────────────

const searchMaxResults = 200

type SearchResult struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

func (m *Manager) SearchFiles(dir, keyword string) ([]SearchResult, error) {
	absDir, err := m.SafeAbsPath(dir)
	if err != nil {
		return nil, err
	}
	kwLower := strings.ToLower(keyword)
	root := m.root
	var results []SearchResult

	err = filepath.WalkDir(absDir, func(p string, d os.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if d.IsDir() && isDangerPath(p) {
			return filepath.SkipDir
		}
		if p == absDir {
			return nil
		}
		if strings.Contains(strings.ToLower(d.Name()), kwLower) {
			rel, _ := filepath.Rel(root, p)
			rel = "/" + filepath.ToSlash(rel)
			results = append(results, SearchResult{
				Name:  d.Name(),
				Path:  rel,
				IsDir: d.IsDir(),
			})
			if len(results) >= searchMaxResults {
				return filepath.SkipAll
			}
		}
		return nil
	})
	return results, err
}

// ── 内部工具函数 ──────────────────────────────────────────────────────────────

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}
