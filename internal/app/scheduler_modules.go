package app

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/scheduler"
)

func RegisterSchedulerModules(cfg *config.Config, logger *zap.Logger, db *gorm.DB, sched *scheduler.Scheduler) error {
	moduleCtx := framework.NewModuleContext(cfg, logger, db, nil, sched, nil, nil, nil, framework.NewEventBus())
	return framework.RegisterModules(moduleCtx, builtinSchedulerModules()...)
}
