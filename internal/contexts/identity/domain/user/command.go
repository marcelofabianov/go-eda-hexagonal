package user

import (
	"context"

	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

// --- CreateUserCommand ---

type CreateUserCommandInput struct {
	CorrelationID   types.UUID
	TraceID         types.UUID
	UserAuthorID    types.NullableUUID
	PreviousEventID types.NullableUUID
	CausationID     types.NullableUUID
	NewUserInput    NewUserInput
}

type CreateUserCommand interface {
	Execute(ctx context.Context, input CreateUserCommandInput) (CreateUserOutput, error)
}
