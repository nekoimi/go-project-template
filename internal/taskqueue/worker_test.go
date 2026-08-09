package taskqueue

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/config"
)

func TestWorkerRejectsDuplicateHandler(t *testing.T) {
	t.Parallel()
	worker := NewWorker(config.DefaultConfig().TaskQueue, zap.NewNop())
	handler := func(context.Context, []byte) error { return nil }

	if err := worker.Handle("example:task", handler); err != nil {
		t.Fatal(err)
	}
	if err := worker.Handle("example:task", handler); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("duplicate Handle error = %v", err)
	}
}
