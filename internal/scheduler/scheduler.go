package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/config"
)

type Scheduler struct {
	cron    *cron.Cron
	logger  *zap.Logger
	cfg     config.SchedulerConfig
	elector LeaderElector
	owner   string
	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	done    chan struct{}
}

type LeaderElector interface {
	AcquireLeadership(ctx context.Context, key, owner string, ttl time.Duration) (bool, error)
	RenewLeadership(ctx context.Context, key, owner string, ttl time.Duration) (bool, error)
	ReleaseLeadership(ctx context.Context, key, owner string) error
}

func New(cfg config.SchedulerConfig, logger *zap.Logger, elector LeaderElector) *Scheduler {
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
		cron:    c,
		logger:  logger,
		cfg:     cfg,
		elector: elector,
		owner:   newOwnerID(),
	}
}

func (s *Scheduler) AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	return s.cron.AddJob(spec, cmd)
}

func (s *Scheduler) Start(ctx context.Context) error {
	if !s.cfg.LeaderElection {
		s.startCron()
		return nil
	}
	if s.elector == nil {
		return fmt.Errorf("scheduler leader election requires an elector")
	}
	electionCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.done = make(chan struct{})
	go s.runElection(electionCtx)
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.logger.Info("scheduler stopping")
	if s.cancel != nil {
		s.cancel()
		select {
		case <-s.done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return s.stopCron(ctx)
}

func (s *Scheduler) runElection(ctx context.Context) {
	defer close(s.done)
	ticker := time.NewTicker(s.cfg.LeaderTTL / 3)
	defer ticker.Stop()
	leader := false
	for {
		if !leader {
			acquired, err := s.elector.AcquireLeadership(ctx, s.cfg.LeaderKey, s.owner, s.cfg.LeaderTTL)
			if err != nil && ctx.Err() == nil {
				s.logger.Warn("scheduler leader election failed", zap.Error(err))
			}
			if acquired {
				leader = true
				s.startCron()
			}
		}
		select {
		case <-ctx.Done():
			if leader {
				s.releaseLeadership()
			}
			return
		case <-ticker.C:
			if leader {
				renewed, err := s.elector.RenewLeadership(ctx, s.cfg.LeaderKey, s.owner, s.cfg.LeaderTTL)
				if err != nil || !renewed {
					s.logger.Warn("scheduler leadership lost", zap.Error(err))
					leader = false
					_ = s.stopCron(context.Background())
				}
			}
		}
	}
}

func (s *Scheduler) startCron() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.cron.Start()
	s.running = true
	s.logger.Info("scheduler started", zap.Bool("leader_election", s.cfg.LeaderElection))
}

func (s *Scheduler) stopCron(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	jobsDone := s.cron.Stop()
	s.running = false
	s.mu.Unlock()
	select {
	case <-jobsDone.Done():
		s.logger.Info("scheduler stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) releaseLeadership() {
	releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.elector.ReleaseLeadership(releaseCtx, s.cfg.LeaderKey, s.owner); err != nil {
		s.logger.Warn("release scheduler leadership", zap.Error(err))
	}
}

func newOwnerID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("scheduler-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(value[:])
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
