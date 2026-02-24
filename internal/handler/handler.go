package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudone/cloudone/internal/auth"
	"github.com/cloudone/cloudone/internal/files"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte("cloudone-secret-change-me")

type Handler struct {
	db      *gorm.DB
	files   *files.Manager
	shares  *files.ShareManager
}

func New(db *gorm.DB, fm *files.Manager, sm *files.ShareManager) *Handler {
	return &Handler{db: db, files: fm, shares: sm}
}

func (h *Handler) AuthStatus(c *gin.Context) {
	var count int64
	h.db.Model(&auth.User{}).Count(&count)
	c.JSON(200, gin.H{"setup": count > 0})
}

func (h *Handler) Setup(c *gin.Context) {
	var count int64
	h.db.Model(&auth.User{}).Count(&count)
	if count > 0 {
		c.JSON(400, gin.H{"error": "already setup"})
		return
	}
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := auth.User{Username: req.Username, Email: req.Email, Password: string(hash)}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	token := h.genToken(user.ID)
	c.JSON(200, gin.H{"token": token, "user": user})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var user auth.User
	if h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	token := h.genToken(user.ID)
	c.JSON(200, gin.H{"token": token, "user": user})
}

func (h *Handler) genToken(userID uint) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	return t
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["sub"].(float64))
		var user auth.User
		if h.db.First(&user, userID).Error != nil {
			c.JSON(401, gin.H{"error": "user not found"})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func (h *Handler) GetUser(c *gin.Context) {
	u, _ := c.Get("user")
	c.JSON(200, u)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	u := c.MustGet("user").(auth.User)
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Username != "" {
		u.Username = req.Username
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		u.Password = string(hash)
	}
	h.db.Save(&u)
	c.JSON(200, u)
}

func (h *Handler) GetSettings(c *gin.Context) {
	var s auth.Settings
	h.db.First(&s)
	c.JSON(200, s)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var s auth.Settings
	h.db.First(&s)
	var req struct {
		StorageDir string `json:"storage_dir"`
		Lang       string `json:"lang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.StorageDir != "" {
		s.StorageDir = req.StorageDir
	}
	if req.Lang != "" {
		s.Lang = req.Lang
	}
	h.db.Save(&s)
	c.JSON(200, s)
}

func (h *Handler) ListFiles(c *gin.Context) {
	dir := c.Query("path")
	if dir == "" {
		dir = "/"
	}
	list, err := h.files.ListDir(dir)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"files": list, "path": dir})
}

func (h *Handler) UploadFile(c *gin.Context) {
	dir := c.Query("path")
	if dir == "" {
		dir = "/"
	}
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	fileHeaders := form.File["files"]
	for _, fh := range fileHeaders {
		f, err := fh.Open()
		if err != nil {
			continue
		}
		rel := filepath.Join(dir, fh.Filename)
		h.files.Write(rel, f)
		f.Close()
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if err := h.files.Delete(path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) Mkdir(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	c.ShouldBindJSON(&req)
	if err := h.files.MkDir(req.Path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) MoveFile(c *gin.Context) {
	var req struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	}
	c.ShouldBindJSON(&req)
	if err := h.files.Move(req.Src, req.Dst); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) CopyFile(c *gin.Context) {
	var req struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	}
	c.ShouldBindJSON(&req)
	if err := h.files.Copy(req.Src, req.Dst); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) DownloadFile(c *gin.Context) {
	path := c.Query("path")
	abs := h.files.AbsPath(path)
	name := filepath.Base(abs)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	c.File(abs)
}

func (h *Handler) CreateFile(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	c.ShouldBindJSON(&req)
	if err := h.files.Write(req.Path, strings.NewReader(req.Content)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) SetVisibility(c *gin.Context) {
	var req struct {
		Path     string `json:"path"`
		IsPublic bool   `json:"is_public"`
	}
	c.ShouldBindJSON(&req)
	if err := h.files.SetVisibility(req.Path, req.IsPublic); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) ListPublicFiles(c *gin.Context) {
	dir := c.Query("path")
	if dir == "" {
		dir = "/"
	}
	list, err := h.files.ListPublic(dir)
	if err != nil {
		c.JSON(200, gin.H{"files": []interface{}{}, "path": dir})
		return
	}
	c.JSON(200, gin.H{"files": list, "path": dir})
}

func (h *Handler) ServePublicFile(c *gin.Context) {
	path := c.Param("path")
	if !h.files.IsPublic(path) {
		// check parent dirs
		parts := strings.Split(strings.Trim(path, "/"), "/")
		found := false
		for i := len(parts) - 1; i >= 0; i-- {
			parentPath := "/" + strings.Join(parts[:i], "/")
			if h.files.IsPublic(parentPath) {
				found = true
				break
			}
		}
		if !found {
			c.JSON(403, gin.H{"error": "not public"})
			return
		}
	}
	abs := h.files.AbsPath(path)
	c.File(abs)
}

func (h *Handler) CreateShare(c *gin.Context) {
	u := c.MustGet("user").(auth.User)
	var req struct {
		Path  string `json:"path"`
		IsDir bool   `json:"is_dir"`
	}
	c.ShouldBindJSON(&req)
	link, err := h.shares.Create(u.ID, req.Path, req.IsDir)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, link)
}

func (h *Handler) ListShares(c *gin.Context) {
	u := c.MustGet("user").(auth.User)
	links, _ := h.shares.List(u.ID)
	c.JSON(200, links)
}

func (h *Handler) DeleteShare(c *gin.Context) {
	u := c.MustGet("user").(auth.User)
	id, _ := strconv.Atoi(c.Param("id"))
	h.shares.Delete(uint(id), u.ID)
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) AccessShare(c *gin.Context) {
	code := c.Param("code")
	link, err := h.shares.Get(code)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if link.IsDir {
		list, _ := h.files.ListDir(link.FilePath)
		c.JSON(200, gin.H{"files": list, "path": link.FilePath, "share": link})
	} else {
		info := gin.H{"path": link.FilePath, "share": link}
		c.JSON(200, info)
	}
}

func (h *Handler) DownloadShare(c *gin.Context) {
	code := c.Param("code")
	link, err := h.shares.Get(code)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if link.IsDir {
		c.JSON(400, gin.H{"error": "cannot download directory directly"})
		return
	}
	abs := h.files.AbsPath(link.FilePath)
	name := filepath.Base(abs)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	c.File(abs)
}
