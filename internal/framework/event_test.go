package framework

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEventBusPublish(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	var topics []string
	bus.Subscribe("user.registered", func(ctx context.Context, event Event) error {
		topics = append(topics, event.Topic)
		return nil
	})

	if err := bus.Publish(context.Background(), Event{Topic: "user.registered"}); err != nil {
		t.Fatal(err)
	}
	if len(topics) != 1 || topics[0] != "user.registered" {
		t.Fatalf("topics = %#v, want user.registered", topics)
	}
}

func TestEventBusPublishJoinsHandlerErrors(t *testing.T) {
	t.Parallel()

	bus := NewEventBus()
	bus.Subscribe("user.registered", func(ctx context.Context, event Event) error {
		return errors.New("first")
	})
	bus.Subscribe("user.registered", func(ctx context.Context, event Event) error {
		return errors.New("second")
	})

	err := bus.Publish(context.Background(), Event{Topic: "user.registered"})
	if err == nil {
		t.Fatal("err = nil, want joined errors")
	}
	if !strings.Contains(err.Error(), "first") || !strings.Contains(err.Error(), "second") {
		t.Fatalf("err = %q, want both handler errors", err.Error())
	}
}
