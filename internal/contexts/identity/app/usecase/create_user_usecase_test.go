package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/app/usecase"
	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/hasher"
	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

type mockCreateUserRepo struct {
	UserExistsFunc func(ctx context.Context, input user.UserExistsRepoInput) (bool, error)
	CreateUserFunc func(ctx context.Context, input user.CreateUserRepoInput) error
}

func (m *mockCreateUserRepo) UserExists(ctx context.Context, input user.UserExistsRepoInput) (bool, error) {
	if m.UserExistsFunc != nil {
		return m.UserExistsFunc(ctx, input)
	}
	return false, nil
}

func (m *mockCreateUserRepo) CreateUser(ctx context.Context, input user.CreateUserRepoInput) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, input)
	}
	return nil
}

func TestCreateUserUseCase_Execute(t *testing.T) {
	h := hasher.NewHasher()
	validInput := user.NewUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Phone:    "5562999998888",
		Password: "ValidPassword123!",
	}

	t.Run("Success: should create a user when all data is valid and user does not exist", func(t *testing.T) {
		mockRepo := &mockCreateUserRepo{
			UserExistsFunc: func(ctx context.Context, input user.UserExistsRepoInput) (bool, error) {
				return false, nil
			},
			CreateUserFunc: func(ctx context.Context, input user.CreateUserRepoInput) error {
				return nil
			},
		}
		uc := usecase.NewCreateUserUseCase(mockRepo, h)

		output, err := uc.Execute(context.Background(), validInput)

		require.NoError(t, err, "Execute should not return an error on success")
		assert.NotNil(t, output.User, "Returned user should not be nil")
		assert.Equal(t, validInput.Name, output.User.Name, "User name should match input")
	})

	t.Run("Failure: should return conflict error if user already exists", func(t *testing.T) {
		mockRepo := &mockCreateUserRepo{
			UserExistsFunc: func(ctx context.Context, input user.UserExistsRepoInput) (bool, error) {
				return true, nil
			},
		}
		uc := usecase.NewCreateUserUseCase(mockRepo, h)

		_, err := uc.Execute(context.Background(), validInput)

		require.Error(t, err, "Execute should return an error when user exists")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be a MessageError")
		assert.Equal(t, msg.CodeConflict, msgErr.Code, "Error code should be CodeConflict")
	})

	t.Run("Failure: should return internal error if UserExists check fails", func(t *testing.T) {
		dbError := errors.New("database connection error")
		mockRepo := &mockCreateUserRepo{
			UserExistsFunc: func(ctx context.Context, input user.UserExistsRepoInput) (bool, error) {
				return false, dbError
			},
		}
		uc := usecase.NewCreateUserUseCase(mockRepo, h)

		_, err := uc.Execute(context.Background(), validInput)

		require.Error(t, err, "Execute should return an error if repository check fails")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be a MessageError")
		assert.Equal(t, msg.CodeInternal, msgErr.Code, "Error code should be CodeInternal")
	})

	t.Run("Failure: should return validation error for invalid input email", func(t *testing.T) {
		invalidInput := validInput
		invalidInput.Email = "not-an-email"
		mockRepo := &mockCreateUserRepo{}
		uc := usecase.NewCreateUserUseCase(mockRepo, h)

		_, err := uc.Execute(context.Background(), invalidInput)

		require.Error(t, err, "Execute should return an error for invalid email")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be a MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
	})

	t.Run("Failure: should return internal error if CreateUser fails in repository", func(t *testing.T) {
		dbError := errors.New("failed to insert user")
		mockRepo := &mockCreateUserRepo{
			UserExistsFunc: func(ctx context.Context, input user.UserExistsRepoInput) (bool, error) {
				return false, nil
			},
			CreateUserFunc: func(ctx context.Context, input user.CreateUserRepoInput) error {
				return dbError
			},
		}
		uc := usecase.NewCreateUserUseCase(mockRepo, h)

		_, err := uc.Execute(context.Background(), validInput)

		require.Error(t, err, "Execute should return an error if CreateUser fails")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be a MessageError")
		assert.Equal(t, msg.CodeInternal, msgErr.Code, "Error code should be CodeInternal")
	})
}
