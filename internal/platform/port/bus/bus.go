package bus

import (
	"context"

	"github.com/marcelofabianov/redtogreen/internal/platform/event"
)

type EventHandler func(ctx context.Context, event *event.Event) error

type EventBusPublisher interface {
	Publish(ctx context.Context, event *event.Event) error
}

type EventBusSubscriber interface {
	Subscribe(eventType event.EventType, handler EventHandler) error
}

type EventBus interface {
	EventBusPublisher
	EventBusSubscriber
}
