package handler

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudone/cloudone/internal/auth"
	"github.com/cloudone/cloudone/internal/config"
	"github.com/cloudone/cloudone/internal/files"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret []byte

// SetJWTSecret 由 main 在读取/生成 conf.ini 中的 jwt_secret 后调用。
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

type Handler struct {
	db       *gorm.DB
	files    *files.Manager
	shares   *files.ShareManager
	confPath string
}

func New(db *gorm.DB, fm *files.Manager, sm *files.ShareManager, confPath string) *Handler {
	return &Handler{db: db, files: fm, shares: sm, confPath: confPath}
}

func (h *Handler) syncConf() {
	cfg, err := config.Load(h.confPath)
	if err != nil {
		return
	}
	var user auth.User
	if h.db.First(&user).Error == nil {
		cfg.Username = user.Username
		cfg.Password = user.Password
	}
	var s auth.Settings
	h.db.First(&s)
	cfg.StorageDir = s.StorageDir
	cfg.Lang = s.Lang
	_ = config.Save(h.confPath, cfg)
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
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}
	user := auth.User{Username: req.Username, Password: string(hash)}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	token, err := h.genToken(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}
	h.syncConf()
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
	if h.db.Where("username = ?", req.Username).First(&user).Error != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := h.genToken(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(200, gin.H{"token": token, "user": user})
}

