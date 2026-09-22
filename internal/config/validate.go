package config

import (
	"fmt"
	"strings"
	"time"
)

type ValidationRequirements struct {
	HTTP      bool
	Scheduler bool
	Worker    bool
	Database  bool
	Storage   bool
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if err := c.ValidateFor(ValidationRequirements{HTTP: c.Server.Enabled, Database: true, Storage: true}); err != nil {
		return err
	}
	if c.Scheduler.Enabled {
		if err := c.validateScheduler(); err != nil {
			return err
		}
	}
	if c.TaskQueue.Enabled {
		return c.validateTaskQueue()
	}
	return nil
}

func (c *Config) ValidateFor(requirements ValidationRequirements) error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if requirements.Database && (isPlaceholder(c.Database.Host) || isPlaceholder(c.Database.Port) || isPlaceholder(c.Database.User) || isPlaceholder(c.Database.DBName)) {
		return fmt.Errorf("database host, port, user and dbname are required")
	}
	if requirements.HTTP && !c.Server.Enabled {
		return fmt.Errorf("http runtime is disabled")
	}
	if requirements.HTTP && c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if requirements.HTTP && (c.Server.ReadTimeout <= 0 || c.Server.ReadHeaderTimeout <= 0 || c.Server.WriteTimeout <= 0 || c.Server.IdleTimeout <= 0) {
		return fmt.Errorf("server timeouts must be positive")
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" && c.Server.Mode != "test" {
		return fmt.Errorf("unsupported server mode %q", c.Server.Mode)
	}
	if requirements.HTTP && c.Server.Mode == "release" && isPlaceholder(c.JWT.Secret) {
		return fmt.Errorf("jwt secret must be set in release mode")
	}
	if requirements.Scheduler {
		if !c.Scheduler.Enabled {
			return fmt.Errorf("scheduler runtime is disabled")
		}
		if err := c.validateScheduler(); err != nil {
			return err
		}
	}
	if requirements.Scheduler || requirements.Worker {
		if !c.TaskQueue.Enabled {
			return fmt.Errorf("task queue runtime is disabled")
		}
		if err := c.validateTaskQueue(); err != nil {
			return err
		}
	}
	if requirements.Storage && c.Storage.Driver == "s3" && c.Storage.S3.Provider != "aws" {
		if isPlaceholder(c.Storage.S3.Endpoint) || isPlaceholder(c.Storage.S3.AccessKey) ||
			isPlaceholder(c.Storage.S3.SecretKey) || isPlaceholder(c.Storage.S3.Bucket) {
			return fmt.Errorf("s3 endpoint, credentials and bucket are required")
		}
	}
	if requirements.Storage {
		if c.Storage.Upload.MaxFileSize <= 0 || c.Storage.Upload.MaxRequestSize <= 0 || c.Storage.Upload.MaxFiles <= 0 {
			return fmt.Errorf("upload limits must be positive")
		}
		if c.Storage.Upload.MaxRequestSize < c.Storage.Upload.MaxFileSize {
			return fmt.Errorf("upload max request size must be at least the max file size")
		}
		if c.Storage.Driver == "s3" && c.Storage.S3.StartupTimeout <= 0 {
			return fmt.Errorf("s3 startup timeout must be positive")
		}
	}
	return nil
}

func (c *Config) validateScheduler() error {
	if _, err := time.LoadLocation(c.Scheduler.Timezone); err != nil {
		return fmt.Errorf("invalid scheduler timezone %q: %w", c.Scheduler.Timezone, err)
	}
	if c.Scheduler.LeaderElection {
		if strings.TrimSpace(c.Scheduler.LeaderKey) == "" {
			return fmt.Errorf("scheduler leader key is required")
		}
		if c.Scheduler.LeaderTTL < 3*time.Second {
			return fmt.Errorf("scheduler leader TTL must be at least 3s")
		}
	}
	return nil
}

func (c *Config) validateTaskQueue() error {
	if c.TaskQueue.Redis.Addr == "" || isPlaceholder(c.TaskQueue.Redis.Addr) {
		return fmt.Errorf("task queue redis address is required")
	}
	if c.TaskQueue.Concurrency <= 0 {
		return fmt.Errorf("task queue concurrency must be positive")
	}
	if c.TaskQueue.ShutdownTimeout <= 0 {
		return fmt.Errorf("task queue shutdown timeout must be positive")
	}
	if len(c.TaskQueue.Queues) == 0 {
		return fmt.Errorf("at least one task queue is required")
	}
	for name, priority := range c.TaskQueue.Queues {
		if strings.TrimSpace(name) == "" || priority <= 0 {
			return fmt.Errorf("task queue names must be non-empty and priorities must be positive")
		}
	}
	return nil
}

func isPlaceholder(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == "change-me-in-production" || strings.Contains(value, "${")
}
