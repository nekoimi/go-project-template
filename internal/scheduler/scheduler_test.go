package scheduler

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/config"
)

func TestStartRequiresElectorWhenLeaderElectionEnabled(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig().Scheduler
	s := New(cfg, zap.NewNop(), nil)
	if err := s.Start(context.Background()); err == nil {
		t.Fatal("expected missing elector error")
	}
}

func TestStartWithoutLeaderElection(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig().Scheduler
	cfg.LeaderElection = false
	s := New(cfg, zap.NewNop(), nil)
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}
