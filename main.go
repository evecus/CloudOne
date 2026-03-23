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
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(50)
	}
	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(512 << 20) // 512 MiB
	}

	// ── 命令行参数解析 ────────────────────────────────────────────────────────
	var configFlag string
	var dirFlag string
	flag.StringVar(&configFlag, "config", "", "数据配置目录 (存放 db, keys, conf.ini)")
	flag.StringVar(&dirFlag, "dir", "", "文件存储目录 (存放用户上传的文件)")
	flag.Parse()

	var dataDir string
	var storageDir string

	// 逻辑判断：按照用户要求的优先级设定目录
	if configFlag != "" && dirFlag != "" {
		dataDir = configFlag
		storageDir = dirFlag
	} else if configFlag != "" && dirFlag == "" {
		dataDir = configFlag
		storageDir = filepath.Join(configFlag, "storage")
	} else if configFlag == "" && dirFlag != "" {
		dataDir = "./data"
		storageDir = dirFlag
	} else {
		dataDir = "./data"
		storageDir = filepath.Join(dataDir, "storage")
	}

	// 创建必要的目录
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}
	// 预创建存储目录（如果不存在）
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatal("Failed to create initial storage directory:", err)
	}

	// ── 配置文件与数据库 ──────────────────────────────────────────────────────
	confPath := filepath.Join(dataDir, "conf.ini")
	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	masterKeyPath := filepath.Join(dataDir, "master.key")
	masterKey := os.Getenv("CLOUDONE_MASTER_KEY")
	if masterKey == "" {
		raw, err := os.ReadFile(masterKeyPath)
		if err != nil {
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
	auth.SetMasterKey(masterKey)

	db, err := auth.InitDB(filepath.Join(dataDir, "cloudone.db"))
	if err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	// ── 确定最终存储目录 ──────────────────────────────────────────────────────
	var settings auth.Settings
	db.First(&settings)

	// 如果命令行明确指定了 --dir，则强制覆盖数据库中的设置并更新数据库
	if dirFlag != "" {
		storageDir = dirFlag
		if settings.StorageDir != storageDir {
			settings.StorageDir = storageDir
			db.Save(&settings)
			log.Println("Storage directory updated from command line:", storageDir)
		}
	} else if settings.StorageDir != "" {
		// 如果命令行没指明，且数据库里有旧记录，则沿用数据库的
		storageDir = settings.StorageDir
	}

	// 再次确保最终确定的存储目录存在
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatal("Failed to ensure storage directory:", err)
	}
	log.Println("Data directory:", dataDir)
	log.Println("Storage directory:", storageDir)

	// ── JWT Secret ──────────────────────────────────────────────────────────
	jwtSecret := os.Getenv("CLOUDONE_JWT_SECRET")
	if jwtSecret == "" {
		secret, err := settings.GetJWTSecret()
		if err != nil {
			log.Println("Warning: failed to decrypt JWT secret, regenerating:", err)
		}
		jwtSecret = secret
	}
	if jwtSecret == "" {
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

	allowedOrigin := os.Getenv("CLOUDONE_ORIGIN")
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if allowedOrigin == "" {
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

	davMethods := []string{"GET", "HEAD", "PUT", "DELETE", "OPTIONS", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK"}
	for _, method := range davMethods {
		r.Handle(method, "/dav", h.WebDAVMiddleware(), h.WebDAVHandler)
		r.Handle(method, "/dav/*path", h.WebDAVMiddleware(), h.WebDAVHandler)
	}

	r.GET("/public", h.ListPublicFiles)
	r.GET("/pub/list", h.ListPublicFilesOpen)
	r.GET("/pub/dl", h.DownloadPublicFile)
	r.GET("/raw/*path", h.ServePublicFile)
	r.GET("/s/:code/raw", h.ServeRaw)

	// 前端静态文件处理
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
