package framework

import (
	"context"
	"testing"
)

type lifecycleModule struct {
	name      string
	booted    bool
	shutdown  bool
	callOrder *[]string
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
	return nil
}

func (m *lifecycleModule) Shutdown(ctx context.Context, moduleCtx *ModuleContext) error {
	m.shutdown = true
	*m.callOrder = append(*m.callOrder, "shutdown:"+m.name)
	return nil
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
	if len(calls) != len(want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls = %#v, want %#v", calls, want)
		}
	}
	if !first.booted || !first.shutdown || !second.booted || !second.shutdown {
		t.Fatalf("lifecycle flags not set: first=%+v second=%+v", first, second)
	}
}
