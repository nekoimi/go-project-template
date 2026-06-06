package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/nekoimi/go-project-template/internal/app"
	"github.com/nekoimi/go-project-template/internal/scheduler"
)

func main() {
	configPath := flag.String("config", "config/config.dev.yaml", "path to config file")
	flag.Parse()

	// 复用 app 包的初始化逻辑
	a, cleanup, err := app.Initialize(*configPath)
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
	defer cleanup()

	// 使用 app 中已经初始化好的 Scheduler
	sched := a.Scheduler
	if sched == nil {
		sched = scheduler.New(a.Config.Scheduler, a.Logger, a.DB)
		if err := app.RegisterSchedulerModules(a.Config, a.Logger, a.DB, sched); err != nil {
			log.Fatalf("failed to register scheduler modules: %v", err)
		}
	}

	sched.Start()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	a.Logger.Info("shutting down scheduler")
	sched.Stop()
}
