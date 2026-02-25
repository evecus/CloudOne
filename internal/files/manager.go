package files

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudone/cloudone/internal/auth"
	"gorm.io/gorm"
)

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
	db   *gorm.DB
}

func NewManager(root string, db *gorm.DB) *Manager {
	return &Manager{root: root, db: db}
}

func (m *Manager) SetRoot(root string) { m.root = root }
func (m *Manager) Root() string        { return m.root }

// normalizePath 统一路径格式：确保以 / 开头，无尾随 /，无重复 /
func normalizePath(p string) string {
	// 去掉所有前导 /，然后加上唯一的 /
	trimmed := strings.TrimLeft(p, "/")
	if trimmed == "" {
		return "/"
	}
	// 用 filepath.Clean 处理 . 和 ..
	cleaned := filepath.Clean("/" + trimmed)
	return cleaned
}

// AbsPath 将相对路径转为绝对路径，验证是否在 root 下（防止路径穿越攻击）。
func (m *Manager) AbsPath(rel string) (string, error) {
	norm := normalizePath(rel)
	abs := filepath.Join(m.root, norm)
	rootClean := filepath.Clean(m.root)
	// 当 root 为文件系统根目录 "/" 时，所有路径都合法，无需额外检查
	if rootClean != string(os.PathSeparator) {
		if abs != rootClean && !strings.HasPrefix(abs, rootClean+string(os.PathSeparator)) {
			return "", errors.New("invalid path: path traversal detected")
		}
	}
	return abs, nil
}

// ancestorPaths 返回路径本身及所有祖先路径（含根 "/"）。
// 例如 "/a/b/c" -> ["/a/b/c", "/a/b", "/a", "/"]
func ancestorPaths(norm string) []string {
	paths := []string{norm}
	parts := strings.Split(strings.Trim(norm, "/"), "/")
	for i := len(parts) - 1; i >= 1; i-- {
		paths = append(paths, "/"+strings.Join(parts[:i], "/"))
	}
	if norm != "/" {
		paths = append(paths, "/")
	}
	return paths
}

// isPublicPath 判断路径是否可公开访问。
// 规则（按优先级）：
//  1. 路径自身有明确的 is_public=false 记录 → 私有（优先级最高）
//  2. 路径自身有 is_public=true → 公开
//  3. 任意祖先目录有 is_public=true 且中间没有 is_public=false 的中断 → 公开（继承）
func (m *Manager) isPublicPath(norm string) bool {
	// 收集路径本身和所有祖先
	paths := ancestorPaths(norm)

	// 批量查询所有相关记录（一次 SQL）
	var records []auth.FileVisibility
	m.db.Where("file_path IN ?", paths).Find(&records)

	// 建立 map：path -> is_public
	visMap := make(map[string]bool, len(records))
	for _, r := range records {
		visMap[r.FilePath] = r.IsPublic
	}

	// 从路径自身向上逐级检查
	for _, p := range paths {
		val, exists := visMap[p]
		if !exists {
			continue // 该层没有记录，继续向上
		}
		if !val {
			return false // 该层明确设为私有，截断继承
		}
		return true // 该层明确设为公开
	}

	return false // 所有层都没有记录，默认私有
}

func (m *Manager) ListDir(rel string) ([]FileInfo, error) {
	abs, err := m.AbsPath(rel)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}

	norm := normalizePath(rel)

	// 构建子路径列表
	childPaths := make([]string, 0, len(entries))
	for _, e := range entries {
		childPaths = append(childPaths, normalizePath(norm+"/"+e.Name()))
	}

	// 批量查询子项自身的可见性记录
	selfVisMap := make(map[string]bool)
	if len(childPaths) > 0 {
		var visList []auth.FileVisibility
		m.db.Where("file_path IN ?", childPaths).Find(&visList)
		for _, v := range visList {
			selfVisMap[v.FilePath] = v.IsPublic
		}
	}

	// 判断父目录（当前 rel）是否公开（供子项继承）
	parentPublic := m.isPublicPath(norm)

	result := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		info, _ := e.Info()
		childRel := normalizePath(norm + "/" + e.Name())

		// 子项公开状态：有自己的记录则用自己的，否则继承父目录
		var isPublic bool
		if selfVal, hasSelf := selfVisMap[childRel]; hasSelf {
			if !selfVal {
				// 子项明确设为私有
				isPublic = false
			} else {
				isPublic = true
			}
		} else {
			isPublic = parentPublic
		}

		fi := FileInfo{
			Name:     e.Name(),
			Path:     childRel,
			IsDir:    e.IsDir(),
			IsPublic: isPublic,
		}
		if info != nil {
			fi.Size = info.Size()
			fi.ModTime = info.ModTime()
			fi.Mode = info.Mode()
		}
		result = append(result, fi)
	}
	return result, nil
}

