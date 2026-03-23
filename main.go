package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"flag"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"

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
	mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	mime.AddExtensionType(".mjs", "application/javascript; charset=utf-8")
	mime.AddExtensionType(".css", "text/css; charset=utf-8")
	mime.AddExtensionType(".html", "text/html; charset=utf-8")
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".gif", "image/gif")
	mime.AddExtensionType(".webp", "image/webp")
	mime.AddExtensionType(".ico", "image/x-icon")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".ttf", "font/ttf")
	mime.AddExtensionType(".map", "application/json")
}

func main() {
	// ── 内存管理配置 ──────────────────────────────────────────────────────────
	// 设置更激进的 GC 触发比例（默认100，改为50意味着堆增长50%就触发GC）
	// 可通过环境变量 GOGC 覆盖（如 GOGC=100 恢复默认）
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(50)
	}
	// 软内存上限：超过此值时 GC 会更频繁运行并积极归还内存给 OS。
	// 默认 512MB，可通过环境变量 GOMEMLIMIT 覆盖（如 GOMEMLIMIT=1GiB）
	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(512 << 20) // 512 MiB
	}

	var dataDir string
	flag.StringVar(&dataDir, "dir", "./data", "数据目录路径，例如 --dir /etc/cloudone/data")
	flag.Parse()

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	confPath := dataDir + "/conf.ini"
	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// ── Master Key ──────────────────────────────────────────────────────────
	// 用于加密数据库中的敏感字段（JWT Secret、WebDAV 密码哈希等）。
	// 优先从环境变量 CLOUDONE_MASTER_KEY 读取（生产环境推荐）。
	// 若未设置，自动生成并持久化到 data/master.key（权限 0600）。
	// 注意：master.key 丢失将导致所有加密数据无法解密，请妥善备份。
	masterKeyPath := dataDir + "/master.key"
	masterKey := os.Getenv("CLOUDONE_MASTER_KEY")
	if masterKey == "" {
		raw, err := os.ReadFile(masterKeyPath)
		if err != nil {
			// 首次启动：生成随机 master key
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				log.Fatal("Failed to generate master key:", err)
			}
			masterKey = hex.EncodeToString(b)
			if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0600); err != nil {
				log.Fatal("Failed to save master key:", err)
			}
			log.Println("Generated new master key, saved to", masterKeyPath)
		} else {
			masterKey = string(raw)
		}
	}
	// 注入到 auth 包，所有加解密操作均使用此密钥
	auth.SetMasterKey(masterKey)

	// ── 数据库 ───────────────────────────────────────────────────────────────
	db, err := auth.InitDB(dataDir + "/cloudone.db")
	if err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	// ── 存储目录 ─────────────────────────────────────────────────────────────
	var settings auth.Settings
	db.First(&settings)

	storageDir := settings.StorageDir
	if storageDir == "" {
		storageDir = dataDir + "/storage"
	}
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatal("Failed to create storage directory:", err)
	}
	log.Println("Storage directory:", storageDir)

	// ── JWT Secret ───────────────────────────────────────────────────────────
	// 优先从环境变量读取；否则从加密数据库读取；首次启动自动生成并加密存储。
	jwtSecret := os.Getenv("CLOUDONE_JWT_SECRET")
	if jwtSecret == "" {
		// 从数据库解密读取
		secret, err := settings.GetJWTSecret()
		if err != nil {
			log.Println("Warning: failed to decrypt JWT secret, regenerating:", err)
		}
		jwtSecret = secret
	}
	if jwtSecret == "" {
		// 首次启动：生成随机 JWT secret，加密后存入 DB
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatal("Failed to generate JWT secret:", err)
		}
		jwtSecret = hex.EncodeToString(b)
		if err := settings.SetJWTSecret(jwtSecret); err != nil {
			log.Fatal("Failed to encrypt JWT secret:", err)
		}
		db.Save(&settings)
		log.Println("Generated new JWT secret, stored encrypted in database")
	}
	handler.SetJWTSecret(jwtSecret)

	// ── conf.ini 只保存 host/port ────────────────────────────────────────────
	if err := config.Save(confPath, cfg); err != nil {
		log.Println("Warning: could not write conf.ini:", err)
	}

	fileManager := files.NewManager(storageDir, db)
	shareManager := files.NewShareManager(db)

	// ── HTTP 路由 ─────────────────────────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// CORS：限制为同源，生产环境可通过环境变量 CLOUDONE_ORIGIN 指定允许的来源
	allowedOrigin := os.Getenv("CLOUDONE_ORIGIN")
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if allowedOrigin == "" {
				// 未配置时：仅允许同源（空 origin 或与服务器地址相同）
				return origin == "" || origin == "null"
			}
			return origin == allowedOrigin
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Depth", "Destination", "Lock-Token", "Overwrite", "Timeout", "If"},
		ExposeHeaders:    []string{"Content-Disposition", "DAV", "Lock-Token"},
		AllowCredentials: false,
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

			authed.GET("/webdav/settings", h.GetWebDAVSettings)
			authed.PUT("/webdav/settings", h.UpdateWebDAVSettings)
		}

		api.GET("/s/:code", h.AccessShare)
		api.GET("/s/:code/download", h.DownloadShare)
	}

	// WebDAV 协议路由
	davMethods := []string{"GET", "HEAD", "PUT", "DELETE", "OPTIONS", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK"}
	for _, method := range davMethods {
		r.Handle(method, "/dav", h.WebDAVMiddleware(), h.WebDAVHandler)
		r.Handle(method, "/dav/*path", h.WebDAVMiddleware(), h.WebDAVHandler)
	}

	// 公开文件 API
	r.GET("/public", h.ListPublicFiles)
	r.GET("/pub/list", h.ListPublicFilesOpen)
	r.GET("/pub/dl", h.DownloadPublicFile)
	r.GET("/raw/*path", h.ServePublicFile)
	r.GET("/s/:code/raw", h.ServeRaw)

	// 前端静态文件
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
