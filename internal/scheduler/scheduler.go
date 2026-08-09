package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nekoimi/go-project-template/internal/config"
)

type Scheduler struct {
	cron   *cron.Cron
	logger *zap.Logger
	db     *gorm.DB
}

func New(cfg config.SchedulerConfig, logger *zap.Logger, db *gorm.DB) *Scheduler {
	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		logger.Warn("invalid timezone, using UTC", zap.String("timezone", cfg.Timezone), zap.Error(err))
		location = time.UTC
	}

	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(location),
		cron.WithChain(
			cron.Recover(zapCronLogger{logger: logger.Sugar()}),
			cron.SkipIfStillRunning(zapCronLogger{logger: logger.Sugar()}),
		),
	)

	return &Scheduler{
		cron:   c,
		logger: logger,
		db:     db,
	}
}

func (s *Scheduler) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	return s.cron.AddJob(spec, cmd)
}

func (s *Scheduler) Start() {
	s.logger.Info("scheduler started")
	s.cron.Start()
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.logger.Info("scheduler stopping")
	jobsDone := s.cron.Stop()
	select {
	case <-jobsDone.Done():
		s.logger.Info("scheduler stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type zapCronLogger struct {
	logger *zap.SugaredLogger
}

func (l zapCronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infow(msg, keysAndValues...)
}

func (l zapCronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	fields := append(keysAndValues, "error", fmt.Sprint(err))
	l.logger.Errorw(msg, fields...)
}
