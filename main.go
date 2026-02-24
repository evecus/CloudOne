package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/cloudone/cloudone/internal/auth"
	"github.com/cloudone/cloudone/internal/files"
	"github.com/cloudone/cloudone/internal/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	db, err := auth.InitDB(dataDir + "/cloudone.db")
	if err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	storageDir := dataDir + "/storage"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatal("Failed to create storage directory:", err)
	}

	fileManager := files.NewManager(storageDir, db)
	shareManager := files.NewShareManager(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
    r.RedirectTrailingSlash = false  // 加这一行
    r.RedirectFixedPath = false      // 加这一行

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
	}))

	h := handler.New(db, fileManager, shareManager)

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

			authed.POST("/share", h.CreateShare)
			authed.GET("/share", h.ListShares)
			authed.DELETE("/share/:id", h.DeleteShare)
		}
	}

	r.GET("/public", h.ListPublicFiles)
	r.GET("/raw/*path", h.ServePublicFile)
	r.GET("/s/:code", h.AccessShare)
	r.GET("/s/:code/raw", h.DownloadShare)

	// 把前端 dist 挂载为子文件系统
	sub, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	staticFS := http.FS(sub)

	// 明确注册 GET / 避免 gin 触发 301
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", staticFS)
	})

	// 所有其他路由：先尝试静态文件，找不到就返回 index.html（SPA fallback）
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 去掉开头的 /
		filePath := path
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		// 尝试打开静态文件
		f, err := sub.Open(filePath)
		if err == nil {
			f.Close()
			c.FileFromFS(filePath, staticFS)
			return
		}
		// 找不到就返回 index.html（Vue Router 接管）
		c.FileFromFS("index.html", staticFS)
	})

	log.Println("CloudOne starting on :6677")
	if err := r.Run(":6677"); err != nil {
		log.Fatal(err)
	}
}
