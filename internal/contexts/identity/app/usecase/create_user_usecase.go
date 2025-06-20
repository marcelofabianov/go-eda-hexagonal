package usecase

import (
	"context"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/port/hasher"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

type createUserUseCase struct {
	repo   user.CreateUserRepository
	hasher hasher.Hasher
}

func NewCreateUserUseCase(repo user.CreateUserRepository, h hasher.Hasher) user.CreateUserUseCase {
	return &createUserUseCase{
		repo:   repo,
		hasher: h,
	}
}

func (uc *createUserUseCase) Execute(ctx context.Context, input user.NewUserInput) (user.CreateUserOutput, error) {
	email, err := types.NewEmail(input.Email)
	if err != nil {
		return user.CreateUserOutput{}, err
	}

	phone, err := types.NewPhone(input.Phone)
	if err != nil {
		return user.CreateUserOutput{}, err
	}

	exists, err := uc.repo.UserExists(ctx, user.UserExistsRepoInput{Email: email, Phone: phone})
	if err != nil {
		return user.CreateUserOutput{}, msg.NewInternalError(err, nil)
	}

	if exists {
		return user.CreateUserOutput{}, msg.NewMessageError(
			nil,
			user.ErrUserAlreadyExists,
			msg.CodeConflict,
			map[string]any{"email": input.Email, "phone": input.Phone},
		)
	}

	newUser, err := user.NewUser(input, uc.hasher)
	if err != nil {
		return user.CreateUserOutput{}, err
	}

	repoInput := user.CreateUserRepoInput{
		User: newUser,
	}

	if err := uc.repo.CreateUser(ctx, repoInput); err != nil {
		return user.CreateUserOutput{}, msg.NewInternalError(err, nil)
	}

	output := user.CreateUserOutput{
		User: newUser,
	}

	return output, nil
}
