package taskqueue

import (
	"context"
	"errors"
	"time"
)

var ErrDuplicateTask = errors.New("task already exists")

type Task struct {
	Type    string
	Payload []byte
}

type EnqueueOptions struct {
	Queue     string
	MaxRetry  int
	Timeout   time.Duration
	UniqueFor time.Duration
}

type TaskInfo struct {
	ID    string
	Queue string
}

type Enqueuer interface {
	Enqueue(ctx context.Context, task Task, options EnqueueOptions) (*TaskInfo, error)
}

type Handler func(ctx context.Context, payload []byte) error

type HandlerRegistrar interface {
	Handle(taskType string, handler Handler) error
}
