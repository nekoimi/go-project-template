package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/handler"
	"github.com/nekoimi/go-project-template/internal/infrastructure/database"
	"github.com/nekoimi/go-project-template/internal/infrastructure/logger"
	"github.com/nekoimi/go-project-template/internal/scheduler"
	"github.com/nekoimi/go-project-template/internal/storage"
	"github.com/nekoimi/go-project-template/internal/storage/local"
	"github.com/nekoimi/go-project-template/internal/storage/minio"
	ws "github.com/nekoimi/go-project-template/internal/websocket"
)

type App struct {
	Engine   *gin.Engine
	Config   *config.Config
	Logger   *zap.Logger
	DB       *gorm.DB
	Storage  storage.FileStorage
	WSManager *ws.Manager
	Scheduler *scheduler.Scheduler
}

func Initialize(configPath string) (*App, func(), error) {
	// 1. Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Logger
	log, err := logger.NewLogger(cfg.Server.Mode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// 3. Database
	db, err := database.NewPostgresDB(cfg.Database, log)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}

	// 4. Storage
	var fileStorage storage.FileStorage
	switch cfg.Storage.Driver {
	case "minio":
		fileStorage, err = minio.New(cfg.Storage)
		if err != nil {
			log.Fatal("failed to create minio storage", zap.Error(err))
		}
	default:
		fileStorage = local.New(cfg.Storage)
	}

	// 5. WebSocket manager
	wsManager := ws.NewManager(log)

	// 6. Setup router
	router := handler.SetupRouter(cfg, log, db, fileStorage, wsManager)

	// 7. Scheduler (optional)
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
			sqlDB.Close()
		}
		log.Sync()
	}

	return app, cleanup, nil
}
