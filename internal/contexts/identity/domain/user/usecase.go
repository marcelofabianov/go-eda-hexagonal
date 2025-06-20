package user

import "context"

// --- CreateUserUseCase ---
type CreateUserOutput struct {
	User *User
}

type CreateUserUseCase interface {
	Execute(ctx context.Context, input NewUserInput) (CreateUserOutput, error)
}
