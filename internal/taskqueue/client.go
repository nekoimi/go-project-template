package taskqueue

import (
	"context"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/nekoimi/go-project-template/internal/config"
)

type Client struct {
	asynq *asynq.Client
	redis *redis.Client
}

func NewClient(cfg config.RedisConfig) *Client {
	return &Client{
		asynq: asynq.NewClient(redisOptions(cfg)),
		redis: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}
}

func (c *Client) Enqueue(ctx context.Context, task Task, options EnqueueOptions) (*TaskInfo, error) {
	if c == nil || c.asynq == nil {
		return nil, fmt.Errorf("task queue client is not initialized")
	}
	if task.Type == "" {
		return nil, fmt.Errorf("task type is required")
	}

	opts := make([]asynq.Option, 0, 4)
	if options.Queue != "" {
		opts = append(opts, asynq.Queue(options.Queue))
	}
	if options.MaxRetry >= 0 {
		opts = append(opts, asynq.MaxRetry(options.MaxRetry))
	}
	if options.Timeout > 0 {
		opts = append(opts, asynq.Timeout(options.Timeout))
	}
	if options.UniqueFor > 0 {
		opts = append(opts, asynq.Unique(options.UniqueFor))
	}

	info, err := c.asynq.EnqueueContext(ctx, asynq.NewTask(task.Type, task.Payload), opts...)
	if err != nil {
		if errors.Is(err, asynq.ErrDuplicateTask) {
			return nil, ErrDuplicateTask
		}
		return nil, fmt.Errorf("enqueue task %s: %w", task.Type, err)
	}
	return &TaskInfo{ID: info.ID, Queue: info.Queue}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.redis == nil {
		return fmt.Errorf("task queue client is not initialized")
	}
	if err := c.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	var errs []error
	if c.asynq != nil {
		errs = append(errs, c.asynq.Close())
	}
	if c.redis != nil {
		errs = append(errs, c.redis.Close())
	}
	return errors.Join(errs...)
}

func redisOptions(cfg config.RedisConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}