func (h *Handler) genToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
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
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Username != "" {
		u.Username = req.Username
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to hash password"})
			return
		}
		u.Password = string(hash)
	}
	h.db.Save(&u)
	h.syncConf()
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
		if err := os.MkdirAll(req.StorageDir, 0755); err != nil {
			c.JSON(500, gin.H{"error": "cannot create storage dir: " + err.Error()})
			return
		}
		s.StorageDir = req.StorageDir
		h.files.SetRoot(req.StorageDir)
	}
	if req.Lang != "" {
		s.Lang = req.Lang
	}
	h.db.Save(&s)
	h.syncConf()
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
	var failed []string
	for _, fh := range fileHeaders {
		safeName := filepath.Base(fh.Filename)
		if safeName == "." || safeName == ".." || safeName == "" {
			failed = append(failed, fh.Filename)
			continue
		}
		f, err := fh.Open()
		if err != nil {
			failed = append(failed, fh.Filename)
			continue
		}
		rel := filepath.Join(dir, safeName)
		if writeErr := h.files.Write(rel, f); writeErr != nil {
			failed = append(failed, fh.Filename)
		}
		f.Close()
	}
	if len(failed) > 0 {
		c.JSON(207, gin.H{"ok": true, "failed": failed})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(400, gin.H{"error": "path is required"})
		return
	}
	if err := h.files.Delete(path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) Mkdir(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.files.MkDir(req.Path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) MoveFile(c *gin.Context) {
	var req struct {
		Src string `json:"src" binding:"required"`
		Dst string `json:"dst" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.files.Move(req.Src, req.Dst); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) CopyFile(c *gin.Context) {
	var req struct {
		Src string `json:"src" binding:"required"`
		Dst string `json:"dst" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.files.Copy(req.Src, req.Dst); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) DownloadFile(c *gin.Context) {
	path := c.Query("path")
	abs, err := h.files.AbsPath(path)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	name := filepath.Base(abs)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	c.File(abs)
}

func (h *Handler) CreateFile(c *gin.Context) {
	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.files.Write(req.Path, strings.NewReader(req.Content)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) GetFileContent(c *gin.Context) {
	path := c.Query("path")
	content, err := h.files.ReadContent(path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"content": content})
}

func (h *Handler) UpdateFileContent(c *gin.Context) {
	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.files.Write(req.Path, strings.NewReader(req.Content)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) GetPermission(c *gin.Context) {
	path := c.Query("path")
	mode, err := h.files.GetPermission(path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"mode": int(mode)})
}

func (h *Handler) SetPermission(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
		Mode int    `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.files.Chmod(req.Path, os.FileMode(req.Mode)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *Handler) SetVisibility(c *gin.Context) {
	var req struct {
		Path     string `json:"path" binding:"required"`
		IsPublic bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.files.SetVisibility(req.Path, req.IsPublic); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// ListPublicFiles 返回公开文件列表（需登录）
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

// ListPublicFilesOpen 无需认证的公开文件列表（用于 /pub 前台页面）
func (h *Handler) ListPublicFilesOpen(c *gin.Context) {
	h.ListPublicFiles(c)
}

// ServePublicFile 服务公开文件（/raw/* 路径）。
// 规则：文件本身或任意祖先目录被设为公开，且没有被中间层明确设为私有。
func (h *Handler) ServePublicFile(c *gin.Context) {
	path := c.Param("path")

	abs, err := h.files.AbsPath(path)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}

	// 使用带私有覆盖检查的IsPublic
	if !h.files.IsPublic(path) {
		c.JSON(403, gin.H{"error": "not public"})
		return
	}

	info, err := os.Stat(abs)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	if info.IsDir() {
		list, err := h.files.ListPublic(path)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"files": list, "path": path})
		return
	}

	// 文件：返回原始内容（类似 GitHub raw）
	serveRaw(c, abs)
}

// DownloadPublicFile 无需认证，通过 /pub/dl?path= 下载公开文件。
func (h *Handler) DownloadPublicFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(400, gin.H{"error": "path is required"})
		return
	}
	if !h.files.IsPublic(path) {
		c.JSON(403, gin.H{"error": "not public"})
		return
	}
	abs, err := h.files.AbsPath(path)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}
	name := filepath.Base(abs)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	c.File(abs)
}

// ServeRaw 通过分享码查看文件原始内容（类似 GitHub raw）。
func (h *Handler) ServeRaw(c *gin.Context) {
	code := c.Param("code")
	link, err := h.shares.Get(code)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if link.IsDir {
		c.JSON(400, gin.H{"error": "cannot view directory as raw"})
		return
	}
	abs, err := h.files.AbsPath(link.FilePath)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}
	serveRaw(c, abs)
}

// DownloadShare 下载分享的文件（单文件分享 或 目录分享内的文件，通过 ?subpath= 指定）
func (h *Handler) DownloadShare(c *gin.Context) {
	code := c.Param("code")
	link, err := h.shares.Get(code)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var filePath string
	if link.IsDir {
		subpath := c.Query("subpath")
		if subpath == "" {
			c.JSON(400, gin.H{"error": "subpath required for directory share"})
			return
		}
		filePath = link.FilePath + "/" + subpath
	} else {
		filePath = link.FilePath
	}
	abs, err := h.files.AbsPath(filePath)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}
	c.FileAttachment(abs, filepath.Base(abs))
}

// browserPreviewable 判断浏览器是否能原生预览该 MIME 类型。
// 可预览 → inline（浏览器直接展示）
// 不可预览 → attachment（强制下载）
func browserPreviewable(mimeType string) bool {
	// 文本类：浏览器直接渲染为文本
	if strings.HasPrefix(mimeType, "text/") {
		return true
	}
	// 图片类：浏览器原生渲染
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	// 音频/视频：浏览器内置播放器
	if strings.HasPrefix(mimeType, "audio/") || strings.HasPrefix(mimeType, "video/") {
		return true
	}
	// 常见可内联类型
	switch mimeType {
	case "application/pdf",
		"application/json",
		"application/javascript",
		"application/x-javascript",
		"application/xml",
		"application/xhtml+xml",
		"application/atom+xml",
		"application/rss+xml",
		"application/svg+xml":
		return true
	}
	// xml、javascript 变体
	if strings.Contains(mimeType, "xml") || strings.Contains(mimeType, "javascript") {
		return true
	}
	return false
}

// extMimeOverride 对系统 mime 库可能缺失或不准确的扩展名补充映射。
var extMimeOverride = map[string]string{
	".md":         "text/plain; charset=utf-8",
	".markdown":   "text/plain; charset=utf-8",
	".yaml":       "text/plain; charset=utf-8",
	".yml":        "text/plain; charset=utf-8",
	".toml":       "text/plain; charset=utf-8",
	".ini":        "text/plain; charset=utf-8",
	".conf":       "text/plain; charset=utf-8",
	".env":        "text/plain; charset=utf-8",
	".log":        "text/plain; charset=utf-8",
	".sh":         "text/plain; charset=utf-8",
	".bash":       "text/plain; charset=utf-8",
	".zsh":        "text/plain; charset=utf-8",
	".fish":       "text/plain; charset=utf-8",
	".dockerfile": "text/plain; charset=utf-8",
	".makefile":   "text/plain; charset=utf-8",
	".go":         "text/plain; charset=utf-8",
	".py":         "text/plain; charset=utf-8",
	".rs":         "text/plain; charset=utf-8",
	".rb":         "text/plain; charset=utf-8",
	".java":       "text/plain; charset=utf-8",
	".c":          "text/plain; charset=utf-8",
	".cpp":        "text/plain; charset=utf-8",
	".h":          "text/plain; charset=utf-8",
	".ts":         "text/plain; charset=utf-8",
	".tsx":        "text/plain; charset=utf-8",
	".jsx":        "text/plain; charset=utf-8",
	".vue":        "text/plain; charset=utf-8",
	".swift":      "text/plain; charset=utf-8",
	".kt":         "text/plain; charset=utf-8",
	".lua":        "text/plain; charset=utf-8",
	".r":          "text/plain; charset=utf-8",
	".sql":        "text/plain; charset=utf-8",
	".graphql":    "text/plain; charset=utf-8",
	".proto":      "text/plain; charset=utf-8",
	".avif":       "image/avif",
	".webp":       "image/webp",
}

// serveRaw 类 GitHub raw 行为：
//   - 浏览器可预览（文本、图片、音视频、PDF 等）→ Content-Disposition: inline，浏览器直接展示
//   - 不可预览（zip、exe、docx 等）→ Content-Disposition: attachment，强制下载
func serveRaw(c *gin.Context, abs string) {
	ext := strings.ToLower(filepath.Ext(abs))

	// 先查补充表
	mimeType, ok := extMimeOverride[ext]
	if !ok {
		mimeType = mime.TypeByExtension(ext)
	}
	if mimeType == "" {
		// 无法识别扩展名，尝试当文本处理（GitHub 也是如此）
		mimeType = "text/plain; charset=utf-8"
	}

	name := filepath.Base(abs)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "public, max-age=3600")

	// 取 mimeType 的 base 部分（去掉 ; charset=utf-8 等参数）用于判断
	baseMime := strings.SplitN(mimeType, ";", 2)[0]
	baseMime = strings.TrimSpace(baseMime)

	if browserPreviewable(baseMime) {
		// 内联：浏览器直接展示
		c.Header("Content-Type", mimeType)
		c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, name))
		c.File(abs)
	} else {
		// 强制下载
		c.Header("Content-Type", mimeType)
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
		c.File(abs)
	}
}

func (h *Handler) CreateShare(c *gin.Context) {
	u := c.MustGet("user").(auth.User)
	var req struct {
		Path  string `json:"path" binding:"required"`
		IsDir bool   `json:"is_dir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
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
		// subpath 参数支持子目录导航
		subpath := c.Query("subpath")
		var listPath string
		if subpath != "" {
			listPath = link.FilePath + "/" + subpath
		} else {
			listPath = link.FilePath
		}
		list, _ := h.files.ListDir(listPath)
		c.JSON(200, gin.H{"files": list, "is_dir": true, "code": link.Code, "file_path": link.FilePath})
	} else {
		c.JSON(200, gin.H{"is_dir": false, "code": link.Code, "file_path": link.FilePath})
	}
}

// BatchDelete 批量删除文件/目录
func (h *Handler) BatchDelete(c *gin.Context) {
	var req struct {
		Paths []string `json:"paths" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var failed []string
	for _, p := range req.Paths {
		if p == "" || p == "/" {
			continue
		}
		if err := h.files.Delete(p); err != nil {
			failed = append(failed, p)
		}
	}
	if len(failed) > 0 {
		c.JSON(207, gin.H{"ok": false, "failed": failed})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// BatchDownload 打包下载多个文件/目录为 ZIP
func (h *Handler) BatchDownload(c *gin.Context) {
	var req struct {
		Paths []string `json:"paths" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="download.zip"`)
	c.Header("Content-Type", "application/zip")

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	for _, rel := range req.Paths {
		abs, err := h.files.AbsPath(rel)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		baseName := filepath.Base(abs)
		if info.IsDir() {
			// 递归打包目录
			_ = filepath.Walk(abs, func(path string, fi os.FileInfo, werr error) error {
				if werr != nil || fi.IsDir() {
					return nil
				}
				relInZip := baseName + "/" + strings.TrimPrefix(path, abs+string(os.PathSeparator))
				relInZip = filepath.ToSlash(relInZip)
				w, err := zw.Create(relInZip)
				if err != nil {
					return nil
				}
				f, err := os.Open(path)
				if err != nil {
					return nil
				}
				_, _ = io.Copy(w, f)
				f.Close()
				return nil
			})
		} else {
			w, err := zw.Create(baseName)
			if err != nil {
				continue
			}
			f, err := os.Open(abs)
			if err != nil {
				continue
			}
			_, _ = io.Copy(w, f)
			f.Close()
		}
	}
}

// BatchMove 批量移动文件/目录到目标目录
func (h *Handler) BatchMove(c *gin.Context) {
	var req struct {
		Paths  []string `json:"paths" binding:"required"`
		Target string   `json:"target" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	targetNorm := filepath.Clean("/" + strings.TrimLeft(req.Target, "/"))

	var failed []string
	var blocked []string
	for _, src := range req.Paths {
		srcNorm := filepath.Clean("/" + strings.TrimLeft(src, "/"))
		// 防止将目录移动到自身或其子目录
		if srcNorm == targetNorm || strings.HasPrefix(targetNorm+"/", srcNorm+"/") {
			blocked = append(blocked, src)
			continue
		}
		name := filepath.Base(srcNorm)
		dst := targetNorm + "/" + name
		if err := h.files.Move(src, dst); err != nil {
			failed = append(failed, src)
		}
	}
	if len(blocked) > 0 {
		c.JSON(400, gin.H{"error": "cannot move a folder into itself or its subdirectory", "blocked": blocked})
		return
	}
	if len(failed) > 0 {
		c.JSON(207, gin.H{"ok": false, "failed": failed})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// UploadFolder 上传整个文件夹（保留相对路径结构）
// 前端通过 webkitRelativePath 传递相对路径，表单字段名为 files，路径字段名为 paths
func (h *Handler) UploadFolder(c *gin.Context) {
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
	paths := form.Value["paths"]

	var failed []string
	for i, fh := range fileHeaders {
		// 相对路径：如 "myFolder/sub/file.txt"
		relPath := fh.Filename
		if i < len(paths) && paths[i] != "" {
			relPath = paths[i]
		}
		// 清理路径，防止穿越
		relPath = filepath.ToSlash(relPath)
		relPath = strings.TrimPrefix(relPath, "/")
		// 目标路径 = 当前目录 + 相对路径
		targetRel := dir + "/" + relPath

		f, err := fh.Open()
		if err != nil {
			failed = append(failed, relPath)
			continue
		}
		if writeErr := h.files.Write(targetRel, f); writeErr != nil {
			failed = append(failed, relPath)
		}
		f.Close()
	}
	if len(failed) > 0 {
		c.JSON(207, gin.H{"ok": true, "failed": failed})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// ListDirTree 返回目录树（用于移动目标选择器）
func (h *Handler) ListDirTree(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "/"
	}
	items, err := h.files.ListDir(path)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// 只返回目录
	var dirs []map[string]interface{}
	for _, item := range items {
		if item.IsDir {
			dirs = append(dirs, map[string]interface{}{
				"name":  item.Name,
				"path":  item.Path,
				"is_dir": true,
			})
		}
	}
	c.JSON(200, gin.H{"dirs": dirs, "path": path})
}

// BatchCopy 批量复制文件/目录到目标目录
func (h *Handler) BatchCopy(c *gin.Context) {
	var req struct {
		Paths  []string `json:"paths" binding:"required"`
		Target string   `json:"target" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var failed []string
	for _, src := range req.Paths {
		name := filepath.Base(src)
		dst := req.Target + "/" + name
		if err := h.files.Copy(src, dst); err != nil {
			failed = append(failed, src)
		}
	}
	if len(failed) > 0 {
		c.JSON(207, gin.H{"ok": false, "failed": failed})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// CompressFiles 将选中文件/目录压缩为指定格式
// 支持 format: zip, tar.gz, tar
func (h *Handler) CompressFiles(c *gin.Context) {
	var req struct {
		Paths  []string `json:"paths" binding:"required"`
		Format string   `json:"format"` // zip | tar.gz | tar
		Output string   `json:"output"` // 输出文件名（含扩展名），相对于当前目录
		Dir    string   `json:"dir"`    // 当前目录
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Format == "" {
		req.Format = "zip"
	}
	if req.Dir == "" {
		req.Dir = "/"
	}
	if req.Output == "" {
		req.Output = "archive"
	}
	// 保证输出有正确扩展名
	switch req.Format {
	case "tar.gz":
		if !strings.HasSuffix(req.Output, ".tar.gz") {
			req.Output += ".tar.gz"
		}
	case "tar":
		if !strings.HasSuffix(req.Output, ".tar") {
			req.Output += ".tar"
		}
	default:
		req.Format = "zip"
		if !strings.HasSuffix(req.Output, ".zip") {
			req.Output += ".zip"
		}
	}

	outRel := req.Dir + "/" + req.Output
	outAbs, err := h.files.AbsPath(outRel)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid output path: " + err.Error()})
		return
	}
	if err := os.MkdirAll(filepath.Dir(outAbs), 0755); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 收集所有绝对路径
	type entry struct{ abs, rel string }
	var entries []entry
	for _, rel := range req.Paths {
		abs, err := h.files.AbsPath(rel)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		baseName := filepath.Base(abs)
		if info.IsDir() {
			filepath.Walk(abs, func(path string, fi os.FileInfo, werr error) error {
				if werr != nil || fi.IsDir() {
					return nil
				}
				relPath := baseName + "/" + strings.TrimPrefix(filepath.ToSlash(path), filepath.ToSlash(abs)+"/")
				entries = append(entries, entry{abs: path, rel: relPath})
				return nil
			})
		} else {
			entries = append(entries, entry{abs: abs, rel: baseName})
		}
	}

	switch req.Format {
	case "zip":
		f, err := os.Create(outAbs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		zw := zip.NewWriter(f)
		for _, e := range entries {
			w, err := zw.Create(e.rel)
			if err != nil {
				continue
			}
			src, err := os.Open(e.abs)
			if err != nil {
				continue
			}
			io.Copy(w, src)
			src.Close()
		}
		zw.Close()
		f.Close()
	case "tar", "tar.gz":
		f, err := os.Create(outAbs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var tw *tar.Writer
		if req.Format == "tar.gz" {
			gw := gzip.NewWriter(f)
			tw = tar.NewWriter(gw)
			defer gw.Close()
		} else {
			tw = tar.NewWriter(f)
		}
		for _, e := range entries {
			info, err := os.Stat(e.abs)
			if err != nil {
				continue
			}
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				continue
			}
			hdr.Name = e.rel
			tw.WriteHeader(hdr)
			src, err := os.Open(e.abs)
			if err != nil {
				continue
			}
			io.Copy(tw, src)
			src.Close()
		}
		tw.Close()
		f.Close()
	}

	c.JSON(200, gin.H{"ok": true, "output": req.Output})
}

// FetchURL 通过 wget/curl 下载远程文件到指定目录（文件浏览器当前目录）
func (h *Handler) FetchURL(c *gin.Context) {
	var req struct {
		URL      string `json:"url" binding:"required"`
		Filename string `json:"filename"` // 留空则由工具自动命名
		Dir      string `json:"dir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Dir == "" {
		req.Dir = "/"
	}

	absDir, err := h.files.AbsPath(req.Dir)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid dir: " + err.Error()})
		return
	}

	// 转为绝对路径，避免 wget/curl 受进程 cwd 影响
	absDir, err = filepath.Abs(absDir)
	if err != nil {
		c.JSON(500, gin.H{"error": "cannot resolve absolute path: " + err.Error()})
		return
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 先检测可用工具，再构建命令
	hasWget := func() bool { _, e := exec.LookPath("wget"); return e == nil }()
	hasCurl := func() bool { _, e := exec.LookPath("curl"); return e == nil }()

	if !hasWget && !hasCurl {
		c.JSON(500, gin.H{"error": "neither wget nor curl is available on this system"})
		return
	}

	var cmd *exec.Cmd
	outPath := filepath.Join(absDir, req.Filename)

	if req.Filename != "" {
		// 指定输出文件名
		if hasWget {
			cmd = exec.Command("wget", "-q", "-O", outPath, req.URL)
		} else {
			cmd = exec.Command("curl", "-L", "-o", outPath, req.URL)
		}
	} else {
		// 自动从 URL 提取文件名，下载到 absDir
		if hasWget {
			cmd = exec.Command("wget", "-q", "-P", absDir, req.URL)
		} else {
			// curl 用 --output-dir + -O 自动命名（curl >= 7.73）
			cmd = exec.Command("curl", "-L", "--output-dir", absDir, "-O", req.URL)
		}
	}
	cmd.Dir = absDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(500, gin.H{"error": "fetch failed: " + err.Error() + " | " + string(out)})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// ExtractFile 解压压缩包到指定目录（当前目录）
// SearchFiles 在指定目录（含子目录）中按文件名精确搜索
func (h *Handler) SearchFiles(c *gin.Context) {
	name := c.Query("name")
	dir := c.Query("dir")
	if name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}
	if dir == "" {
		dir = "/"
	}

	absDir, err := h.files.AbsPath(dir)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid dir: " + err.Error()})
		return
	}

	type Result struct {
		Name  string `json:"name"`
		Path  string `json:"path"` // 相对于 storageRoot 的路径（含前导 /）
		IsDir bool   `json:"is_dir"`
	}

	var results []Result
	root := h.files.Root()

	err = filepath.Walk(absDir, func(p string, info os.FileInfo, werr error) error {
		if werr != nil {
			return nil // 跳过无权限目录
		}
		if p == absDir {
			return nil // 跳过起始目录本身
		}
		if info.Name() == name {
			// 转为相对于 storageRoot 的路径
			rel, _ := filepath.Rel(root, p)
			rel = "/" + filepath.ToSlash(rel)
			results = append(results, Result{
				Name:  info.Name(),
				Path:  rel,
				IsDir: info.IsDir(),
			})
		}
		return nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if results == nil {
		results = []Result{}
	}
	c.JSON(200, gin.H{"results": results})
}

// DecompressFile 解压文件到当前目录
// 支持 .zip .tar.gz .tar .gz
func (h *Handler) DecompressFile(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"` // 压缩文件的路径
		Dir  string `json:"dir"`                     // 解压目标目录（默认与文件同目录）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	absPath, err := h.files.AbsPath(req.Path)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path: " + err.Error()})
		return
	}

	// 解压目录：默认为文件所在目录
	destDir := req.Dir
	if destDir == "" {
		parentRel := filepath.Dir(req.Path)
		if parentRel == "." {
			parentRel = "/"
		}
		destDir = parentRel
	}
	absDestDir, err := h.files.AbsPath(destDir)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid dest dir: " + err.Error()})
		return
	}
	absDestDir, err = filepath.Abs(absDestDir)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if err := os.MkdirAll(absDestDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	name := strings.ToLower(filepath.Base(absPath))
	switch {
	case strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz"):
		err = extractTarGz(absPath, absDestDir)
	case strings.HasSuffix(name, ".tar"):
		err = extractTar(absPath, absDestDir)
	case strings.HasSuffix(name, ".gz"):
		// 单文件 gzip
		err = extractGz(absPath, absDestDir)
	case strings.HasSuffix(name, ".zip"):
		err = extractZip(absPath, absDestDir)
	default:
		c.JSON(400, gin.H{"error": "unsupported archive format"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "decompress failed: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue // 跳过路径穿越
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(fpath)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractTar(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return untar(f, dest)
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()
	return untar(gr, dest)
}

func extractGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()
	outName := strings.TrimSuffix(filepath.Base(src), ".gz")
	outPath := filepath.Join(dest, outName)
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, gr)
	return err
}

func untar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fpath := filepath.Join(dest, filepath.FromSlash(hdr.Name))
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(fpath, 0755)
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}
			out, err := os.Create(fpath)
			if err != nil {
				return err
			}
			_, err = io.Copy(out, tr)
			out.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
