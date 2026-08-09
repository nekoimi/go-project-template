package framework

import (
	"context"
	"errors"
	"testing"
)

type lifecycleModule struct {
	name      string
	booted    bool
	shutdown  bool
	callOrder *[]string
	bootErr   error
	stopErr   error
}

func (m *lifecycleModule) Name() string {
	return m.name
}

func (m *lifecycleModule) Register(ctx *ModuleContext) error {
	return nil
}

func (m *lifecycleModule) Boot(ctx context.Context, moduleCtx *ModuleContext) error {
	m.booted = true
	*m.callOrder = append(*m.callOrder, "boot:"+m.name)
	return m.bootErr
}

func (m *lifecycleModule) Shutdown(ctx context.Context, moduleCtx *ModuleContext) error {
	m.shutdown = true
	*m.callOrder = append(*m.callOrder, "shutdown:"+m.name)
	return m.stopErr
}

func TestBootModulesRollsBackPreviouslyBootedModules(t *testing.T) {
	t.Parallel()

	bootErr := errors.New("boot failed")
	var calls []string
	first := &lifecycleModule{name: "first", callOrder: &calls}
	second := &lifecycleModule{name: "second", callOrder: &calls, bootErr: bootErr}

	err := BootModules(context.Background(), nil, first, second)
	if !errors.Is(err, bootErr) {
		t.Fatalf("BootModules error = %v, want %v", err, bootErr)
	}
	want := []string{"boot:first", "boot:second", "shutdown:first"}
	assertCallOrder(t, calls, want)
}

func TestShutdownModulesContinuesAfterErrors(t *testing.T) {
	t.Parallel()

	firstErr := errors.New("first stop failed")
	secondErr := errors.New("second stop failed")
	var calls []string
	first := &lifecycleModule{name: "first", callOrder: &calls, stopErr: firstErr}
	second := &lifecycleModule{name: "second", callOrder: &calls, stopErr: secondErr}

	err := ShutdownModules(context.Background(), nil, first, second)
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("ShutdownModules error = %v, want both shutdown errors", err)
	}
	assertCallOrder(t, calls, []string{"shutdown:second", "shutdown:first"})
}

func assertCallOrder(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("calls = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("calls = %#v, want %#v", got, want)
		}
	}
}

func TestLifecycleModules(t *testing.T) {
	t.Parallel()

	var calls []string
	first := &lifecycleModule{name: "first", callOrder: &calls}
	second := &lifecycleModule{name: "second", callOrder: &calls}

	if err := BootModules(context.Background(), nil, first, second); err != nil {
		t.Fatal(err)
	}
	if err := ShutdownModules(context.Background(), nil, first, second); err != nil {
		t.Fatal(err)
	}

	want := []string{"boot:first", "boot:second", "shutdown:second", "shutdown:first"}
	assertCallOrder(t, calls, want)
	if !first.booted || !first.shutdown || !second.booted || !second.shutdown {
		t.Fatalf("lifecycle flags not set: first=%+v second=%+v", first, second)
	}
}
