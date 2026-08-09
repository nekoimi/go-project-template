package config

import (
	"fmt"
	"strings"
	"time"
)

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if isPlaceholder(c.Database.Host) || isPlaceholder(c.Database.Port) || isPlaceholder(c.Database.User) || isPlaceholder(c.Database.DBName) {
		return fmt.Errorf("database host, port, user and dbname are required")
	}
	if c.Server.Enabled && c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" && c.Server.Mode != "test" {
		return fmt.Errorf("unsupported server mode %q", c.Server.Mode)
	}
	if c.Server.Mode == "release" && isPlaceholder(c.JWT.Secret) {
		return fmt.Errorf("jwt secret must be set in release mode")
	}
	if c.Scheduler.Enabled {
		if _, err := time.LoadLocation(c.Scheduler.Timezone); err != nil {
			return fmt.Errorf("invalid scheduler timezone %q: %w", c.Scheduler.Timezone, err)
		}
	}
	if c.TaskQueue.Enabled {
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
	}
	if c.Storage.Driver == "s3" && c.Storage.S3.Provider != "aws" {
		if isPlaceholder(c.Storage.S3.Endpoint) || isPlaceholder(c.Storage.S3.AccessKey) ||
			isPlaceholder(c.Storage.S3.SecretKey) || isPlaceholder(c.Storage.S3.Bucket) {
			return fmt.Errorf("s3 endpoint, credentials and bucket are required")
		}
	}
	return nil
}

func isPlaceholder(value string) bool {
	value = strings.TrimSpace(value)
	return value == "" || value == "change-me-in-production" || strings.Contains(value, "${")
}
