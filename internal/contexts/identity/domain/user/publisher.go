package user

import (
	"context"
)

type UserCreatedEventPublisher interface {
	PublishUserCreatedEvent(ctx context.Context, input USerCreatedEventInput) error
}

type UserPublisher interface {
	UserCreatedEventPublisher
}
