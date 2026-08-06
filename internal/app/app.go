package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/handler"
	"github.com/nekoimi/go-project-template/internal/pkg/database"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/pkg/logger"
	"github.com/nekoimi/go-project-template/internal/pkg/timeutil"
	"github.com/nekoimi/go-project-template/internal/scheduler"
	"github.com/nekoimi/go-project-template/internal/storage"
	storagefactory "github.com/nekoimi/go-project-template/internal/storage/factory"
	ws "github.com/nekoimi/go-project-template/internal/websocket"
)

type App struct {
	Engine     *gin.Engine
	Config     *config.Config
	Logger     *zap.Logger
	DB         *gorm.DB
	Storage    storage.FileStorage
	WSManager  *ws.Manager
	Scheduler  *scheduler.Scheduler
	HTTPServer *http.Server
	httpErr    chan error
	Modules    []framework.Module
	ModuleCtx  *framework.ModuleContext
}

func (a *App) Boot(ctx context.Context) error {
	if a == nil {
		return nil
	}
	return framework.BootModules(ctx, a.ModuleCtx, a.Modules...)
}

func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}
	return framework.ShutdownModules(ctx, a.ModuleCtx, a.Modules...)
}

func Initialize(configPath string) (*App, func(), error) {
	return initialize(configPath, registeredModules())
}

func initialize(configPath string, modules []framework.Module) (*App, func(), error) {
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
	fileStorage, err := storagefactory.New(cfg.Storage)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// 7. WebSocket manager
	wsManager := ws.NewManager(log)

	// 8. Health checks
	health := framework.NewHealthRegistry()
	health.Register("database", func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	})
	events := framework.NewEventBus()

	// 9. Scheduler
	var sched *scheduler.Scheduler
	if cfg.Scheduler.Enabled {
		sched = scheduler.New(cfg.Scheduler, log, db)
	}

	// 10. Setup base router
	router := handler.SetupRouter(cfg, log, health)

	// 11. Register feature modules
	moduleCtx := framework.NewModuleContext(cfg, log, db, router, sched, fileStorage, wsManager, health, events)
	if err := framework.RegisterModules(moduleCtx, modules...); err != nil {
		return nil, nil, err
	}

	app := &App{
		Engine:    router.Engine,
		Config:    cfg,
		Logger:    log,
		DB:        db,
		Storage:   fileStorage,
		WSManager: wsManager,
		Scheduler: sched,
		Modules:   modules,
		ModuleCtx: moduleCtx,
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
