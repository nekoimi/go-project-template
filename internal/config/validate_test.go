package config

import (
	"strings"
	"testing"
)

func TestValidateRejectsReleaseSecretPlaceholder(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Server.Mode = "release"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "jwt secret") {
		t.Fatalf("Validate error = %v, want jwt secret error", err)
	}
}

func TestValidateTaskQueue(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.TaskQueue.Enabled = true
	cfg.TaskQueue.Redis.Addr = ""

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "redis address") {
		t.Fatalf("Validate error = %v, want redis address error", err)
	}
}

func TestValidateForHTTPDoesNotRequireTaskQueue(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.TaskQueue.Enabled = true
	cfg.TaskQueue.Redis.Addr = ""

	err := cfg.ValidateFor(ValidationRequirements{HTTP: true, Database: true, Storage: true})
	if err != nil {
		t.Fatalf("HTTP-only validation unexpectedly required task queue: %v", err)
	}
}

func TestValidateForWorkerDoesNotRequireHTTPOrStorage(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Server.Enabled = false
	cfg.Server.Mode = "release"
	cfg.JWT.Secret = ""
	cfg.Storage.Driver = "s3"
	cfg.Storage.S3.AccessKey = ""
	cfg.TaskQueue.Enabled = true

	err := cfg.ValidateFor(ValidationRequirements{Worker: true})
	if err != nil {
		t.Fatalf("worker validation unexpectedly required HTTP or storage: %v", err)
	}
}
