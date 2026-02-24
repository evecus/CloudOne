package auth

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"uniqueIndex" json:"username"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Settings struct {
	ID         uint   `gorm:"primarykey"`
	StorageDir string `json:"storage_dir"`
	Lang       string `json:"lang"`
}

type ShareLink struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Code      string    `gorm:"uniqueIndex" json:"code"`
	FilePath  string    `json:"file_path"`
	IsDir     bool      `json:"is_dir"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type FileVisibility struct {
	ID       uint   `gorm:"primarykey"`
	FilePath string `gorm:"uniqueIndex" json:"file_path"`
	IsPublic bool   `json:"is_public"`
}

func InitDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&User{}, &Settings{}, &ShareLink{}, &FileVisibility{})

	// Ensure default settings
	var s Settings
	if db.First(&s).Error != nil {
		db.Create(&Settings{StorageDir: "./data/storage", Lang: "zh"})
	}

	return db, nil
}
