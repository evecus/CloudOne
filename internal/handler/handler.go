package handler

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net"
	"net/url"
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

func SetJWTSecret(secret string) { jwtSecret = []byte(secret) }

type Handler struct {
	db           *gorm.DB
	files        *files.Manager
	shares       *files.ShareManager
	confPath     string
	loginLimiter *loginRateLimiter
}

func New(db *gorm.DB, fm *files.Manager, sm *files.ShareManager, confPath string) *Handler {
	return &Handler{
		db:           db,
		files:        fm,
		shares:       sm,
		confPath:     confPath,
		loginLimiter: newLoginRateLimiter(),
	}
}

// syncConf 只保存 host/port 到 conf.ini，其余数据全在 DB
func (h *Handler) syncConf() {
	cfg, err := config.Load(h.confPath)
	if err != nil {
		return
	}
	_ = config.Save(h.confPath, cfg)
}

// ── 认证 ──────────────────────────────────────────────────────────────────────

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
	token, err := h.genToken(user)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(200, gin.H{"token": token, "user": safeUser(user)})
}

func (h *Handler) Login(c *gin.Context) {
	// 频率限制：同一 IP 每分钟最多 5 次
	if !h.loginLimiter.Allow(c.ClientIP()) {
		c.JSON(429, gin.H{"error": "too many login attempts, please try again later"})
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
	var user auth.User
	if h.db.Where("username = ?", req.Username).First(&user).Error != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := h.genToken(user)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(200, gin.H{"token": token, "user": safeUser(user)})
}

// genToken 生成 JWT，有效期 7 天，携带 token_version 用于主动吊销
func (h *Handler) genToken(user auth.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":     user.ID,
		"version": user.TokenVersion,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

// safeUser 返回不含密码等敏感字段的用户信息
func safeUser(u auth.User) gin.H {
	return gin.H{
		"id":         u.ID,
		"username":   u.Username,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
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
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["sub"].(float64))
		tokenVersion := int(claims["version"].(float64))

		var user auth.User
		if h.db.First(&user, userID).Error != nil {
			c.JSON(401, gin.H{"error": "user not found"})
			c.Abort()
			return
		}
		// 校验 token 版本，密码修改后旧 token 立即失效
		if user.TokenVersion != tokenVersion {
			c.JSON(401, gin.H{"error": "token has been revoked, please login again"})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func (h *Handler) GetUser(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(auth.User)
	c.JSON(200, safeUser(user))
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
	passwordChanged := false
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to hash password"})
			return
		}
		u.Password = string(hash)
		u.TokenVersion++ // 密码修改：旧 token 立即全部失效
		passwordChanged = true
	}
	h.db.Save(&u)

	resp := gin.H{"user": safeUser(u)}
	// 密码改了，下发新 token（当前 session 继续有效）
	if passwordChanged {
		newToken, err := h.genToken(u)
		if err == nil {
			resp["token"] = newToken
		}
	}
	c.JSON(200, resp)
}

// ── Settings ──────────────────────────────────────────────────────────────────

func (h *Handler) GetSettings(c *gin.Context) {
	var s auth.Settings
	h.db.First(&s)
	c.JSON(200, s) // json:"-" 字段自动过滤，敏感字段不暴露
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var s auth.Settings
	h.db.First(&s)
	var req struct {
		StorageDir string `json:"storage_dir"`
		Lang       string `json:"lang"`
		UITheme    string `json:"ui_theme"`
		UIFont     string `json:"ui_font"`
		EditorFont string `json:"editor_font"`
		ShowHidden *bool  `json:"show_hidden"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.StorageDir != "" {
		// 安全检查：禁止指向系统关键目录
		forbidden := []string{"/etc", "/bin", "/sbin", "/usr/bin", "/usr/sbin",
			"/boot", "/sys", "/proc", "/dev", "/run", "/var/run"}
		clean := filepath.Clean(req.StorageDir)
		for _, f := range forbidden {
			if clean == f || strings.HasPrefix(clean, f+"/") {
				c.JSON(400, gin.H{"error": "storage_dir points to a system directory"})
				return
			}
		}
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
	if req.UITheme != "" {
		s.UITheme = req.UITheme
	}
	if req.UIFont != "" {
		s.UIFont = req.UIFont
	}
	if req.EditorFont != "" {
		s.EditorFont = req.EditorFont
	}
	if req.ShowHidden != nil {
		s.ShowHidden = *req.ShowHidden
	}
	h.db.Save(&s)
	c.JSON(200, s)
}

// ── WebDAV Settings ───────────────────────────────────────────────────────────

func (h *Handler) GetWebDAVSettings(c *gin.Context) {
	var s auth.Settings
	h.db.First(&s)
	// 密码字段只返回"是否已设置"，不返回任何密码内容
	c.JSON(200, gin.H{
		"webdav_enabled":      s.WebDAVEnabled,
		"webdav_sub_path":     s.WebDAVSubPath,
		"webdav_username":     s.WebDAVUsername,
		"webdav_has_password": s.WebDAVPasswordEnc != "",
	})
}

func (h *Handler) UpdateWebDAVSettings(c *gin.Context) {
	var s auth.Settings
	h.db.First(&s)
	var req struct {
		Enabled  bool   `json:"webdav_enabled"`
		SubPath  string `json:"webdav_sub_path"`
		Username string `json:"webdav_username"`
		Password string `json:"webdav_password"` // 空字符串=不修改密码
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	s.WebDAVEnabled = req.Enabled
	s.WebDAVSubPath = req.SubPath
	s.WebDAVUsername = req.Username
	// 仅当传入新密码时才更新
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to hash password"})
			return
		}
		if err := s.SetWebDAVPasswordHash(string(hash)); err != nil {
			c.JSON(500, gin.H{"error": "failed to encrypt password"})
			return
		}
	}
	h.db.Save(&s)
	c.JSON(200, gin.H{
		"webdav_enabled":      s.WebDAVEnabled,
		"webdav_sub_path":     s.WebDAVSubPath,
		"webdav_username":     s.WebDAVUsername,
		"webdav_has_password": s.WebDAVPasswordEnc != "",
	})
}

// ── 文件操作 ──────────────────────────────────────────────────────────────────

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
	// MultipartForm 会把文件数据缓存在内存或临时文件中，
	// 函数退出时必须手动释放，否则内存不会回收。
	defer form.RemoveAll()
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

func (h *Handler) ListPublicFiles(c *gin.Context) {
	dir := c.Query("path")
	if dir == "" {
		dir = "/"
	}
	var (
		list []files.FileInfo
		err  error
	)
	if dir == "/" {
		list, err = h.files.GetAllPublicFlat()
	} else {
		list, err = h.files.ListPublic(dir)
	}
	if err != nil || list == nil {
		c.JSON(200, gin.H{"files": []interface{}{}, "path": dir})
		return
	}
	c.JSON(200, gin.H{"files": list, "path": dir})
}

func (h *Handler) ListPublicFilesOpen(c *gin.Context) {
	h.ListPublicFiles(c)
}

func (h *Handler) ServePublicFile(c *gin.Context) {
	path := c.Param("path")
	abs, err := h.files.AbsPath(path)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}
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
	serveRaw(c, abs)
}

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
		// 净化 subpath，防路径穿越
		subpath = filepath.ToSlash(filepath.Clean("/" + subpath))
		subpath = strings.TrimPrefix(subpath, "/")
		filePath = link.FilePath + "/" + subpath

		// 校验最终路径仍在分享目录内
		shareAbs, err := h.files.AbsPath(link.FilePath)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid share path"})
			return
		}
		targetAbs, err := h.files.AbsPath(filePath)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid path"})
			return
		}
		if targetAbs != shareAbs && !strings.HasPrefix(targetAbs, shareAbs+string(os.PathSeparator)) {
			c.JSON(403, gin.H{"error": "access denied"})
			return
		}
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

// ── 批量操作 ──────────────────────────────────────────────────────────────────

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

func (h *Handler) BatchDownload(c *gin.Context) {
	var req struct {
		Paths    []string `json:"paths" binding:"required"`
		Filename string   `json:"filename"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 确定下载文件名
	zipName := req.Filename
	if zipName == "" {
		zipName = "download"
	}
	if !strings.HasSuffix(strings.ToLower(zipName), ".zip") {
		zipName += ".zip"
	}

	// 收集所有有效的绝对路径
	storageRoot := h.files.Root()
	type entry struct {
		abs     string
		relRoot string // 相对于 storageRoot 的路径，用作 zip 内路径前缀
	}
	var entries []entry
	for _, rel := range req.Paths {
		abs, err := h.files.SafeAbsPath(rel)
		if err != nil {
			continue
		}
		relToRoot, err := filepath.Rel(storageRoot, abs)
		if err != nil {
			continue
		}
		entries = append(entries, entry{abs: abs, relRoot: relToRoot})
	}
	if len(entries) == 0 {
		c.JSON(400, gin.H{"error": "no valid paths"})
		return
	}

	// 使用 Go 标准库直接流式写入响应，无需外部 zip 命令，无需临时文件
	c.Header("Content-Disposition", `attachment; filename="`+zipName+`"`)
	c.Header("Content-Type", "application/zip")
	c.Status(200)

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	// addToZip 递归把文件/目录写入 zip
	var addToZip func(absPath, zipPrefix string) error
	addToZip = func(absPath, zipPrefix string) error {
		info, err := os.Lstat(absPath)
		if err != nil {
			return nil // 跳过无法访问的文件
		}
		if info.IsDir() {
			des, err := os.ReadDir(absPath)
			if err != nil {
				return nil
			}
			for _, de := range des {
				child := filepath.Join(absPath, de.Name())
				childPrefix := zipPrefix + "/" + de.Name()
				if err := addToZip(child, childPrefix); err != nil {
					return err
				}
			}
			return nil
		}
		// 普通文件
		fh, err := zip.FileInfoHeader(info)
		if err != nil {
			return nil
		}
		fh.Name = zipPrefix
		fh.Method = zip.Deflate
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}
		f, err := os.Open(absPath)
		if err != nil {
			return nil
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	}

	for _, e := range entries {
		_ = addToZip(e.abs, e.relRoot)
	}
}

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
	var failed, blocked []string
	for _, src := range req.Paths {
		srcNorm := filepath.Clean("/" + strings.TrimLeft(src, "/"))
		if srcNorm == targetNorm || strings.HasPrefix(targetNorm+"/", srcNorm+"/") {
			blocked = append(blocked, src)
			continue
		}
		dst := targetNorm + "/" + filepath.Base(srcNorm)
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

func (h *Handler) BatchCopy(c *gin.Context) {
	var req struct {
		Paths  []string `json:"paths" binding:"required"`
		Target string   `json:"target" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	targetNorm := filepath.Clean("/" + strings.TrimLeft(req.Target, "/"))
	var failed, blocked []string
	for _, src := range req.Paths {
		srcNorm := filepath.Clean("/" + strings.TrimLeft(src, "/"))
		// 防止复制到自身或子目录（会导致无限递归）
		if srcNorm == targetNorm || strings.HasPrefix(targetNorm+"/", srcNorm+"/") {
			blocked = append(blocked, src)
			continue
		}
		dst := targetNorm + "/" + filepath.Base(srcNorm)
		if err := h.files.Copy(src, dst); err != nil {
			failed = append(failed, src)
		}
	}
	if len(blocked) > 0 {
		c.JSON(400, gin.H{"error": "cannot copy a folder into itself or its subdirectory", "blocked": blocked})
		return
	}
	if len(failed) > 0 {
		c.JSON(207, gin.H{"ok": false, "failed": failed})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

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
	defer form.RemoveAll()
	fileHeaders := form.File["files"]
	paths := form.Value["paths"]
	var failed []string
	for i, fh := range fileHeaders {
		relPath := fh.Filename
		if i < len(paths) && paths[i] != "" {
			relPath = paths[i]
		}
		// 净化路径，防止路径穿越
		relPath = filepath.ToSlash(filepath.Clean("/" + strings.TrimPrefix(relPath, "/")))
		relPath = strings.TrimPrefix(relPath, "/")
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
	var dirs []map[string]interface{}
	for _, item := range items {
		if item.IsDir {
			dirs = append(dirs, map[string]interface{}{
				"name":   item.Name,
				"path":   item.Path,
				"is_dir": true,
			})
		}
	}
	c.JSON(200, gin.H{"dirs": dirs, "path": path})
}

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
	results, err := h.files.SearchFiles(dir, name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if results == nil {
		results = []files.SearchResult{}
	}
	c.JSON(200, gin.H{"results": results})
}

// ── 分享 ──────────────────────────────────────────────────────────────────────

func (h *Handler) CreateShare(c *gin.Context) {
	u := c.MustGet("user").(auth.User)
	var req struct {
		Path     string `json:"path" binding:"required"`
		IsDir    bool   `json:"is_dir"`
		ExpireIn int    `json:"expire_in"` // 秒数，0 = 永不过期
		MaxViews int    `json:"max_views"` // 0 = 不限次数
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var expiresAt *time.Time
	if req.ExpireIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpireIn) * time.Second)
		expiresAt = &t
	}
	link, err := h.shares.Create(u.ID, req.Path, req.IsDir, expiresAt, req.MaxViews)
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
	// 目录浏览用 Peek，不递增 view_count；实际下载才用 Get 计数
	link, err := h.shares.Peek(code)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if link.IsDir {
		subpath := c.Query("subpath")
		listPath := link.FilePath
		if subpath != "" {
			// 净化 subpath
			subpath = filepath.ToSlash(filepath.Clean("/" + subpath))
			subpath = strings.TrimPrefix(subpath, "/")
			listPath = link.FilePath + "/" + subpath
		}
		list, _ := h.files.ListDir(listPath)
		c.JSON(200, gin.H{"files": list, "is_dir": true, "code": link.Code, "file_path": link.FilePath})
	} else {
		c.JSON(200, gin.H{"is_dir": false, "code": link.Code, "file_path": link.FilePath})
	}
}

// ── 压缩/解压 ─────────────────────────────────────────────────────────────────

func (h *Handler) CompressFiles(c *gin.Context) {
	var req struct {
		Paths  []string `json:"paths" binding:"required"`
		Format string   `json:"format"`
		Output string   `json:"output"`
		Dir    string   `json:"dir"`
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

	// 把每个逻辑路径转为相对于 storage root 的相对路径，
	// 命令从 root 目录执行，压缩包内只保留 basename 起的相对路径，不含绝对路径。
	storageRoot := h.files.Root()
	var relArgs []string
	for _, rel := range req.Paths {
		abs, err := h.files.AbsPath(rel)
		if err != nil {
			continue
		}
		// 转为相对于 storageRoot 的路径，作为命令参数
		relToRoot, err := filepath.Rel(storageRoot, abs)
		if err != nil {
			continue
		}
		relArgs = append(relArgs, relToRoot)
	}
	if len(relArgs) == 0 {
		c.JSON(400, gin.H{"error": "no valid paths"})
		return
	}

	var cmd *exec.Cmd
	switch req.Format {
	case "zip":
		// zip -r output.zip file1 dir2 ...  （从 storageRoot 执行）
		args := append([]string{"-r", outAbs}, relArgs...)
		cmd = exec.Command("zip", args...)
	case "tar.gz":
		// tar -czf output.tar.gz file1 dir2 ...
		args := append([]string{"-czf", outAbs}, relArgs...)
		cmd = exec.Command("tar", args...)
	case "tar":
		// tar -cf output.tar file1 dir2 ...
		args := append([]string{"-cf", outAbs}, relArgs...)
		cmd = exec.Command("tar", args...)
	}

	// 关键：从 storageRoot 执行，保证压缩包内路径是相对路径
	cmd.Dir = storageRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = io.Discard

	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"error": "compress failed: " + stderr.String()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "output": req.Output})
}

func (h *Handler) DecompressFile(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
		Dir  string `json:"dir"`
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
	var cmd *exec.Cmd
	switch {
	case strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz"):
		// tar -xzf file.tar.gz -C destdir
		cmd = exec.Command("tar", "-xzf", absPath, "-C", absDestDir)
	case strings.HasSuffix(name, ".tar"):
		// tar -xf file.tar -C destdir
		cmd = exec.Command("tar", "-xf", absPath, "-C", absDestDir)
	case strings.HasSuffix(name, ".gz"):
		// gzip -dk file.gz 解压到原目录，再 mv 到目标目录
		// 直接用 gzip -c -d 输出到目标文件，不占主进程内存
		outName := strings.TrimSuffix(filepath.Base(absPath), ".gz")
		outPath := filepath.Join(absDestDir, outName)
		cmd = exec.Command("sh", "-c", fmt.Sprintf("gzip -c -d %q > %q", absPath, outPath))
	case strings.HasSuffix(name, ".zip"):
		// unzip -o file.zip -d destdir
		cmd = exec.Command("unzip", "-o", absPath, "-d", absDestDir)
	default:
		c.JSON(400, gin.H{"error": "unsupported archive format"})
		return
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = io.Discard
	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"error": "decompress failed: " + stderr.String()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// ── FetchURL（SSRF 防护加强版）────────────────────────────────────────────────

func (h *Handler) FetchURL(c *gin.Context) {
	var req struct {
		URL      string `json:"url" binding:"required"`
		Filename string `json:"filename"`
		Dir      string `json:"dir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := validatePublicURL(req.URL); err != nil {
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
	absDir, err = filepath.Abs(absDir)
	if err != nil {
		c.JSON(500, gin.H{"error": "cannot resolve absolute path: " + err.Error()})
		return
	}
	if err := os.MkdirAll(absDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	hasWget := func() bool { _, e := exec.LookPath("wget"); return e == nil }()
	hasCurl := func() bool { _, e := exec.LookPath("curl"); return e == nil }()
	if !hasWget && !hasCurl {
		c.JSON(500, gin.H{"error": "neither wget nor curl is available on this system"})
		return
	}
	var cmd *exec.Cmd
	outPath := filepath.Join(absDir, req.Filename)
	if req.Filename != "" {
		if hasWget {
			cmd = exec.Command("wget", "-q", "-O", outPath, req.URL)
		} else {
			cmd = exec.Command("curl", "-L", "-o", outPath, req.URL)
		}
	} else {
		if hasWget {
			cmd = exec.Command("wget", "-q", "-P", absDir, req.URL)
		} else {
			cmd = exec.Command("curl", "-L", "--output-dir", absDir, "-O", req.URL)
		}
	}
	cmd.Dir = absDir
	// 用 Run() 而非 CombinedOutput()：后者会把 stdout/stderr 全量读入内存 buffer。
	// wget/curl 下载文件时 stdout 可能就是文件内容本身（未指定 -O 时），
	// 用 Run() + 丢弃 stdout/stderr 可避免大文件被意外吸入内存。
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"error": "fetch failed: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// validatePublicURL DNS 解析后校验 IP，防止 DNS 重绑定和 IP 格式绕过
func validatePublicURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http/https URLs are allowed")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host")
	}
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("requests to private/internal addresses are not allowed")
	}
	// DNS 解析，对所有返回的 IP 做校验
	addrs, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("cannot resolve host: %w", err)
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if isPrivateIP(ip) {
			return fmt.Errorf("requests to private/internal addresses are not allowed")
		}
	}
	return nil
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"100.64.0.0/10", // CGNAT
		"fc00::/7",      // IPv6 ULA
		"fe80::/10",     // IPv6 link-local
		"0.0.0.0/8",
	}
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// ── serveRaw ──────────────────────────────────────────────────────────────────

func browserPreviewable(mimeType string) bool {
	if strings.HasPrefix(mimeType, "text/") ||
		strings.HasPrefix(mimeType, "image/") ||
		strings.HasPrefix(mimeType, "audio/") ||
		strings.HasPrefix(mimeType, "video/") {
		return true
	}
	switch mimeType {
	case "application/pdf", "application/json", "application/javascript",
		"application/x-javascript", "application/xml", "application/xhtml+xml",
		"application/atom+xml", "application/rss+xml", "application/svg+xml":
		return true
	}
	return strings.Contains(mimeType, "xml") || strings.Contains(mimeType, "javascript")
}

var extMimeOverride = map[string]string{
	".md": "text/plain; charset=utf-8", ".markdown": "text/plain; charset=utf-8",
	".yaml": "text/plain; charset=utf-8", ".yml": "text/plain; charset=utf-8",
	".toml": "text/plain; charset=utf-8", ".ini": "text/plain; charset=utf-8",
	".conf": "text/plain; charset=utf-8", ".env": "text/plain; charset=utf-8",
	".log": "text/plain; charset=utf-8", ".sh": "text/plain; charset=utf-8",
	".bash": "text/plain; charset=utf-8", ".zsh": "text/plain; charset=utf-8",
	".fish": "text/plain; charset=utf-8", ".dockerfile": "text/plain; charset=utf-8",
	".go": "text/plain; charset=utf-8", ".py": "text/plain; charset=utf-8",
	".rs": "text/plain; charset=utf-8", ".rb": "text/plain; charset=utf-8",
	".java": "text/plain; charset=utf-8", ".c": "text/plain; charset=utf-8",
	".cpp": "text/plain; charset=utf-8", ".h": "text/plain; charset=utf-8",
	".ts": "text/plain; charset=utf-8", ".tsx": "text/plain; charset=utf-8",
	".jsx": "text/plain; charset=utf-8", ".vue": "text/plain; charset=utf-8",
	".swift": "text/plain; charset=utf-8", ".kt": "text/plain; charset=utf-8",
	".lua": "text/plain; charset=utf-8", ".r": "text/plain; charset=utf-8",
	".sql": "text/plain; charset=utf-8", ".graphql": "text/plain; charset=utf-8",
	".avif": "image/avif", ".webp": "image/webp",
}

func serveRaw(c *gin.Context, abs string) {
	ext := strings.ToLower(filepath.Ext(abs))
	mimeType, ok := extMimeOverride[ext]
	if !ok {
		mimeType = mime.TypeByExtension(ext)
	}
	if mimeType == "" {
		mimeType = "text/plain; charset=utf-8"
	}
	name := filepath.Base(abs)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "public, max-age=3600")
	baseMime := strings.TrimSpace(strings.SplitN(mimeType, ";", 2)[0])
	c.Header("Content-Type", mimeType)
	if browserPreviewable(baseMime) {
		c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, name))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	}
	c.File(abs)
}

