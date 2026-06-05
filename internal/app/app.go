package app

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/handler"
	v1 "github.com/nekoimi/go-project-template/internal/handler/v1"
	"github.com/nekoimi/go-project-template/internal/pkg/database"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/pkg/logger"
	"github.com/nekoimi/go-project-template/internal/pkg/timeutil"
	"github.com/nekoimi/go-project-template/internal/repository"
	"github.com/nekoimi/go-project-template/internal/scheduler"
	"github.com/nekoimi/go-project-template/internal/service"
	"github.com/nekoimi/go-project-template/internal/storage"
	"github.com/nekoimi/go-project-template/internal/storage/local"
	"github.com/nekoimi/go-project-template/internal/storage/minio"
	ws "github.com/nekoimi/go-project-template/internal/websocket"
)

type App struct {
	Engine    *gin.Engine
	Config    *config.Config
	Logger    *zap.Logger
	DB        *gorm.DB
	Storage   storage.FileStorage
	WSManager *ws.Manager
	Scheduler *scheduler.Scheduler
}

func Initialize(configPath string) (*App, func(), error) {
	// 1. Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Timezone
	if err := timeutil.SetGlobalLocation(cfg.Server.Timezone); err != nil {
		return nil, nil, fmt.Errorf("failed to set timezone: %w", err)
	}

	// 3. ID Generator (Snowflake)
	if err := idgen.Init(cfg.Snowflake.NodeID); err != nil {
		return nil, nil, fmt.Errorf("failed to init idgen: %w", err)
	}

	// 4. Logger
	log, err := logger.NewLogger(cfg.Server.Mode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// 5. Database
	db, err := database.NewPostgresDB(cfg.Database, log, cfg.Server.Mode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 6. Storage
	var fileStorage storage.FileStorage
	switch cfg.Storage.Driver {
	case "minio":
		fileStorage, err = minio.New(cfg.Storage)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create minio storage: %w", err)
		}
	default:
		fileStorage = local.New(cfg.Storage)
	}

	// 7. WebSocket manager
	wsManager := ws.NewManager(log)

	// 8. Repositories
	userRepo := repository.NewUserRepository(db)

	// 9. Services
	jwtExpire := time.Duration(cfg.JWT.ExpireHours) * time.Hour
	authService := service.NewAuthService(userRepo, db, cfg.JWT.Secret, jwtExpire)
	userService := service.NewUserService(userRepo)
	fileService := service.NewFileService(fileStorage, cfg.Storage.Local.AllowedExts, cfg.Storage.Local.AllowedMIMEs)

	// 10. Handlers
	authHandler := v1.NewAuthHandler(authService, log)
	userHandler := v1.NewUserHandler(userService, log)
	uploadHandler := v1.NewUploadHandler(fileService, log)

	var wsHandler *v1.WSHandler
	if cfg.Websocket.Enabled {
		wsHandler = v1.NewWSHandler(ws.NewWSHandler(wsManager, cfg.JWT.Secret, log, cfg.Server.AllowedOrigins, cfg.Websocket))
	}

	// 11. Setup router
	router := handler.SetupRouter(cfg, log, db, authHandler, userHandler, uploadHandler, wsHandler)

	// 12. Scheduler (optional)
	var sched *scheduler.Scheduler
	if cfg.Scheduler.Enabled {
		sched = scheduler.New(cfg.Scheduler, log, db)
		sched.RegisterJobs()
	}

	app := &App{
		Engine:    router,
		Config:    cfg,
		Logger:    log,
		DB:        db,
		Storage:   fileStorage,
		WSManager: wsManager,
		Scheduler: sched,
	}

	cleanup := func() {
		log.Info("cleaning up resources")
		if sqlDB, err := db.DB(); err == nil {
			if cerr := sqlDB.Close(); cerr != nil {
				log.Warn("failed to close database", zap.Error(cerr))
			}
		}
		_ = log.Sync()
	}

	return app, cleanup, nil
}
