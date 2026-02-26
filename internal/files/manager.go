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

// skipDirs 存储根为 / 时跳过的危险虚拟文件系统目录。
var skipDirs = map[string]struct{}{
	"/proc": {},
	"/sys":  {},
	"/dev":  {},
	"/run":  {},
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
	db   *gorm.DB
}

func NewManager(root string, db *gorm.DB) *Manager {
	return &Manager{root: root, db: db}
}

func (m *Manager) SetRoot(root string) { m.root = root }
func (m *Manager) Root() string        { return m.root }

// normalizePath 统一路径格式：以 / 开头，无尾随 /，无重复 /
func normalizePath(p string) string {
	trimmed := strings.TrimLeft(p, "/")
	if trimmed == "" {
		return "/"
	}
	return filepath.Clean("/" + trimmed)
}

// AbsPath 将逻辑路径转为磁盘绝对路径，并验证不超出 root（防路径穿越）。
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

// ── 可见性核心逻辑 ────────────────────────────────────────────────────────────
//
// 设计原则：
//   公开性是对路径本身的「显式标记」，不存在父目录继承。
//   设置公开时，对该路径及其所有子路径逐一写入 is_public=true 记录。
//   设置私有时，删除该路径及所有子路径的所有记录（回归默认私有）。
//   移动/复制：可见性记录随文件一起迁移，不丢失也不新增。
//   新建/上传：新路径没有记录，默认私有，不受任何已有记录影响。
//   删除：先删文件，再清除所有相关记录，保持 DB 干净。

// IsPublic 判断路径是否有精确的 is_public=true 记录。
func (m *Manager) IsPublic(rel string) bool {
	norm := normalizePath(rel)
	var count int64
	m.db.Model(&auth.FileVisibility{}).
		Where("file_path = ? AND is_public = ?", norm, true).
		Count(&count)
	return count > 0
}

// SetVisibility 设置文件/目录的公开状态。
//   isPublic=true  → 对该路径及文件系统中所有子路径写入 is_public=true
//   isPublic=false → 删除该路径及所有子路径的所有记录（恢复默认私有）
func (m *Manager) SetVisibility(rel string, isPublic bool) error {
	norm := normalizePath(rel)
	likePattern := norm + "/%"
	if norm == "/" {
		likePattern = "/%"
	}

	if !isPublic {
		// 私有：直接删除所有相关记录
		m.db.Where("file_path = ? OR file_path LIKE ?", norm, likePattern).
			Delete(&auth.FileVisibility{})
		return nil
	}

	// 公开：对该路径及文件系统所有子路径写入显式标记
	return m.markPublicRecursive(norm)
}

