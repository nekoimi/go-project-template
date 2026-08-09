package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/pkg/database"
	"github.com/nekoimi/go-project-template/internal/pkg/idgen"
	"github.com/nekoimi/go-project-template/internal/pkg/logger"
	"github.com/nekoimi/go-project-template/internal/pkg/timeutil"
	"github.com/nekoimi/go-project-template/internal/scheduler"
	"github.com/nekoimi/go-project-template/internal/storage"
	storagefactory "github.com/nekoimi/go-project-template/internal/storage/factory"
	"github.com/nekoimi/go-project-template/internal/taskqueue"
	httptransport "github.com/nekoimi/go-project-template/internal/transport/http"
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
	TaskQueue  *taskqueue.Client
	Worker     *taskqueue.Worker
	HTTPServer *http.Server
	Modules    []framework.Module
	ModuleCtx  *framework.ModuleContext

	runtimeErr chan error
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
	return initialize(configPath, framework.ScopeHTTP)
}

func initialize(configPath string, scopes ...framework.Scope) (*App, func(), error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}
	if err := validateRuntime(cfg, scopes...); err != nil {
		return nil, nil, err
	}
	if err := timeutil.SetGlobalLocation(cfg.Server.Timezone); err != nil {
		return nil, nil, fmt.Errorf("failed to set timezone: %w", err)
	}
	if err := idgen.Init(cfg.Snowflake.NodeID); err != nil {
		return nil, nil, fmt.Errorf("failed to init idgen: %w", err)
	}

	log, err := logger.NewLogger(cfg.Server.Mode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	var cleanupOnce sync.Once
	var db *gorm.DB
	var queueClient *taskqueue.Client
	cleanup := func() {
		cleanupOnce.Do(func() {
			log.Info("cleaning up resources")
			if queueClient != nil {
				if closeErr := queueClient.Close(); closeErr != nil {
					log.Warn("failed to close task queue", zap.Error(closeErr))
				}
			}
			if db != nil {
				if sqlDB, dbErr := db.DB(); dbErr == nil {
					if closeErr := sqlDB.Close(); closeErr != nil {
						log.Warn("failed to close database", zap.Error(closeErr))
					}
				}
			}
			_ = log.Sync()
		})
	}
	fail := func(err error) (*App, func(), error) {
		cleanup()
		return nil, nil, err
	}

	db, err = database.NewPostgresDB(cfg.Database, log, cfg.Server.Mode)
	if err != nil {
		return fail(fmt.Errorf("failed to connect database: %w", err))
	}
	fileStorage, err := storagefactory.New(cfg.Storage)
	if err != nil {
		return fail(fmt.Errorf("failed to create storage: %w", err))
	}

	health := framework.NewHealthRegistry()
	health.Register("database", func(ctx context.Context) error {
		sqlDB, dbErr := db.DB()
		if dbErr != nil {
			return dbErr
		}
		return sqlDB.PingContext(ctx)
	})
	if cfg.TaskQueue.Enabled {
		queueClient = taskqueue.NewClient(cfg.TaskQueue.Redis)
		health.Register("redis", queueClient.Ping)
	}

	var sched *scheduler.Scheduler
	if hasScope(scopes, framework.ScopeScheduler) {
		sched = scheduler.New(cfg.Scheduler, log, db)
	}
	var worker *taskqueue.Worker
	if hasScope(scopes, framework.ScopeWorker) {
		worker = taskqueue.NewWorker(cfg.TaskQueue, log)
	}
	var router *framework.RouterContext
	var wsManager *ws.Manager
	if hasScope(scopes, framework.ScopeHTTP) {
		wsManager = ws.NewManager(log)
		router = httptransport.SetupRouter(cfg, log, health)
	}

	events := framework.NewEventBus()
	moduleCtx := framework.NewModuleContext(
		cfg, log, db, router, sched, fileStorage, wsManager, health, events, queueClient, worker,
	)
	modules := registeredModules(scopes...)
	if err := framework.RegisterModules(moduleCtx, modules...); err != nil {
		return fail(err)
	}

	app := &App{
		Config:     cfg,
		Logger:     log,
		DB:         db,
		Storage:    fileStorage,
		WSManager:  wsManager,
		Scheduler:  sched,
		TaskQueue:  queueClient,
		Worker:     worker,
		Modules:    modules,
		ModuleCtx:  moduleCtx,
		runtimeErr: make(chan error, 2),
	}
	if router != nil {
		app.Engine = router.Engine
	}
	return app, cleanup, nil
}

func hasScope(scopes []framework.Scope, wanted framework.Scope) bool {
	for _, scope := range scopes {
		if scope == wanted {
			return true
		}
	}
	return false
}

func validateRuntime(cfg *config.Config, scopes ...framework.Scope) error {
	if hasScope(scopes, framework.ScopeHTTP) && !cfg.Server.Enabled {
		return fmt.Errorf("http runtime is disabled")
	}
	if hasScope(scopes, framework.ScopeScheduler) {
		if !cfg.Scheduler.Enabled {
			return fmt.Errorf("scheduler runtime is disabled")
		}
		if !cfg.TaskQueue.Enabled {
			return fmt.Errorf("scheduler runtime requires task queue")
		}
	}
	if hasScope(scopes, framework.ScopeWorker) && !cfg.TaskQueue.Enabled {
		return fmt.Errorf("worker runtime requires task queue")
	}
	return nil
}
