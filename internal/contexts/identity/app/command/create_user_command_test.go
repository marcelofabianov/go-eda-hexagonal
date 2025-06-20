package command_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/app/command"
	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/hasher"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

// --- Mocks for Dependencies ---

type mockCreateUserUseCase struct {
	ExecuteFunc func(ctx context.Context, input user.NewUserInput) (user.CreateUserOutput, error)
}

func (m *mockCreateUserUseCase) Execute(ctx context.Context, input user.NewUserInput) (user.CreateUserOutput, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, input)
	}
	return user.CreateUserOutput{}, nil
}

type mockUserPublisher struct {
	PublishUserCreatedEventFunc func(ctx context.Context, input user.USerCreatedEventInput) error
}

func (m *mockUserPublisher) PublishUserCreatedEvent(ctx context.Context, input user.USerCreatedEventInput) error {
	if m.PublishUserCreatedEventFunc != nil {
		return m.PublishUserCreatedEventFunc(ctx, input)
	}
	return nil
}

// --- Test Suite ---

func TestCreateUserCommand_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	newUserDomainInput := user.NewUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Phone:    "+15551234567",
		Password: "ValidPassword123!",
	}
	correlationID := types.MustNewUUID()
	traceID := types.MustNewUUID()
	authorUserID := types.NewNullableUUID(types.MustNewUUID(), true)
	previousEventID := types.NewNullableUUID(types.MustNewUUID(), true)
	causationID := types.NewNullableUUID(types.MustNewUUID(), true)

	commandInput := user.CreateUserCommandInput{
		CorrelationID:   correlationID,
		TraceID:         traceID,
		UserAuthorID:    authorUserID,
		PreviousEventID: previousEventID,
		CausationID:     causationID,
		NewUserInput:    newUserDomainInput,
	}

	t.Run("Success: should execute use case and publish event", func(t *testing.T) {
		mockUser, err := user.NewUser(newUserDomainInput, hasher.NewHasher())
		require.NoError(t, err, "Setup: Failed to create mock user")

		useCase := &mockCreateUserUseCase{
			ExecuteFunc: func(ctx context.Context, input user.NewUserInput) (user.CreateUserOutput, error) {
				return user.CreateUserOutput{User: mockUser}, nil
			},
		}

		publisherCalled := false
		publisher := &mockUserPublisher{
			PublishUserCreatedEventFunc: func(ctx context.Context, eventInput user.USerCreatedEventInput) error {
				publisherCalled = true

				assert.Equal(t, commandInput.CorrelationID, eventInput.CorrelationID, "Publisher eventInput.CorrelationID should match commandInput.CorrelationID")
				assert.Equal(t, commandInput.TraceID, eventInput.TraceID, "Publisher eventInput.TraceID should match commandInput.TraceID")
				assert.Equal(t, commandInput.UserAuthorID, eventInput.UserID, "Publisher eventInput.UserID (author) should match commandInput.UserAuthorID")
				assert.Equal(t, commandInput.PreviousEventID, eventInput.PreviousEventID, "Publisher eventInput.PreviousEventID should match commandInput.PreviousEventID")
				assert.Equal(t, commandInput.CausationID, eventInput.CausationID, "Publisher eventInput.CausationID should match commandInput.CausationID")

				assert.Equal(t, mockUser.ID, eventInput.Payload.UserID, "Publisher payload.UserID should match the created user ID")
				assert.Equal(t, mockUser.Name, eventInput.Payload.Name, "Publisher payload.Name should match the created user name")
				assert.Equal(t, mockUser.Email.String(), eventInput.Payload.Email, "Publisher payload.Email should match the created user email")
				assert.Equal(t, mockUser.Phone.String(), eventInput.Payload.Phone, "Publisher payload.Phone should match the created user phone")

				return nil
			},
		}

		cmd := command.NewCreateUserCommand(useCase, publisher, logger)

		output, err := cmd.Execute(context.Background(), commandInput)

		require.NoError(t, err, "Command Execute should not return an error on success")
		assert.Equal(t, mockUser.ID, output.User.ID, "Output should contain the user returned by the use case")
		assert.True(t, publisherCalled, "Publisher's PublishUserCreatedEvent method should have been called")
	})

	t.Run("Failure: should not publish event if use case returns an error", func(t *testing.T) {
		useCaseError := errors.New("use case failed")
		useCase := &mockCreateUserUseCase{
			ExecuteFunc: func(ctx context.Context, input user.NewUserInput) (user.CreateUserOutput, error) {
				return user.CreateUserOutput{}, useCaseError
			},
		}

		publisherCalled := false
		publisher := &mockUserPublisher{
			PublishUserCreatedEventFunc: func(ctx context.Context, input user.USerCreatedEventInput) error {
				publisherCalled = true
				return nil
			},
		}

		cmd := command.NewCreateUserCommand(useCase, publisher, logger)

		_, err := cmd.Execute(context.Background(), commandInput)

		require.Error(t, err, "Command Execute should return an error when use case fails")
		assert.Equal(t, useCaseError, err, "The error returned should be the one from the use case")
		assert.False(t, publisherCalled, "Publisher's method should NOT be called when the use case fails")
	})

	t.Run("Failure: should log error if event bus publish fails but not block command success", func(t *testing.T) {
		mockUser, err := user.NewUser(newUserDomainInput, hasher.NewHasher())
		require.NoError(t, err, "Setup: Failed to create mock user")

		useCase := &mockCreateUserUseCase{
			ExecuteFunc: func(ctx context.Context, input user.NewUserInput) (user.CreateUserOutput, error) {
				return user.CreateUserOutput{User: mockUser}, nil
			},
		}

		busError := errors.New("event bus publish failed (simulated)")
		publisher := &mockUserPublisher{
			PublishUserCreatedEventFunc: func(ctx context.Context, eventInput user.USerCreatedEventInput) error {
				return busError
			},
		}

		var logOutputBuffer = &mockLogOutput{}

		customLogger := slog.New(slog.NewTextHandler(logOutputBuffer, nil))

		cmd := command.NewCreateUserCommand(useCase, publisher, customLogger)

		output, err := cmd.Execute(context.Background(), commandInput)

		require.NoError(t, err, "Command Execute should not return an error even if event publication fails (fire-and-forget)")
		assert.Equal(t, mockUser.ID, output.User.ID, "Output should contain the user returned by the use case")

		assert.Contains(t, logOutputBuffer.String(), "level=ERROR", "Log output should contain an ERROR level entry")
		assert.Contains(t, logOutputBuffer.String(), "failed to publish user created event", "Log output should contain the specific error message")
		assert.Contains(t, logOutputBuffer.String(), busError.Error(), "Log output should contain the bus error message")
	})
}

type mockLogOutput struct {
	content []byte
}

func (m *mockLogOutput) Write(p []byte) (n int, err error) {
	m.content = append(m.content, p...)
	return len(p), nil
}

func (m *mockLogOutput) String() string {
	return string(m.content)
}
