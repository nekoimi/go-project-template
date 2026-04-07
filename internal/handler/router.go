package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	v1 "github.com/nekoimi/go-project-template/internal/handler/v1"
	"github.com/nekoimi/go-project-template/internal/handler/middleware"
	"github.com/nekoimi/go-project-template/internal/repository"
	"github.com/nekoimi/go-project-template/internal/service"
	"github.com/nekoimi/go-project-template/internal/storage"
	ws "github.com/nekoimi/go-project-template/internal/websocket"
)

func SetupRouter(cfg *config.Config, logger *zap.Logger, db *gorm.DB, fileStorage storage.FileStorage, wsManager *ws.Manager) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	// Middleware
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Local file serving
	if cfg.Storage.Driver == "local" {
		r.Static("/uploads", cfg.Storage.Local.UploadDir)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)

	// Services
	jwtExpire := time.Duration(cfg.JWT.ExpireHours) * time.Hour
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, jwtExpire)
	userService := service.NewUserService(userRepo)
	fileService := service.NewFileService(fileStorage)

	// Handlers
	authHandler := v1.NewAuthHandler(authService)
	userHandler := v1.NewUserHandler(userService)
	uploadHandler := v1.NewUploadHandler(fileService)
	wsHandler := v1.NewWSHandler(ws.NewWSHandler(wsManager, cfg.JWT.Secret, logger))

	// API v1 routes
	api := r.Group("/v1")
	{
		// Auth (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// Users
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
			}

			// Upload
			upload := protected.Group("/upload")
			{
				upload.POST("/single", uploadHandler.UploadSingle)
				upload.POST("/multiple", uploadHandler.UploadMultiple)
			}
		}
	}

	// WebSocket
	r.GET("/ws/v1/chat", wsHandler.Upgrade)

	return r
}
