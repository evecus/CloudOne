package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ── 加密工具 ──────────────────────────────────────────────────────────────────
// 所有敏感字段（WebDAV 密码哈希、JWT Secret）均使用 AES-256-GCM 加密后
// 以 base64 字符串存入数据库，解密密钥由启动时传入的 masterKey 派生。

var masterKey []byte // 32 字节，由 main.go 在启动时调用 SetMasterKey 设置

// SetMasterKey 接收任意长度字符串，SHA-256 派生为 32 字节密钥
func SetMasterKey(key string) {
	h := sha256.Sum256([]byte(key))
	masterKey = h[:]
}

// Encrypt 使用 AES-256-GCM 加密明文，返回 base64 编码的密文
func Encrypt(plaintext string) (string, error) {
	if len(masterKey) == 0 {
		return "", errors.New("master key not set")
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Encrypt 产生的 base64 密文
func Decrypt(encoded string) (string, error) {
	if len(masterKey) == 0 {
		return "", errors.New("master key not set")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// ── 数据模型 ──────────────────────────────────────────────────────────────────

// User 存储账户信息。Password 字段存 bcrypt 哈希（非敏感，bcrypt 本身不可逆）。
type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Username     string    `gorm:"uniqueIndex" json:"username"`
	Password     string    `json:"-"` // bcrypt hash，不通过 API 暴露
	TokenVersion int       `json:"-"` // 密码修改时递增，使旧 token 立即失效
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Settings 系统设置。
// JWTSecretEnc      — AES-GCM 加密的 JWT 签名密钥（base64）
// WebDAVPasswordEnc — AES-GCM 加密的 WebDAV 独立密码 bcrypt 哈希（base64）
// WebDAVUsername    — 非敏感，明文存储
type Settings struct {
	ID               uint   `gorm:"primarykey"`
	StorageDir       string `json:"storage_dir"`
	Lang             string `json:"lang"`
	UITheme          string `json:"ui_theme"`
	UIFont           string `json:"ui_font"`
	EditorFont       string `json:"editor_font"`
	WebDAVEnabled    bool   `json:"webdav_enabled"`
	WebDAVSubPath    string `json:"webdav_sub_path"`
	WebDAVUsername   string `json:"webdav_username"`
	WebDAVPasswordEnc string `json:"-"` // AES-GCM(bcrypt(password))，不暴露
	JWTSecretEnc     string `json:"-"` // AES-GCM(jwt_secret)，不暴露
	ShowHidden       bool   `json:"show_hidden"`
}

// GetJWTSecret 解密并返回 JWT 密钥明文
func (s *Settings) GetJWTSecret() (string, error) {
	if s.JWTSecretEnc == "" {
		return "", nil
	}
	return Decrypt(s.JWTSecretEnc)
}

// SetJWTSecret 加密并存储 JWT 密钥
func (s *Settings) SetJWTSecret(secret string) error {
	enc, err := Encrypt(secret)
	if err != nil {
		return err
	}
	s.JWTSecretEnc = enc
	return nil
}

// GetWebDAVPasswordHash 解密并返回 WebDAV 密码的 bcrypt 哈希
func (s *Settings) GetWebDAVPasswordHash() (string, error) {
	if s.WebDAVPasswordEnc == "" {
		return "", nil
	}
	return Decrypt(s.WebDAVPasswordEnc)
}

// SetWebDAVPasswordHash 加密并存储 WebDAV 密码的 bcrypt 哈希
func (s *Settings) SetWebDAVPasswordHash(bcryptHash string) error {
	if bcryptHash == "" {
		s.WebDAVPasswordEnc = ""
		return nil
	}
	enc, err := Encrypt(bcryptHash)
	if err != nil {
		return err
	}
	s.WebDAVPasswordEnc = enc
	return nil
}

// ShareLink 分享链接，增加过期时间和访问次数限制
type ShareLink struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Code      string     `gorm:"uniqueIndex" json:"code"`
	FilePath  string     `json:"file_path"`
	IsDir     bool       `json:"is_dir"`
	UserID    uint       `json:"user_id"`
	ExpiresAt *time.Time `json:"expires_at"` // nil = 永不过期
	MaxViews  int        `json:"max_views"`  // 0 = 不限次数
	ViewCount int        `json:"view_count"`
	CreatedAt time.Time  `json:"created_at"`
}

// FileVisibility 文件可见性记录
type FileVisibility struct {
	ID       uint   `gorm:"primarykey"`
	FilePath string `gorm:"uniqueIndex" json:"file_path"`
	IsPublic bool   `json:"is_public"`
}

// ── 数据库初始化 ───────────────────────────────────────────────────────────────

func InitDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{}, &Settings{}, &ShareLink{}, &FileVisibility{})

	var count int64
	db.Model(&Settings{}).Count(&count)
	if count == 0 {
		db.Create(&Settings{StorageDir: "./data/storage", Lang: "zh"})
	}

	return db, nil
}
