package taskqueue

import (
	"context"
	"fmt"
	"sync"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/config"
)

type Worker struct {
	server   *asynq.Server
	mux      *asynq.ServeMux
	mu       sync.Mutex
	handlers map[string]struct{}
}

func NewWorker(cfg config.TaskQueueConfig, logger *zap.Logger) *Worker {
	server := asynq.NewServer(redisOptions(cfg.Redis), asynq.Config{
		Concurrency:     cfg.Concurrency,
		Queues:          cfg.Queues,
		ShutdownTimeout: cfg.ShutdownTimeout,
		Logger:          asynqLogger{logger: logger},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retryCount, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			logger.Error("task execution failed",
				zap.String("task_type", task.Type()),
				zap.Int("retry_count", retryCount),
				zap.Int("max_retry", maxRetry),
				zap.Error(err),
			)
		}),
		HealthCheckFunc: func(err error) {
			if err != nil {
				logger.Error("task queue health check failed", zap.Error(err))
			}
		},
	})

	return &Worker{
		server:   server,
		mux:      asynq.NewServeMux(),
		handlers: make(map[string]struct{}),
	}
}

func (w *Worker) Handle(taskType string, handler Handler) error {
	if w == nil || w.mux == nil {
		return fmt.Errorf("task worker is not initialized")
	}
	if taskType == "" || handler == nil {
		return fmt.Errorf("task type and handler are required")
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if _, exists := w.handlers[taskType]; exists {
		return fmt.Errorf("task handler %q is already registered", taskType)
	}
	w.handlers[taskType] = struct{}{}
	w.mux.HandleFunc(taskType, func(ctx context.Context, task *asynq.Task) error {
		return handler(ctx, task.Payload())
	})
	return nil
}

func (w *Worker) Start() error {
	if w == nil || w.server == nil || w.mux == nil {
		return fmt.Errorf("task worker is not initialized")
	}
	return w.server.Start(w.mux)
}

func (w *Worker) Shutdown() {
	if w == nil || w.server == nil {
		return
	}
	w.server.Shutdown()
}

type asynqLogger struct {
	logger *zap.Logger
}

func (l asynqLogger) Debug(args ...interface{}) { l.logger.Debug(fmt.Sprint(args...)) }
func (l asynqLogger) Info(args ...interface{})  { l.logger.Info(fmt.Sprint(args...)) }
func (l asynqLogger) Warn(args ...interface{})  { l.logger.Warn(fmt.Sprint(args...)) }
func (l asynqLogger) Error(args ...interface{}) { l.logger.Error(fmt.Sprint(args...)) }
func (l asynqLogger) Fatal(args ...interface{}) { l.logger.Error(fmt.Sprint(args...)) }
