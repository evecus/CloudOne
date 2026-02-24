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
	// Create data directory next to binary
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Init DB
	db, err := auth.InitDB(dataDir + "/cloudone.db")
	if err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	// Default storage dir
	storageDir := dataDir + "/storage"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatal("Failed to create storage directory:", err)
	}

	fileManager := files.NewManager(storageDir, db)
	shareManager := files.NewShareManager(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
	}))

	h := handler.New(db, fileManager, shareManager)

	// API routes
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

	// Public file access
	r.GET("/public", h.ListPublicFiles)
	r.GET("/raw/*path", h.ServePublicFile)
	r.GET("/s/:code", h.AccessShare)
	r.GET("/s/:code/raw", h.DownloadShare)

	// Serve frontend
	sub, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	r.NoRoute(func(c *gin.Context) {
		// Try static file first
		path := c.Request.URL.Path
		if path == "/" || path == "" {
			c.FileFromFS("index.html", http.FS(sub))
			return
		}
		f, err := sub.Open(path[1:])
		if err != nil {
			c.FileFromFS("index.html", http.FS(sub))
			return
		}
		f.Close()
		c.FileFromFS(path[1:], http.FS(sub))
	})

	log.Println("CloudOne starting on :6677")
	if err := r.Run(":6677"); err != nil {
		log.Fatal(err)
	}
}
