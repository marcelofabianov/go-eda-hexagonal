package subscriber_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/contexts/audit/app/subscriber"
	"github.com/marcelofabianov/redtogreen/internal/contexts/audit/domain/audit"
	identityUser "github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/event"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

type mockRegisterAuditLogRepository struct {
	RegisterAuditLogFunc func(ctx context.Context, input audit.RegisterAuditLogRepoInput) error
}

func (m *mockRegisterAuditLogRepository) RegisterAuditLog(ctx context.Context, input audit.RegisterAuditLogRepoInput) error {
	if m.RegisterAuditLogFunc != nil {
		return m.RegisterAuditLogFunc(ctx, input)
	}
	return nil
}

type mockLogOutput struct {
	b strings.Builder
}

func (m *mockLogOutput) Write(p []byte) (n int, err error) {
	return m.b.Write(p)
}

func (m *mockLogOutput) String() string {
	return m.b.String()
}

func TestUserCreatedSubscriber_Handle(t *testing.T) {
	eventID := types.MustNewUUID()
	correlationID := types.MustNewUUID()
	traceID := types.MustNewUUID()
	userAuthorID := types.NewValidNullableUUID(types.MustNewUUID())
	previousEventID := types.NewValidNullableUUID(types.MustNewUUID())
	causationID := types.NewValidNullableUUID(types.MustNewUUID())

	userPayload := identityUser.UserCreatedPayload{
		UserID: types.MustNewUUID(),
		Name:   "Subscriber Test User",
		Email:  "subscriber.test@example.com",
		Phone:  "987654321",
	}
	userPayloadBytes, err := json.Marshal(userPayload)
	require.NoError(t, err, "Failed to marshal user payload for test setup")

	mockEvent := &event.Event{
		Header: event.EventHeader{
			EventID:       eventID,
			EventType:     identityUser.UserCreatedEventType,
			Timestamp:     time.Now().UTC(),
			Source:        identityUser.UserEventSource,
			SchemaVersion: identityUser.UserCreatedEventVersion,
		},
		Context: event.EventContext{
			CorrelationID: correlationID,
			UserID:        userAuthorID,
		},
		Metadata: event.EventMetadata{
			TraceID:         traceID,
			PreviousEventID: previousEventID,
			CausationID:     causationID,
		},
		Payload: userPayloadBytes,
	}

	t.Run("Success: should handle user.created event and register audit log", func(t *testing.T) {
		logOutputBuffer := &mockLogOutput{}
		logger := slog.New(slog.NewTextHandler(logOutputBuffer, nil))

		repoCalled := false
		mockRepo := &mockRegisterAuditLogRepository{
			RegisterAuditLogFunc: func(ctx context.Context, input audit.RegisterAuditLogRepoInput) error {
				repoCalled = true
				assert.NotNil(t, input.AuditLog.ID, "AuditLog ID should not be nil")
				assert.Equal(t, mockEvent.Header.EventType, input.AuditLog.EventType, "AuditLog EventType should match event header")
				return nil
			},
		}

		s := subscriber.NewUserCreatedSubscriber(mockRepo, logger)
		err := s.Handle(context.Background(), mockEvent)

		require.NoError(t, err, "Handle should not return an error on success")
		assert.True(t, repoCalled, "RegisterAuditLog should have been called")

		logString := logOutputBuffer.String()
		assert.Contains(t, logString, "level=INFO", "Should log INFO level")
		assert.Contains(t, logString, "component=audit_subscriber", "Should log component")
		assert.Contains(t, logString, "action=handle_user_created_event", "Should log action")
		assert.Contains(t, logString, "event_id="+mockEvent.Header.EventID.String(), "Should log event ID")
		assert.Contains(t, logString, "event_type="+string(mockEvent.Header.EventType), "Should log event type")
		assert.Contains(t, logString, "msg=\"user.created event audited successfully\"", "Should log success message")
	})

	t.Run("Failure: should return error if payload unmarshalling fails", func(t *testing.T) {
		logOutputBuffer := &mockLogOutput{}
		logger := slog.New(slog.NewTextHandler(logOutputBuffer, nil))

		invalidPayloadEvent := &event.Event{
			Header:   mockEvent.Header,
			Context:  mockEvent.Context,
			Metadata: mockEvent.Metadata,
			Payload:  json.RawMessage(`{"invalid_json":}`),
		}

		mockRepo := &mockRegisterAuditLogRepository{}
		s := subscriber.NewUserCreatedSubscriber(mockRepo, logger)
		err := s.Handle(context.Background(), invalidPayloadEvent)

		require.Error(t, err, "Handle should return an error if payload unmarshalling fails")

		logString := logOutputBuffer.String()
		assert.Contains(t, logString, "level=ERROR", "Should log ERROR level")
		assert.Contains(t, logString, "component=audit_subscriber", "Should log component")
		assert.Contains(t, logString, "action=handle_user_created_event", "Should log action")
		assert.Contains(t, logString, "event_id="+invalidPayloadEvent.Header.EventID.String(), "Should log event ID")
		assert.Contains(t, logString, "event_type="+string(invalidPayloadEvent.Header.EventType), "Should log event type")
		assert.Contains(t, logString, "msg=\"failed to unmarshal user.created event payload for auditing\"", "Should log unmarshal failure message")
		assert.Contains(t, logString, "err=\"invalid character '}' looking for beginning of value\"", "Should log the error attribute")
		assert.Contains(t, logString, "payload_content=\"{\\\"invalid_json\\\":}\"", "Should log payload content")
	})

	t.Run("Failure: should return error if audit log repository fails", func(t *testing.T) {
		logOutputBuffer := &mockLogOutput{}
		logger := slog.New(slog.NewTextHandler(logOutputBuffer, nil))

		repoError := errors.New("failed to save audit log")
		mockRepo := &mockRegisterAuditLogRepository{
			RegisterAuditLogFunc: func(ctx context.Context, input audit.RegisterAuditLogRepoInput) error {
				return repoError
			},
		}

		s := subscriber.NewUserCreatedSubscriber(mockRepo, logger)
		err := s.Handle(context.Background(), mockEvent)

		require.Error(t, err, "Handle should return an error if repository fails")
		assert.Equal(t, repoError, err, "The returned error should be the one from the repository")

		logString := logOutputBuffer.String()
		assert.Contains(t, logString, "level=ERROR", "Should log ERROR level")
		assert.Contains(t, logString, "component=audit_subscriber", "Should log component")
		assert.Contains(t, logString, "action=handle_user_created_event", "Should log action")
		assert.Contains(t, logString, "event_id="+mockEvent.Header.EventID.String(), "Should log event ID")
		assert.Contains(t, logString, "event_type="+string(mockEvent.Header.EventType), "Should log event type")
		assert.Contains(t, logString, "msg=\"failed to save audit log to repository\"", "Should log repository failure message")
		assert.Contains(t, logString, "err=\""+repoError.Error()+"\"", "Should log the repository error details")
	})

	t.Run("Failure: should return error if audit.NewAuditLog fails", func(t *testing.T) {
		logOutputBuffer := &mockLogOutput{}
		logger := slog.New(slog.NewTextHandler(logOutputBuffer, nil))

		mockRepo := &mockRegisterAuditLogRepository{}
		s := subscriber.NewUserCreatedSubscriber(mockRepo, logger)

		mockEventWithInvalidAuditInput := &event.Event{
			Header: event.EventHeader{
				EventType: "",
			},
			Context:  mockEvent.Context,
			Metadata: mockEvent.Metadata,
			Payload:  mockEvent.Payload,
		}

		err := s.Handle(context.Background(), mockEventWithInvalidAuditInput)

		require.Error(t, err, "Handle should return an error if audit.NewAuditLog fails")

		logString := logOutputBuffer.String()
		assert.Contains(t, logString, "level=ERROR", "Should log ERROR level")
		assert.Contains(t, logString, "component=audit_subscriber", "Should log component")
		assert.Contains(t, logString, "action=handle_user_created_event", "Should log action")
		assert.Contains(t, logString, "msg=\"failed to create audit log entity from event\"", "Should log audit log creation failure message")
		assert.Contains(t, logString, "err=", "Should log the error attribute")
	})
}
