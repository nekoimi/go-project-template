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
