package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/infrastructure/database"
	"github.com/nekoimi/go-project-template/internal/infrastructure/logger"
	"github.com/nekoimi/go-project-template/internal/scheduler"
)

func main() {
	configPath := flag.String("config", "configs/config.dev.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLogger, err := logger.NewLogger(cfg.Server.Mode)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.Database, zapLogger)
	if err != nil {
		zapLogger.Fatal("failed to connect database", zap.Error(err))
	}

	sched := scheduler.New(cfg.Scheduler, zapLogger, db)
	sched.RegisterJobs()

	sched.Start()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	zapLogger.Info("shutting down scheduler")
	sched.Stop()

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
	zapLogger.Sync()
}
