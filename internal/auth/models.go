package auth

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"uniqueIndex" json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Settings struct {
	ID              uint   `gorm:"primarykey"`
	StorageDir      string `json:"storage_dir"`
	Lang            string `json:"lang"`
	UITheme         string `json:"ui_theme"`
	UIFont          string `json:"ui_font"`
	EditorFont      string `json:"editor_font"`
	WebDAVEnabled   bool   `json:"webdav_enabled"`
	WebDAVSubPath   string `json:"webdav_sub_path"`
	WebDAVUsername  string `json:"webdav_username"`
	WebDAVPassword  string `json:"webdav_password"`
}

type ShareLink struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Code      string    `gorm:"uniqueIndex" json:"code"`
	FilePath  string    `json:"file_path"`
	IsDir     bool      `json:"is_dir"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type FileVisibility struct {
	ID       uint   `gorm:"primarykey"`
	FilePath string `gorm:"uniqueIndex" json:"file_path"`
	IsPublic bool   `json:"is_public"`
}

func InitDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 不打印 record not found
	})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{}, &Settings{}, &ShareLink{}, &FileVisibility{})

	// 用 Count 代替 First，避免触发 "record not found" 错误日志
	var count int64
	db.Model(&Settings{}).Count(&count)
	if count == 0 {
		db.Create(&Settings{StorageDir: "./data/storage", Lang: "zh"})
	}

	return db, nil
}