// ── WebDAV ────────────────────────────────────────────────────────────────────

func (h *Handler) GetWebDAVSettings2(c *gin.Context) {
	h.GetWebDAVSettings(c)
}

func (h *Handler) webdavRoot() string {
	var s auth.Settings
	h.db.First(&s)
	base := h.files.Root()
	sub := strings.TrimSpace(s.WebDAVSubPath)
	if sub == "" {
		sub = "webdav"
	}
	sub = filepath.Clean(strings.TrimLeft(sub, "/"))
	root := filepath.Join(base, sub)
	os.MkdirAll(root, 0755)
	return root
}

func (h *Handler) WebDAVMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var s auth.Settings
		h.db.First(&s)
		if !s.WebDAVEnabled {
			c.Header("WWW-Authenticate", `Basic realm="CloudOne WebDAV"`)
			c.AbortWithStatus(503)
			return
		}
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="CloudOne WebDAV"`)
			c.AbortWithStatus(401)
			return
		}
		// 校验用户名
		expectedUser := s.WebDAVUsername
		if expectedUser == "" {
			var user auth.User
			if h.db.First(&user).Error == nil {
				expectedUser = user.Username
			}
		}
		if username != expectedUser {
			c.Header("WWW-Authenticate", `Basic realm="CloudOne WebDAV"`)
			c.AbortWithStatus(401)
			return
		}
		// 校验密码（从加密存储中解密 bcrypt hash 后比对）
		if s.WebDAVPasswordEnc != "" {
			hash, err := s.GetWebDAVPasswordHash()
			if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
				c.Header("WWW-Authenticate", `Basic realm="CloudOne WebDAV"`)
				c.AbortWithStatus(401)
				return
			}
		} else {
			// 未设置独立密码：回落到 CloudOne 账户密码
			var user auth.User
			if h.db.Where("username = ?", username).First(&user).Error != nil {
				c.Header("WWW-Authenticate", `Basic realm="CloudOne WebDAV"`)
				c.AbortWithStatus(401)
				return
			}
			if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
				c.Header("WWW-Authenticate", `Basic realm="CloudOne WebDAV"`)
				c.AbortWithStatus(401)
				return
			}
		}
		c.Next()
	}
}

func (h *Handler) WebDAVHandler(c *gin.Context) {
	var s auth.Settings
	h.db.First(&s)
	if !s.WebDAVEnabled {
		c.Status(503)
		return
	}
	root := h.webdavRoot()
	stripPrefix := "/dav"
	method := c.Request.Method
	rawPath := c.Request.URL.Path
	relPath := strings.TrimPrefix(rawPath, stripPrefix)
	if relPath == "" {
		relPath = "/"
	}
	absTarget := filepath.Join(root, filepath.FromSlash(relPath))
	if !strings.HasPrefix(absTarget, root) {
		c.Status(403)
		return
	}
	switch method {
	case "OPTIONS":
		c.Header("DAV", "1, 2")
		c.Header("Allow", "OPTIONS, GET, HEAD, PUT, DELETE, MKCOL, COPY, MOVE, PROPFIND, PROPPATCH, LOCK, UNLOCK")
		c.Header("MS-Author-Via", "DAV")
		c.Status(200)
	case "PROPFIND":
		h.davPropfind(c, root, relPath, absTarget)
	case "GET", "HEAD":
		info, err := os.Stat(absTarget)
		if err != nil {
			c.Status(404)
			return
		}
		if info.IsDir() {
			h.davPropfind(c, root, relPath, absTarget)
			return
		}
		c.File(absTarget)
	case "PUT":
		if err := os.MkdirAll(filepath.Dir(absTarget), 0755); err != nil {
			c.Status(500)
			return
		}
		f, err := os.Create(absTarget)
		if err != nil {
			c.Status(500)
			return
		}
		defer f.Close()
		io.Copy(f, c.Request.Body)
		c.Status(201)
	case "DELETE":
		if err := os.RemoveAll(absTarget); err != nil {
			c.Status(500)
			return
		}
		c.Status(204)
	case "MKCOL":
		if err := os.MkdirAll(absTarget, 0755); err != nil {
			c.Status(500)
			return
		}
		c.Status(201)
	case "COPY", "MOVE":
		dest := c.Request.Header.Get("Destination")
		if dest == "" {
			c.Status(400)
			return
		}
		destPath := dest
		if idx := strings.Index(dest, "/dav"); idx >= 0 {
			destPath = dest[idx+len("/dav"):]
		}
		absDest := filepath.Join(root, filepath.FromSlash(destPath))
		if !strings.HasPrefix(absDest, root) {
			c.Status(403)
			return
		}
		os.MkdirAll(filepath.Dir(absDest), 0755)
		if method == "COPY" {
			if err := copyPath(absTarget, absDest); err != nil {
				c.Status(500)
				return
			}
			c.Status(201)
		} else {
			if err := os.Rename(absTarget, absDest); err != nil {
				c.Status(500)
				return
			}
			c.Status(201)
		}
	case "LOCK":
		b := make([]byte, 8)
		rand.Read(b)
		token := "urn:uuid:cloudone-lock-" + hex.EncodeToString(b)
		c.Header("Lock-Token", "<"+token+">")
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.String(200, `<?xml version="1.0" encoding="utf-8"?>
<D:prop xmlns:D="DAV:"><D:lockdiscovery><D:activelock>
<D:locktype><D:write/></D:locktype>
<D:lockscope><D:exclusive/></D:lockscope>
<D:depth>infinity</D:depth>
<D:timeout>Second-3600</D:timeout>
<D:locktoken><D:href>`+token+`</D:href></D:locktoken>
</D:activelock></D:lockdiscovery></D:prop>`)
	case "UNLOCK":
		c.Status(204)
	case "PROPPATCH":
		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.String(207, `<?xml version="1.0" encoding="utf-8"?><D:multistatus xmlns:D="DAV:"></D:multistatus>`)
	default:
		c.Status(405)
	}
}

func (h *Handler) davPropfind(c *gin.Context, root, relPath, absTarget string) {
	info, err := os.Stat(absTarget)
	if err != nil {
		c.Status(404)
		return
	}
	depth := c.Request.Header.Get("Depth")
	if depth == "" {
		depth = "1"
	}
	var entries []os.FileInfo
	if info.IsDir() && depth != "0" {
		dirEntries, err := os.ReadDir(absTarget)
		if err != nil {
			c.Status(500)
			return
		}
		for _, de := range dirEntries {
			fi, err := de.Info()
			if err == nil {
				entries = append(entries, fi)
			}
		}
	}
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Status(207)
	c.Writer.WriteString(`<?xml version="1.0" encoding="utf-8"?><D:multistatus xmlns:D="DAV:">`)
	writePropfindEntry(c.Writer, relPath, info)
	if info.IsDir() && depth != "0" {
		childBase := relPath
		if !strings.HasSuffix(childBase, "/") {
			childBase += "/"
		}
		for _, child := range entries {
			childRel := childBase + child.Name()
			writePropfindEntry(c.Writer, childRel, child)
		}
	}
	c.Writer.WriteString(`</D:multistatus>`)
}

func writePropfindEntry(w gin.ResponseWriter, relPath string, info os.FileInfo) {
	href := "/dav" + relPath
	if info.IsDir() && !strings.HasSuffix(href, "/") {
		href += "/"
	}
	modTime := info.ModTime().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	resourceType := ""
	contentLength := ""
	if info.IsDir() {
		resourceType = "<D:resourcetype><D:collection/></D:resourcetype>"
	} else {
		resourceType = "<D:resourcetype/>"
		contentLength = fmt.Sprintf("<D:getcontentlength>%d</D:getcontentlength>", info.Size())
	}
	fmt.Fprintf(w, `<D:response><D:href>%s</D:href><D:propstat><D:prop>%s%s<D:getlastmodified>%s</D:getlastmodified><D:displayname>%s</D:displayname></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat></D:response>`,
		href, resourceType, contentLength, modTime, info.Name())
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
