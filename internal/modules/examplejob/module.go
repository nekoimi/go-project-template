package examplejob

import (
	"github.com/nekoimi/go-project-template/internal/framework"
)

func init() {
	framework.Register(NewModule(), framework.ScopeScheduler, framework.ScopeWorker)
}

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "example_job"
}

func (m *Module) Register(ctx *framework.ModuleContext) error {
	if !ctx.ModuleEnabled(m.Name()) {
		return nil
	}

	if ctx.Scheduler != nil {
		if ctx.Queue == nil {
			return ErrTaskQueueRequired
		}
		if _, err := ctx.AddCronJob("0 */5 * * * *", NewSchedulerJob(ctx.Queue, ctx.Logger)); err != nil {
			return err
		}
	}
	if ctx.Tasks != nil {
		if err := ctx.Tasks.Handle(TaskTypeExample, NewTaskHandler(ctx.Logger)); err != nil {
			return err
		}
	}
	return nil
}
