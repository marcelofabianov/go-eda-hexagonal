package user

import (
	"context"

	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

// --- UserExistsRepository ---
type UserExistsRepoInput struct {
	Email types.Email
	Phone types.Phone
}

type UserExistsRepository interface {
	UserExists(ctx context.Context, input UserExistsRepoInput) (bool, error)
}

// --- CreateUserRepository ---

type CreateUserRepoInput struct {
	User *User
}

type CreateUserRepository interface {
	UserExistsRepository
	CreateUser(ctx context.Context, input CreateUserRepoInput) error
}

// --- UserRepository ---
type UserRepository interface {
	CreateUserRepository
}
