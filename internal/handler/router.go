package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/middleware"
)

func SetupRouter(cfg *config.Config, logger *zap.Logger, health *framework.HealthRegistry) *framework.RouterContext {
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

	// Readiness check
	r.GET("/ready", func(c *gin.Context) {
		result := health.Check(c.Request.Context())
		if result.Status != "ready" {
			c.JSON(503, result)
			return
		}
		c.JSON(200, result)
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Local file serving
	if cfg.Storage.Driver == "local" {
		r.Static("/uploads", cfg.Storage.Local.UploadDir)
	}

	// API v1 routes
	api := r.Group("/v1")
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWT.Secret))

	return &framework.RouterContext{
		Engine:    r,
		API:       api,
		Protected: protected,
	}
}
