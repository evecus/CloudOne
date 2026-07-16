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

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ── 加密工具 ──────────────────────────────────────────────────────────────────

var masterKey []byte

func SetMasterKey(key string) {
	h := sha256.Sum256([]byte(key))
	masterKey = h[:]
}

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

type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Username     string    `gorm:"uniqueIndex" json:"username"`
	Password     string    `json:"-"`
	TokenVersion int       `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Settings 系统设置（SSH 字段已于迁移中移除）。
// JWTSecretEnc      — AES-GCM 加密的 JWT 签名密钥（base64）
// WebDAVPasswordEnc — AES-GCM 加密的 WebDAV 独立密码 bcrypt 哈希（base64）
type Settings struct {
	ID                uint   `gorm:"primarykey"`
	StorageDir        string `json:"storage_dir"`
	Lang              string `json:"lang"`
	UITheme           string `json:"ui_theme"`
	UIFont            string `json:"ui_font"`
	EditorFont        string `json:"editor_font"`
	WebDAVEnabled     bool   `json:"webdav_enabled"`
	WebDAVSubPath     string `json:"webdav_sub_path"`
	WebDAVUsername    string `json:"webdav_username"`
	WebDAVPasswordEnc string `json:"-"` // AES-GCM(bcrypt(password))
	JWTSecretEnc      string `json:"-"` // AES-GCM(jwt_secret)
	ShowHidden        bool   `json:"show_hidden"`
}

func (s *Settings) GetJWTSecret() (string, error) {
	if s.JWTSecretEnc == "" {
		return "", nil
	}
	return Decrypt(s.JWTSecretEnc)
}

func (s *Settings) SetJWTSecret(secret string) error {
	enc, err := Encrypt(secret)
	if err != nil {
		return err
	}
	s.JWTSecretEnc = enc
	return nil
}

func (s *Settings) GetWebDAVPasswordHash() (string, error) {
	if s.WebDAVPasswordEnc == "" {
		return "", nil
	}
	return Decrypt(s.WebDAVPasswordEnc)
}

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

type ShareLink struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Code      string     `gorm:"uniqueIndex" json:"code"`
	FilePath  string     `json:"file_path"`
	IsDir     bool       `json:"is_dir"`
	UserID    uint       `json:"user_id"`
	ExpiresAt *time.Time `json:"expires_at"`
	MaxViews  int        `json:"max_views"`
	ViewCount int        `json:"view_count"`
	CreatedAt time.Time  `json:"created_at"`
}

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

	// 数据库迁移：移除旧 SSH 列（SQLite 不支持 DROP COLUMN，用重建表方式）
	migrateDropSSHColumns(db)

	// 正常 AutoMigrate（添加新列、创建新表）
	db.AutoMigrate(&User{}, &Settings{}, &ShareLink{}, &FileVisibility{})

	var count int64
	db.Model(&Settings{}).Count(&count)
	if count == 0 {
		db.Create(&Settings{StorageDir: "./data/storage", Lang: "zh"})
	}

	return db, nil
}

// migrateDropSSHColumns 检测 settings 表是否含有旧 SSH 列，
// 若存在则通过重建表的方式将其移除。
// SQLite 从 3.35.0 才支持 DROP COLUMN，为保证兼容性使用重建方式。
func migrateDropSSHColumns(db *gorm.DB) {
	// 检查是否存在任意一个 SSH 列，有则执行迁移
	type colInfo struct {
		Name string
	}
	var cols []colInfo
	db.Raw("PRAGMA table_info(settings)").Scan(&cols)

	hasSSH := false
	sshCols := map[string]bool{
		"ssh_host": true, "ssh_port": true, "ssh_user": true,
		"ssh_auth_type": true, "ssh_password_enc": true, "ssh_private_key_enc": true,
	}
	for _, c := range cols {
		if sshCols[c.Name] {
			hasSSH = true
			break
		}
	}
	if !hasSSH {
		return // 已经是干净的，无需迁移
	}

	// 在事务内重建 settings 表，只保留当前 Settings 结构体对应的列
	db.Exec(`
		BEGIN;

		CREATE TABLE IF NOT EXISTS settings_new (
			id                  INTEGER PRIMARY KEY,
			storage_dir         TEXT,
			lang                TEXT,
			ui_theme            TEXT,
			ui_font             TEXT,
			editor_font         TEXT,
			web_dav_enabled     BOOLEAN,
			web_dav_sub_path    TEXT,
			web_dav_username    TEXT,
			web_dav_password_enc TEXT,
			jwt_secret_enc      TEXT,
			show_hidden         BOOLEAN
		);

		INSERT INTO settings_new (
			id, storage_dir, lang, ui_theme, ui_font, editor_font,
			web_dav_enabled, web_dav_sub_path, web_dav_username,
			web_dav_password_enc, jwt_secret_enc, show_hidden
		)
		SELECT
			id, storage_dir, lang, ui_theme, ui_font, editor_font,
			web_dav_enabled, web_dav_sub_path, web_dav_username,
			web_dav_password_enc, jwt_secret_enc, show_hidden
		FROM settings;

		DROP TABLE settings;
		ALTER TABLE settings_new RENAME TO settings;

		COMMIT;
	`)
}
