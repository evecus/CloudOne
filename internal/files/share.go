package files

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/cloudone/cloudone/internal/auth"
)

type ShareManager struct {
	db *auth.DB
}

func NewShareManager(db *auth.DB) *ShareManager {
	return &ShareManager{db: db}
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// randCode 使用 crypto/rand 生成密码学安全的随机分享码
func randCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic("share: crypto/rand failed: " + err.Error())
		}
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}

// Create 创建分享链接，支持过期时间（nil=永不过期）和最大访问次数（0=不限）
func (s *ShareManager) Create(userID uint, filePath string, isDir bool, expiresAt *time.Time, maxViews int) (*auth.ShareLink, error) {
	code := randCode(8)
	link := &auth.ShareLink{
		Code:      code,
		FilePath:  filePath,
		IsDir:     isDir,
		UserID:    userID,
		ExpiresAt: expiresAt,
		MaxViews:  maxViews,
	}
	if err := s.db.CreateShare(link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *ShareManager) List(userID uint) ([]auth.ShareLink, error) {
	return s.db.ListSharesByUser(userID)
}

func (s *ShareManager) Delete(id uint, userID uint) error {
	return s.db.DeleteShare(id, userID)
}

// Peek 只读取分享链接（校验过期/次数），不递增访问计数，用于目录浏览
func (s *ShareManager) Peek(code string) (*auth.ShareLink, error) {
	link, err := s.db.GetShareByCode(code)
	if err != nil {
		return nil, err
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, fmt.Errorf("share link has expired")
	}
	if link.MaxViews > 0 && link.ViewCount >= link.MaxViews {
		return nil, fmt.Errorf("share link has reached maximum views")
	}
	return link, nil
}

// Get 获取分享链接，同时校验过期时间和访问次数，有效则递增计数
func (s *ShareManager) Get(code string) (*auth.ShareLink, error) {
	link, err := s.db.GetShareByCode(code)
	if err != nil {
		return nil, err
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, fmt.Errorf("share link has expired")
	}
	if link.MaxViews > 0 && link.ViewCount >= link.MaxViews {
		return nil, fmt.Errorf("share link has reached maximum views")
	}
	link.ViewCount++
	_ = s.db.UpdateShareViewCount(link)
	return link, nil
}
