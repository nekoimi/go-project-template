package examplejob

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/taskqueue"
)

type captureQueue struct {
	task    taskqueue.Task
	options taskqueue.EnqueueOptions
}

func (q *captureQueue) Enqueue(_ context.Context, task taskqueue.Task, options taskqueue.EnqueueOptions) (*taskqueue.TaskInfo, error) {
	q.task = task
	q.options = options
	return &taskqueue.TaskInfo{ID: "task-1", Queue: options.Queue}, nil
}

func TestSchedulerJobEnqueuesIdempotentTask(t *testing.T) {
	t.Parallel()
	queue := &captureQueue{}
	job := NewSchedulerJob(queue, zap.NewNop())
	job.now = func() time.Time {
		return time.Date(2026, 8, 9, 10, 7, 31, 0, time.FixedZone("CST", 8*60*60))
	}

	job.Run()

	if queue.task.Type != TaskTypeExample {
		t.Fatalf("task type = %q, want %q", queue.task.Type, TaskTypeExample)
	}
	if queue.options.Queue != "default" || queue.options.MaxRetry != 5 || queue.options.UniqueFor != 4*time.Minute {
		t.Fatalf("enqueue options = %+v", queue.options)
	}
	var payload TaskPayload
	if err := json.Unmarshal(queue.task.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 8, 9, 2, 5, 0, 0, time.UTC)
	if !payload.TriggeredAt.Equal(want) {
		t.Fatalf("triggered at = %s, want %s", payload.TriggeredAt, want)
	}
}

func TestTaskHandlerRejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	handler := NewTaskHandler(zap.NewNop())
	if err := handler(context.Background(), []byte("not-json")); err == nil {
		t.Fatal("handler accepted invalid payload")
	}
}
