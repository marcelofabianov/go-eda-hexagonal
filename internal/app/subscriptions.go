package app

import (
	"fmt"

	"go.uber.org/dig"

	auditSubscriber "github.com/marcelofabianov/redtogreen/internal/contexts/audit/app/subscriber"
	identityDomain "github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	platformBus "github.com/marcelofabianov/redtogreen/internal/platform/port/bus"
)

func setupEventSubscriptions(container *dig.Container) error {
	return container.Invoke(func(
		busSubscriber platformBus.EventBusSubscriber,
		userCreatedSubscriber *auditSubscriber.UserCreatedSubscriber,
	) error {
		if err := subscribeToUserEvents(busSubscriber, userCreatedSubscriber); err != nil {
			return err
		}

		return nil
	})
}

func subscribeToUserEvents(
	busSubscriber platformBus.EventBusSubscriber,
	userCreatedSubscriber *auditSubscriber.UserCreatedSubscriber,
) error {
	if err := busSubscriber.Subscribe(identityDomain.UserCreatedEventType, userCreatedSubscriber.Handle); err != nil {
		return fmt.Errorf("failed to subscribe to %s event: %w", identityDomain.UserCreatedEventType, err)
	}

	return nil
}
