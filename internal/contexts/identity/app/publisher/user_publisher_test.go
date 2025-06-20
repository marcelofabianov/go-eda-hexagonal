package publisher_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/app/publisher"
	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/event"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

// Mock for EventBusPublisher
type mockEventBusPublisher struct {
	PublishFunc func(ctx context.Context, event *event.Event) error
}

func (m *mockEventBusPublisher) Publish(ctx context.Context, event *event.Event) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, event)
	}
	return nil
}

func TestUserPublisher_PublishUserCreatedEvent(t *testing.T) {
	// Common test data setup for all sub-tests
	correlationID := types.MustNewUUID()
	userAuthorIDValue := types.MustNewUUID()
	userAuthorID := types.NewValidNullableUUID(userAuthorIDValue)
	traceID := types.MustNewUUID()
	previousEventIDValue := types.MustNewUUID()
	previousEventID := types.NewValidNullableUUID(previousEventIDValue)
	causationIDValue := types.MustNewUUID()
	causationID := types.NewValidNullableUUID(causationIDValue)

	inputPayload := user.UserCreatedPayload{
		UserID: types.MustNewUUID(),
		Name:   "Test User",
		Email:  "test@example.com",
		Phone:  "+15551234567",
	}

	// Create the USerCreatedEventInput struct to be passed to the publisher
	publisherInput := user.USerCreatedEventInput{
		CorrelationID:   correlationID,
		UserID:          userAuthorID,
		TraceID:         traceID,
		PreviousEventID: previousEventID,
		CausationID:     causationID,
		Payload:         inputPayload,
	}

	t.Run("Success: should publish user created event successfully", func(t *testing.T) {
		publishCalled := false
		mockBus := &mockEventBusPublisher{
			PublishFunc: func(ctx context.Context, event *event.Event) error {
				publishCalled = true
				// Assertions for Header
				assert.Equal(t, user.UserCreatedEventType, event.Header.EventType, "Event Header.EventType should match user.UserCreatedEventType")
				assert.Equal(t, user.UserCreatedEventVersion, event.Header.SchemaVersion, "Event Header.SchemaVersion should match user.UserCreatedEventVersion") // Changed to SchemaVersion
				assert.Equal(t, user.UserEventSource, event.Header.Source, "Event Header.Source should match user.UserEventSource")                               // Assert Source

				// Assertions for Context
				assert.Equal(t, publisherInput.CorrelationID, event.Context.CorrelationID, "Event Context.CorrelationID should match the provided CorrelationID")
				assert.Equal(t, publisherInput.UserID, event.Context.UserID, "Event Context.UserID should match the provided UserID")

				// Assertions for Metadata
				assert.Equal(t, publisherInput.TraceID, event.Metadata.TraceID, "Event Metadata.TraceID should match the provided TraceID")
				assert.Equal(t, publisherInput.PreviousEventID, event.Metadata.PreviousEventID, "Event Metadata.PreviousEventID should match the provided PreviousEventID")
				assert.Equal(t, publisherInput.CausationID, event.Metadata.CausationID, "Event Metadata.CausationID should match the provided CausationID")

				// Assertions for Payload
				var actualPayload user.UserCreatedPayload
				err := json.Unmarshal(event.Payload, &actualPayload)
				require.NoError(t, err, "Failed to unmarshal event payload")
				assert.Equal(t, publisherInput.Payload, actualPayload, "Event Payload should match the provided UserCreatedPayload")
				return nil
			},
		}
		userPublisher := publisher.NewUserPublisher(mockBus)

		err := userPublisher.PublishUserCreatedEvent(context.Background(), publisherInput) // Pass the input struct

		require.NoError(t, err, "PublishUserCreatedEvent should not return an error when event creation and publication succeed")
		assert.True(t, publishCalled, "EventBusPublisher.Publish should be called when publishing the user created event")
	})

	t.Run("Failure: should return error if event bus publish fails", func(t *testing.T) {
		busError := errors.New("event bus publish failed")
		mockBus := &mockEventBusPublisher{
			PublishFunc: func(ctx context.Context, event *event.Event) error {
				return busError
			},
		}
		userPublisher := publisher.NewUserPublisher(mockBus)

		err := userPublisher.PublishUserCreatedEvent(context.Background(), publisherInput) // Pass the input struct

		require.Error(t, err, "PublishUserCreatedEvent should return an error when EventBusPublisher.Publish fails")
		assert.Equal(t, busError, err, "Error returned by PublishUserCreatedEvent should match the EventBusPublisher.Publish error")
	})
}
