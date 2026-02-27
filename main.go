package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cloudone/cloudone/internal/auth"
	"github.com/cloudone/cloudone/internal/config"
	"github.com/cloudone/cloudone/internal/files"
	"github.com/cloudone/cloudone/internal/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

func init() {
	mime.AddExtensionType(".js",    "application/javascript; charset=utf-8")
	mime.AddExtensionType(".mjs",   "application/javascript; charset=utf-8")
	mime.AddExtensionType(".css",   "text/css; charset=utf-8")
	mime.AddExtensionType(".html",  "text/html; charset=utf-8")
	mime.AddExtensionType(".json",  "application/json")
	mime.AddExtensionType(".svg",   "image/svg+xml")
	mime.AddExtensionType(".png",   "image/png")
	mime.AddExtensionType(".jpg",   "image/jpeg")
	mime.AddExtensionType(".jpeg",  "image/jpeg")
	mime.AddExtensionType(".gif",   "image/gif")
	mime.AddExtensionType(".webp",  "image/webp")
	mime.AddExtensionType(".ico",   "image/x-icon")
	mime.AddExtensionType(".woff",  "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".ttf",   "font/ttf")
	mime.AddExtensionType(".map",   "application/json")
}

func main() {
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	confPath := dataDir + "/conf.ini"

	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := auth.InitDB(dataDir + "/cloudone.db")
	if err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	storageDir := cfg.StorageDir
	if storageDir == "" {
		var settings auth.Settings
		db.First(&settings)
		storageDir = settings.StorageDir
	}
	if storageDir == "" {
		storageDir = dataDir + "/storage"
	}
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatal("Failed to create storage directory:", err)
	}
	log.Println("Storage directory:", storageDir)

	cfg.StorageDir = storageDir
	var settings auth.Settings
	db.First(&settings)
	if cfg.Lang == "" {
		cfg.Lang = settings.Lang
	}
	var user auth.User
	if db.First(&user).Error == nil {
		cfg.Username = user.Username
		cfg.Password = user.Password
	}
	// 生成并持久化 JWT secret（首次启动时自动生成，之后从 conf.ini 读取）
	if cfg.JWTSecret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatal("Failed to generate JWT secret:", err)
		}
		cfg.JWTSecret = hex.EncodeToString(b)
	}
	handler.SetJWTSecret(cfg.JWTSecret)

	if err := config.Save(confPath, cfg); err != nil {
		log.Println("Warning: could not write conf.ini:", err)
	}

	fileManager := files.NewManager(storageDir, db)
	shareManager := files.NewShareManager(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// CORS：明确指定允许的来源（不使用 AllowAllOrigins + AllowCredentials）
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// 开发环境允许 localhost，生产环境根据需要调整
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Depth", "Destination", "Lock-Token", "Overwrite", "Timeout", "If"},
		ExposeHeaders:    []string{"Content-Disposition", "DAV", "Lock-Token"},
		AllowCredentials: false, // 使用 Bearer token，不需要 credentials
	}))

	h := handler.New(db, fileManager, shareManager, confPath)

	api := r.Group("/api")
	{
		api.POST("/auth/setup", h.Setup)
		api.GET("/auth/status", h.AuthStatus)
		api.POST("/auth/login", h.Login)

		authed := api.Group("/")
		authed.Use(h.AuthMiddleware())
		{
			authed.GET("/user", h.GetUser)
			authed.PUT("/user", h.UpdateUser)
			authed.GET("/settings", h.GetSettings)
			authed.PUT("/settings", h.UpdateSettings)

			authed.GET("/files", h.ListFiles)
			authed.POST("/files/upload", h.UploadFile)
			authed.DELETE("/files", h.DeleteFile)
			authed.POST("/files/mkdir", h.Mkdir)
			authed.POST("/files/move", h.MoveFile)
			authed.POST("/files/copy", h.CopyFile)
			authed.GET("/files/download", h.DownloadFile)
			authed.POST("/files/create", h.CreateFile)
			authed.PUT("/files/visibility", h.SetVisibility)
			authed.GET("/files/content", h.GetFileContent)
			authed.PUT("/files/content", h.UpdateFileContent)
			authed.GET("/files/permission", h.GetPermission)
			authed.PUT("/files/permission", h.SetPermission)
			authed.POST("/files/batch-delete", h.BatchDelete)
			authed.POST("/files/batch-download", h.BatchDownload)
			authed.POST("/files/batch-move", h.BatchMove)
			authed.POST("/files/batch-copy", h.BatchCopy)
			authed.POST("/files/compress", h.CompressFiles)
			authed.POST("/files/decompress", h.DecompressFile)
			authed.POST("/files/fetch-url", h.FetchURL)
			authed.GET("/files/search", h.SearchFiles)
			authed.POST("/files/upload-folder", h.UploadFolder)
			authed.GET("/files/dirtree", h.ListDirTree)

			authed.POST("/share", h.CreateShare)
			authed.GET("/share", h.ListShares)
			authed.DELETE("/share/:id", h.DeleteShare)

			// WebDAV 设置（已登录用户管理）
			authed.GET("/webdav/settings", h.GetWebDAVSettings)
			authed.PUT("/webdav/settings", h.UpdateWebDAVSettings)

		}

		// 分享码 API（无需登录）
		api.GET("/s/:code", h.AccessShare)
		api.GET("/s/:code/download", h.DownloadShare)
	}

	// WebDAV 协议路由（Basic Auth，独立于 JWT）
	davMethods := []string{"GET", "HEAD", "PUT", "DELETE", "OPTIONS", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK"}
	for _, method := range davMethods {
		r.Handle(method, "/dav", h.WebDAVMiddleware(), h.WebDAVHandler)
		r.Handle(method, "/dav/*path", h.WebDAVMiddleware(), h.WebDAVHandler)
	}

	// 公开文件 API（无需登录）
	r.GET("/public", h.ListPublicFiles)        // 公开文件列表（已登录页面用）
	r.GET("/pub/list", h.ListPublicFilesOpen)  // 公开文件列表（未登录前台用）
	r.GET("/pub/dl", h.DownloadPublicFile)     // 下载公开文件（无需登录）

	// /raw/* → 查看公开文件原始内容（类似 GitHub raw）
	r.GET("/raw/*path", h.ServePublicFile)

	// /s/:code/raw  → 通过分享码查看文件内容（类似 GitHub raw）
	r.GET("/s/:code/raw", h.ServeRaw)

	// /s/:code      → Vue Router 渲染 ShareView
	// /pub/*        → 前台公开文件浏览页面（Vue Router 渲染）

	sub, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	serveFile := func(c *gin.Context, fsPath string) bool {
		data, err := fs.ReadFile(sub, fsPath)
		if err != nil {
			return false
		}
		ext := filepath.Ext(fsPath)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		c.Data(http.StatusOK, mimeType, data)
		return true
	}

	serveIndex := func(c *gin.Context) {
		data, err := fs.ReadFile(sub, "index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}

	r.GET("/", serveIndex)

	r.GET("/assets/*filepath", func(c *gin.Context) {
		fsPath := "assets" + c.Param("filepath")
		if !serveFile(c, fsPath) {
			c.Status(http.StatusNotFound)
		}
	})

	r.NoRoute(func(c *gin.Context) {
		filePath := c.Request.URL.Path
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		if filePath != "" {
			if serveFile(c, filePath) {
				return
			}
		}
		serveIndex(c)
	})

	addr := cfg.Host + ":" + cfg.Port
	log.Println("CloudOne starting on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