// markPublicRecursive 递归地为 norm 及其文件系统子路径写入 is_public=true。
func (m *Manager) markPublicRecursive(norm string) error {
	if err := m.upsertVisibility(norm, true); err != nil {
		return err
	}
	abs, err := m.AbsPath(norm)
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

// upsertVisibility 写入或更新单条可见性记录。
func (m *Manager) upsertVisibility(path string, isPublic bool) error {
	var vis auth.FileVisibility
	if m.db.Where("file_path = ?", path).First(&vis).Error != nil {
		return m.db.Create(&auth.FileVisibility{FilePath: path, IsPublic: isPublic}).Error
	}
	vis.IsPublic = isPublic
	return m.db.Save(&vis).Error
}

// deleteVisibilityTree 删除 norm 及其所有子路径的可见性记录。
func (m *Manager) deleteVisibilityTree(norm string) {
	likePattern := norm + "/%"
	if norm == "/" {
		likePattern = "/%"
	}
	m.db.Where("file_path = ? OR file_path LIKE ?", norm, likePattern).
		Delete(&auth.FileVisibility{})
}

// migrateVisibility 将 srcNorm 及子路径的记录迁移到 dstNorm（用于 Move）。
func (m *Manager) migrateVisibility(srcNorm, dstNorm string) {
	likePattern := srcNorm + "/%"
	if srcNorm == "/" {
		likePattern = "/%"
	}
	var records []auth.FileVisibility
	m.db.Where("file_path = ? OR file_path LIKE ?", srcNorm, likePattern).Find(&records)
	if len(records) == 0 {
		return
	}
	// 先删旧记录，再用新路径写入
	m.db.Where("file_path = ? OR file_path LIKE ?", srcNorm, likePattern).
		Delete(&auth.FileVisibility{})
	for _, r := range records {
		newPath := dstNorm
		if r.FilePath != srcNorm {
			newPath = dstNorm + strings.TrimPrefix(r.FilePath, srcNorm)
		}
		m.upsertVisibility(newPath, r.IsPublic)
	}
}

// copyVisibility 将 srcNorm 及子路径的记录复制到 dstNorm（用于 Copy）。
func (m *Manager) copyVisibility(srcNorm, dstNorm string) {
	likePattern := srcNorm + "/%"
	if srcNorm == "/" {
		likePattern = "/%"
	}
	var records []auth.FileVisibility
	m.db.Where("file_path = ? OR file_path LIKE ?", srcNorm, likePattern).Find(&records)
	for _, r := range records {
		newPath := dstNorm
		if r.FilePath != srcNorm {
			newPath = dstNorm + strings.TrimPrefix(r.FilePath, srcNorm)
		}
		m.upsertVisibility(newPath, r.IsPublic)
	}
}

// ── 文件操作 ──────────────────────────────────────────────────────────────────

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

	// 批量查询子项的可见性（精确匹配，无继承）
	childPaths := make([]string, 0, len(entries))
	for _, e := range entries {
		childPaths = append(childPaths, normalizePath(norm+"/"+e.Name()))
	}
	visMap := make(map[string]bool)
	if len(childPaths) > 0 {
		var visList []auth.FileVisibility
		m.db.Where("file_path IN ? AND is_public = ?", childPaths, true).Find(&visList)
		for _, v := range visList {
			visMap[v.FilePath] = true
		}
	}

	result := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			if _, skip := skipDirs["/"+e.Name()]; skip && abs == "/" {
				continue
			}
		}
		info, _ := e.Info()
		childRel := normalizePath(norm + "/" + e.Name())
		fi := FileInfo{
			Name:     e.Name(),
			Path:     childRel,
			IsDir:    e.IsDir(),
			IsPublic: visMap[childRel],
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

// Delete 删除文件/目录，并清除所有相关可见性记录。
func (m *Manager) Delete(rel string) error {
	if rel == "" || rel == "/" {
		return errors.New("cannot delete root directory")
	}
	abs, err := m.AbsPath(rel)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(abs); err != nil {
		return err
	}
	m.deleteVisibilityTree(normalizePath(rel))
	return nil
}

// Move 移动文件/目录，可见性记录随之迁移。
func (m *Manager) Move(src, dst string) error {
	srcAbs, err := m.AbsPath(src)
	if err != nil {
		return err
	}
	dstAbs, err := m.AbsPath(dst)
	if err != nil {
		return err
	}
	if err := os.Rename(srcAbs, dstAbs); err != nil {
		return err
	}
	m.migrateVisibility(normalizePath(src), normalizePath(dst))
	return nil
}

// Copy 复制文件/目录，可见性记录随之复制。
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

// ReadContent 读取文件内容，最多 2MB。
// 使用 io.LimitReader + io.ReadAll，避免单次 Read 不保证读满的问题。
func (m *Manager) ReadContent(rel string) (string, error) {
	abs, err := m.AbsPath(rel)
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

// ListPublic 返回目录下有 is_public=true 标记的直接子条目（用于已登录管理界面浏览某目录时）。
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

// GetAllPublicFlat 返回所有公开条目的平铺列表，规则：
//
//   - 从 DB 取出全部 is_public=true 的路径
//   - 若某路径的直接父路径也有 is_public=true 记录，则跳过该路径
//     （它会在父文件夹被浏览时展示，不需要在顶层重复出现）
//   - 若某路径的直接父路径没有 is_public=true 记录，则直接列在顶层
//
// 效果：无论文件/文件夹在多深层级都直接浮现到顶层，不显示任何父文件夹。
// 被整体设为公开的文件夹，其内部子项不会重复出现在顶层。
func (m *Manager) GetAllPublicFlat() ([]FileInfo, error) {
	// 1. 取出所有 is_public=true 的路径
	var records []auth.FileVisibility
	if err := m.db.Where("is_public = ?", true).Find(&records).Error; err != nil {
		return nil, err
	}

	// 2. 建立快速查找集合
	publicSet := make(map[string]struct{}, len(records))
	for _, r := range records {
		publicSet[r.FilePath] = struct{}{}
	}

	// 3. 过滤：只保留「父路径没有 is_public=true 记录」的条目
	result := make([]FileInfo, 0)
	for _, r := range records {
		parentPath := filepath.Dir(r.FilePath)
		if parentPath == "." {
			parentPath = "/"
		}
		// 父路径也是公开的，跳过（避免重复，由父文件夹包含）
		if _, parentPublic := publicSet[parentPath]; parentPublic {
			continue
		}

		// 4. 从文件系统获取该条目的实际信息
		abs, err := m.AbsPath(r.FilePath)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			// 文件已不存在，清理这条脏记录
			m.db.Delete(&r)
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

// SearchFiles 在指定目录（含子目录）中按文件名模糊搜索（大小写不敏感，包含匹配）。
// 结果上限 200 条，超过后停止遍历。
func (m *Manager) SearchFiles(dir, keyword string) ([]SearchResult, error) {
	absDir, err := m.AbsPath(dir)
	if err != nil {
		return nil, err
	}
	kwLower := strings.ToLower(keyword)
	root := m.root
	var results []SearchResult

	err = filepath.WalkDir(absDir, func(p string, d os.DirEntry, werr error) error {
		if werr != nil {
			return nil // 跳过无权限目录
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

type SearchResult struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
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
