package files

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudone/cloudone/internal/auth"
	"gorm.io/gorm"
)

type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	IsPublic  bool      `json:"is_public"`
}

type Manager struct {
	root string
	db   *gorm.DB
}

func NewManager(root string, db *gorm.DB) *Manager {
	return &Manager{root: root, db: db}
}

func (m *Manager) AbsPath(rel string) string {
	cleaned := filepath.Clean("/" + rel)
	return filepath.Join(m.root, cleaned)
}

func (m *Manager) ListDir(rel string) ([]FileInfo, error) {
	abs := m.AbsPath(rel)
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}

	var result []FileInfo
	for _, e := range entries {
		info, _ := e.Info()
		childRel := filepath.Join(rel, e.Name())
		var vis auth.FileVisibility
		isPublic := false
		if m.db.Where("file_path = ?", childRel).First(&vis).Error == nil {
			isPublic = vis.IsPublic
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
		}
		result = append(result, fi)
	}
	return result, nil
}

func (m *Manager) MkDir(rel string) error {
	return os.MkdirAll(m.AbsPath(rel), 0755)
}

func (m *Manager) Delete(rel string) error {
	return os.RemoveAll(m.AbsPath(rel))
}

func (m *Manager) Move(src, dst string) error {
	return os.Rename(m.AbsPath(src), m.AbsPath(dst))
}

func (m *Manager) Copy(src, dst string) error {
	srcAbs := m.AbsPath(src)
	dstAbs := m.AbsPath(dst)

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
	abs := m.AbsPath(rel)
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

func (m *Manager) Open(rel string) (*os.File, error) {
	return os.Open(m.AbsPath(rel))
}

func (m *Manager) SetVisibility(rel string, isPublic bool) error {
	var vis auth.FileVisibility
	if m.db.Where("file_path = ?", rel).First(&vis).Error != nil {
		vis = auth.FileVisibility{FilePath: rel, IsPublic: isPublic}
		return m.db.Create(&vis).Error
	}
	vis.IsPublic = isPublic
	return m.db.Save(&vis).Error
}

func (m *Manager) IsPublic(rel string) bool {
	var vis auth.FileVisibility
	if m.db.Where("file_path = ?", rel).First(&vis).Error == nil {
		return vis.IsPublic
	}
	return false
}

func (m *Manager) ListPublic(rel string) ([]FileInfo, error) {
	all, err := m.ListDir(rel)
	if err != nil {
		return nil, err
	}
	var result []FileInfo
	for _, f := range all {
		if f.IsPublic {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *Manager) Root() string { return m.root }

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
