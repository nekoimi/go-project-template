package framework

import (
	"context"
	"errors"
	"testing"
)

func TestHealthRegistryCheckReady(t *testing.T) {
	t.Parallel()

	registry := NewHealthRegistry()
	registry.Register("database", func(ctx context.Context) error {
		return nil
	})

	result := registry.Check(context.Background())
	if result.Status != "ready" {
		t.Fatalf("Status = %q, want ready", result.Status)
	}
	if result.Checks["database"] != "ok" {
		t.Fatalf("database check = %q, want ok", result.Checks["database"])
	}
}

func TestHealthRegistryCheckNotReady(t *testing.T) {
	t.Parallel()

	registry := NewHealthRegistry()
	registry.Register("database", func(ctx context.Context) error {
		return errors.New("connection refused")
	})

	result := registry.Check(context.Background())
	if result.Status != "not ready" {
		t.Fatalf("Status = %q, want not ready", result.Status)
	}
	if result.Checks["database"] != "connection refused" {
		t.Fatalf("database check = %q, want connection refused", result.Checks["database"])
	}
}
