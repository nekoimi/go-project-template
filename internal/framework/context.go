package framework

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/scheduler"
	"github.com/nekoimi/go-project-template/internal/storage"
	ws "github.com/nekoimi/go-project-template/internal/websocket"
)

type RouterContext struct {
	Engine    *gin.Engine
	API       *gin.RouterGroup
	Protected *gin.RouterGroup
}

type ModuleContext struct {
	Config    *config.Config
	Logger    *zap.Logger
	DB        *gorm.DB
	Router    *gin.Engine
	API       *gin.RouterGroup
	Protected *gin.RouterGroup
	Scheduler *scheduler.Scheduler
	Storage   storage.FileStorage
	WSManager *ws.Manager
	Health    *HealthRegistry
	Events    *EventBus
}

func NewModuleContext(
	cfg *config.Config,
	logger *zap.Logger,
	db *gorm.DB,
	router *RouterContext,
	sched *scheduler.Scheduler,
	fileStorage storage.FileStorage,
	wsManager *ws.Manager,
	health *HealthRegistry,
	events *EventBus,
) *ModuleContext {
	var engine *gin.Engine
	var api *gin.RouterGroup
	var protected *gin.RouterGroup
	if router != nil {
		engine = router.Engine
		api = router.API
		protected = router.Protected
	}

	return &ModuleContext{
		Config:    cfg,
		Logger:    logger,
		DB:        db,
		Router:    engine,
		API:       api,
		Protected: protected,
		Scheduler: sched,
		Storage:   fileStorage,
		WSManager: wsManager,
		Health:    health,
		Events:    events,
	}
}

func (ctx *ModuleContext) ModuleEnabled(name string) bool {
	if ctx == nil || ctx.Config == nil {
		return false
	}
	return ctx.Config.ModuleEnabled(name)
}

func (ctx *ModuleContext) AddCronJob(spec string, job cron.Job) (cron.EntryID, error) {
	if ctx == nil || ctx.Scheduler == nil {
		return 0, nil
	}
	return ctx.Scheduler.AddJob(spec, job)
}

func (ctx *ModuleContext) AddHealthCheck(name string, check HealthCheck) {
	if ctx == nil || ctx.Health == nil {
		return
	}
	ctx.Health.Register(name, check)
}

func (ctx *ModuleContext) Subscribe(topic string, handler EventHandler) {
	if ctx == nil || ctx.Events == nil {
		return
	}
	ctx.Events.Subscribe(topic, handler)
}

func (ctx *ModuleContext) Publish(publishCtx context.Context, event Event) error {
	if ctx == nil || ctx.Events == nil {
		return nil
	}
	return ctx.Events.Publish(publishCtx, event)
}
