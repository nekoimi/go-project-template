package examplejob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/taskqueue"
)

const TaskTypeExample = "example:run"

var ErrTaskQueueRequired = errors.New("example job requires task queue")

type TaskPayload struct {
	TriggeredAt time.Time `json:"triggered_at"`
}

type SchedulerJob struct {
	queue  taskqueue.Enqueuer
	logger *zap.Logger
	now    func() time.Time
}

func NewSchedulerJob(queue taskqueue.Enqueuer, logger *zap.Logger) *SchedulerJob {
	return &SchedulerJob{queue: queue, logger: logger, now: time.Now}
}

func (j *SchedulerJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	triggeredAt := j.now().UTC().Truncate(5 * time.Minute)
	payload, err := json.Marshal(TaskPayload{TriggeredAt: triggeredAt})
	if err != nil {
		j.logger.Error("failed to marshal example task", zap.Error(err))
		return
	}
	info, err := j.queue.Enqueue(ctx, taskqueue.Task{Type: TaskTypeExample, Payload: payload}, taskqueue.EnqueueOptions{
		Queue:     "default",
		MaxRetry:  5,
		Timeout:   time.Minute,
		UniqueFor: 4 * time.Minute,
	})
	if errors.Is(err, taskqueue.ErrDuplicateTask) {
		j.logger.Debug("example task already enqueued", zap.Time("triggered_at", triggeredAt))
		return
	}
	if err != nil {
		j.logger.Error("failed to enqueue example task", zap.Error(err))
		return
	}
	j.logger.Info("example task enqueued", zap.String("task_id", info.ID), zap.String("queue", info.Queue))
}

func NewTaskHandler(logger *zap.Logger) taskqueue.Handler {
	return func(ctx context.Context, payload []byte) error {
		var taskPayload TaskPayload
		if err := json.Unmarshal(payload, &taskPayload); err != nil {
			return fmt.Errorf("decode example task payload: %w", err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		logger.Info("example task executed", zap.Time("triggered_at", taskPayload.TriggeredAt))
		return nil
	}
}