func (m *Manager) MkDir(rel string) error {
	abs, err := m.AbsPath(rel)
	if err != nil {
		return err
	}
	return os.MkdirAll(abs, 0755)
}

func (m *Manager) Delete(rel string) error {
	if rel == "" || rel == "/" {
		return errors.New("cannot delete root directory")
	}
	abs, err := m.AbsPath(rel)
	if err != nil {
		return err
	}
	return os.RemoveAll(abs)
}

func (m *Manager) Move(src, dst string) error {
	srcAbs, err := m.AbsPath(src)
	if err != nil {
		return err
	}
	dstAbs, err := m.AbsPath(dst)
	if err != nil {
		return err
	}
	return os.Rename(srcAbs, dstAbs)
}

func (m *Manager) Copy(src, dst string) error {
	srcAbs, err := m.AbsPath(src)
	if err != nil {
		return err
	}
	dstAbs, err := m.AbsPath(dst)
	if err != nil {
		return err
	}
	info, err := os.Stat(srcAbs)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(srcAbs, dstAbs)
	}
	return copyFile(srcAbs, dstAbs)
}

func (m *Manager) Write(rel string, r io.Reader) error {
	abs, err := m.AbsPath(rel)
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
	abs, err := m.AbsPath(rel)
	if err != nil {
		return "", err
	}
	const maxSize = 2 * 1024 * 1024
	f, err := os.Open(abs)
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := make([]byte, maxSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	return string(buf[:n]), nil
}

func (m *Manager) GetPermission(rel string) (os.FileMode, error) {
	abs, err := m.AbsPath(rel)
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
	abs, err := m.AbsPath(rel)
	if err != nil {
		return err
	}
	return os.Chmod(abs, mode)
}

func (m *Manager) Open(rel string) (*os.File, error) {
	abs, err := m.AbsPath(rel)
	if err != nil {
		return nil, err
	}
	return os.Open(abs)
}

// IsPublic 判断路径是否可公开访问（含继承和私有覆盖逻辑）。
func (m *Manager) IsPublic(rel string) bool {
	return m.isPublicPath(normalizePath(rel))
}

// SetVisibility 设置文件/目录的公开状态（级联逻辑）。
//
//   - 设为公开：只写该路径自身的记录；子文件通过 isPublicPath 继承
//   - 设为私有：删除该路径及所有子路径的记录，并写入 is_public=false
//     （即使祖先是公开的，也会被本条记录截断）
func (m *Manager) SetVisibility(rel string, isPublic bool) error {
	norm := normalizePath(rel)

	if isPublic {
		return m.upsertVisibility(norm, true)
	}

	// 设为私有：删除自身和所有子路径的记录
	likePattern := ""
	if norm == "/" {
		likePattern = "/%"
	} else {
		likePattern = norm + "/%"
	}
	m.db.Where("file_path = ? OR file_path LIKE ?", norm, likePattern).
		Delete(&auth.FileVisibility{})

	// 写入明确的 is_public=false（阻断祖先继承）
	return m.upsertVisibility(norm, false)
}

func (m *Manager) upsertVisibility(path string, isPublic bool) error {
	var vis auth.FileVisibility
	if m.db.Where("file_path = ?", path).First(&vis).Error != nil {
		vis = auth.FileVisibility{FilePath: path, IsPublic: isPublic}
		return m.db.Create(&vis).Error
	}
	vis.IsPublic = isPublic
	return m.db.Save(&vis).Error
}

// ListPublic 返回指定目录下所有"可公开访问"的条目。
// 包含：自身公开的文件/目录，以及虽然自身私有但内部有公开文件的目录。
func (m *Manager) ListPublic(rel string) ([]FileInfo, error) {
	all, err := m.ListDir(rel)
	if err != nil {
		return nil, err
	}

	result := make([]FileInfo, 0)
	for _, f := range all {
		if f.IsPublic {
			result = append(result, f)
			continue
		}
		// 私有目录：检查内部是否有公开文件（任意深度）
		if f.IsDir {
			if m.hasPublicDescendant(f.Path) {
				result = append(result, f) // is_public=false 但含公开子文件
			}
		}
	}
	return result, nil
}

// hasPublicDescendant 检查目录下是否存在任何公开的后代。
func (m *Manager) hasPublicDescendant(dirPath string) bool {
	norm := normalizePath(dirPath)
	likePattern := norm + "/%"
	if norm == "/" {
		likePattern = "/%"
	}

	var count int64
	m.db.Model(&auth.FileVisibility{}).
		Where("(file_path = ? OR file_path LIKE ?) AND is_public = ?",
			norm, likePattern, true).
		Count(&count)
	return count > 0
}

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
