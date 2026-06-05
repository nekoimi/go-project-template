package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	v1 "github.com/nekoimi/go-project-template/internal/handler/v1"
	"github.com/nekoimi/go-project-template/internal/middleware"
	"github.com/nekoimi/go-project-template/internal/pkg/resp"
)

func SetupRouter(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	authHandler *v1.AuthHandler,
	userHandler *v1.UserHandler,
	uploadHandler *v1.UploadHandler,
	wsHandler *v1.WSHandler,
) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	// Middleware
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.CORS(cfg.Server.AllowedOrigins))

	// Rate limiting
	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimit(cfg.RateLimit.RPS, cfg.RateLimit.Burst))
	}

	// Health check (liveness)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Readiness check (DB ping)
	r.GET("/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(503, gin.H{"status": "not ready"})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Local file serving
	if cfg.Storage.Driver == "local" {
		r.Static("/uploads", cfg.Storage.Local.UploadDir)
	}

	// API v1 routes
	api := r.Group("/v1")
	{
		// Auth (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", resp.Handle(authHandler.Register, logger))
			auth.POST("/login", resp.Handle(authHandler.Login, logger))
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// Users
			users := protected.Group("/users")
			{
				users.GET("/profile", resp.Handle(userHandler.GetProfile, logger))
			}

			// Upload
			upload := protected.Group("/upload")
			{
				upload.POST("/single", resp.Handle(uploadHandler.UploadSingle, logger))
				upload.POST("/multiple", resp.Handle(uploadHandler.UploadMultiple, logger))
			}
		}
	}

	if cfg.Websocket.Enabled && wsHandler != nil {
		r.GET("/ws/v1/chat", wsHandler.Upgrade)
	}

	return r
}
