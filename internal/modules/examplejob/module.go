package examplejob

import (
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/module"
	"github.com/nekoimi/go-project-template/internal/scheduler/jobs"
)

func init() {
	module.Register(NewModule(), module.ScopeHTTP, module.ScopeScheduler)
}

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "example_job"
}

func (m *Module) Register(ctx *framework.ModuleContext) error {
	if !ctx.ModuleEnabled(m.Name()) || ctx.Scheduler == nil {
		return nil
	}

	if _, err := ctx.AddCronJob("0 */5 * * * *", jobs.NewExampleJob(ctx.Logger)); err != nil {
		ctx.Logger.Error("failed to register example job", zap.Error(err))
		return err
	}
	return nil
}
