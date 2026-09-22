package app

import (
	"testing"

	"github.com/nekoimi/go-project-template/internal/config"
	"github.com/nekoimi/go-project-template/internal/framework"
)

func TestRuntimeScopesDoNotOverlap(t *testing.T) {
	t.Parallel()

	assertModuleNames(t, registeredModules(framework.ScopeHTTP), []string{"auth", "upload", "user", "websocket"})
	assertModuleNames(t, registeredModules(framework.ScopeScheduler), []string{"example_job"})
	assertModuleNames(t, registeredModules(framework.ScopeWorker), []string{"example_job"})
	assertModuleNames(t, registeredModules(
		framework.ScopeHTTP,
		framework.ScopeScheduler,
		framework.ScopeWorker,
	), []string{"auth", "example_job", "upload", "user", "websocket"})
}

func TestValidateRuntimeRequiresTaskQueue(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultConfig()
	cfg.TaskQueue.Enabled = false

	if err := validateRuntime(cfg, framework.ScopeHTTP); err != nil {
		t.Fatalf("HTTP runtime validation: %v", err)
	}
	if err := validateRuntime(cfg, framework.ScopeScheduler); err == nil {
		t.Fatal("scheduler runtime accepted disabled task queue")
	}
	if err := validateRuntime(cfg, framework.ScopeWorker); err == nil {
		t.Fatal("worker runtime accepted disabled task queue")
	}
	if err := validateRuntime(cfg, framework.ScopeHTTP, framework.ScopeScheduler, framework.ScopeWorker); err == nil {
		t.Fatal("all runtime accepted disabled task queue")
	}
}

func TestRuntimeTaskQueueRequirement(t *testing.T) {
	t.Parallel()

	if requiresTaskQueue(framework.ScopeHTTP) {
		t.Fatal("HTTP-only runtime unexpectedly requires task queue resources")
	}
	if !requiresTaskQueue(framework.ScopeScheduler) {
		t.Fatal("scheduler runtime does not require task queue resources")
	}
	if !requiresTaskQueue(framework.ScopeWorker) {
		t.Fatal("worker runtime does not require task queue resources")
	}
	if !requiresTaskQueue(framework.ScopeHTTP, framework.ScopeScheduler, framework.ScopeWorker) {
		t.Fatal("all runtime does not require task queue resources")
	}
}

func assertModuleNames(t *testing.T, modules []framework.Module, want []string) {
	t.Helper()
	if len(modules) != len(want) {
		t.Fatalf("module count = %d, want %d", len(modules), len(want))
	}
	for i := range want {
		if modules[i].Name() != want[i] {
			t.Fatalf("module %d = %q, want %q", i, modules[i].Name(), want[i])
		}
	}
}
