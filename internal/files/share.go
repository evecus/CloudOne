package files

import (
	"crypto/rand"
	"math/big"

	"github.com/cloudone/cloudone/internal/auth"
	"gorm.io/gorm"
)

type ShareManager struct {
	db *gorm.DB
}

func NewShareManager(db *gorm.DB) *ShareManager {
	return &ShareManager{db: db}
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// randCode 使用 crypto/rand 生成密码学安全的随机分享码，防止枚举攻击。
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

func (s *ShareManager) Create(userID uint, filePath string, isDir bool) (*auth.ShareLink, error) {
	code := randCode(8)
	link := &auth.ShareLink{
		Code:     code,
		FilePath: filePath,
		IsDir:    isDir,
		UserID:   userID,
	}
	if err := s.db.Create(link).Error; err != nil {
		return nil, err
	}
	return link, nil
}

func (s *ShareManager) List(userID uint) ([]auth.ShareLink, error) {
	var links []auth.ShareLink
	s.db.Where("user_id = ?", userID).Find(&links)
	return links, nil
}

func (s *ShareManager) Delete(id uint, userID uint) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&auth.ShareLink{}).Error
}

func (s *ShareManager) Get(code string) (*auth.ShareLink, error) {
	var link auth.ShareLink
	if err := s.db.Where("code = ?", code).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}
