package publisher

import (
	"context"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/port/bus"
)

type UserPublisher struct {
	bus bus.EventBusPublisher
}

func NewUserPublisher(bus bus.EventBusPublisher) *UserPublisher {
	return &UserPublisher{
		bus: bus,
	}
}

func (u *UserPublisher) PublishUserCreatedEvent(ctx context.Context, input user.USerCreatedEventInput) error {
	event, err := user.NewUserCreatedEvent(input)
	if err != nil {
		return err
	}

	return u.bus.Publish(ctx, event)
}

// TODO: Add other user events
