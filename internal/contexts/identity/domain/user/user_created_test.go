package user

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewUserCreatedEvent(t *testing.T) {
	t.Run("Success_WithAllValidFields_ShouldCreateUserCreatedEventCorrectly", func(t *testing.T) {
		correlationID := types.MustNewUUID()
		actorUserIDValue := types.MustNewUUID()
		actorUserID := types.NewValidNullableUUID(actorUserIDValue)
		traceID := types.MustNewUUID()
		previousEventIDValue := types.MustNewUUID()
		previousEventID := types.NewValidNullableUUID(previousEventIDValue)
		causationIDValue := types.MustNewUUID()
		causationID := types.NewValidNullableUUID(causationIDValue)

		inputPayload := UserCreatedPayload{
			UserID: types.MustNewUUID(),
			Name:   "Test User",
			Email:  "test@example.com",
			Phone:  "1234567890",
		}

		input := USerCreatedEventInput{
			CorrelationID:   correlationID,
			UserID:          actorUserID,
			TraceID:         traceID,
			PreviousEventID: previousEventID,
			CausationID:     causationID,
			Payload:         inputPayload,
		}

		beforeCall := time.Now().UTC()
		event, err := NewUserCreatedEvent(input)
		afterCall := time.Now().UTC()

		require.NoError(t, err, "NewUserCreatedEvent should not return an error")
		require.NotNil(t, event, "The event should not be nil")

		// Header assertions
		assert.NotEqual(t, types.Nil, event.Header.EventID, "Header.EventID should be a non-nil UUID")
		_, uuidErr := types.ParseUUID(event.Header.EventID.String())
		assert.NoError(t, uuidErr, "Header.EventID should be a valid UUID")
		assert.Equal(t, UserCreatedEventType, event.Header.EventType, "Header.EventType should match")
		assert.Equal(t, UserCreatedEventVersion, event.Header.SchemaVersion, "Header.SchemaVersion should match")
		assert.Equal(t, UserEventSource, event.Header.Source, "Header.Source should match")
		assert.True(t, (event.Header.Timestamp.Equal(beforeCall) || event.Header.Timestamp.After(beforeCall)) && (event.Header.Timestamp.Equal(afterCall) || event.Header.Timestamp.Before(afterCall)),
			"Header.Timestamp should be within the call window. Got: %v, Before: %v, After: %v", event.Header.Timestamp, beforeCall, afterCall)
		assert.Equal(t, time.UTC, event.Header.Timestamp.Location(), "Header.Timestamp should be in UTC")

		// Context assertions
		assert.Equal(t, input.CorrelationID, event.Context.CorrelationID, "Context.CorrelationID should match input")
		assert.Equal(t, input.UserID, event.Context.UserID, "Context.UserID should match input")
		assert.True(t, event.Context.UserID.Valid, "Context.UserID should be valid")
		actualContextUserID, _ := event.Context.UserID.GetUUID()
		assert.Equal(t, actorUserIDValue, actualContextUserID, "Context.UserID UUID value should match")

		// Metadata assertions
		assert.Equal(t, input.TraceID, event.Metadata.TraceID, "Metadata.TraceID should match input")
		assert.Equal(t, input.PreviousEventID, event.Metadata.PreviousEventID, "Metadata.PreviousEventID should match input")
		assert.True(t, event.Metadata.PreviousEventID.Valid, "Metadata.PreviousEventID should be valid")
		actualPreviousEventID, _ := event.Metadata.PreviousEventID.GetUUID()
		assert.Equal(t, previousEventIDValue, actualPreviousEventID, "Metadata.PreviousEventID UUID value should match")

		assert.Equal(t, input.CausationID, event.Metadata.CausationID, "Metadata.CausationID should match input")
		assert.True(t, event.Metadata.CausationID.Valid, "Metadata.CausationID should be valid")
		actualCausationID, _ := event.Metadata.CausationID.GetUUID()
		assert.Equal(t, causationIDValue, actualCausationID, "Metadata.CausationID UUID value should match")

		// Payload assertions
		var payload UserCreatedPayload
		err = json.Unmarshal(event.Payload, &payload)
		require.NoError(t, err, "Payload unmarshalling should not return error")
		assert.Equal(t, inputPayload.UserID, payload.UserID, "Payload.UserID should match")
		assert.Equal(t, inputPayload.Name, payload.Name, "Payload.Name should match")
		assert.Equal(t, inputPayload.Email, payload.Email, "Payload.Email should match")
		assert.Equal(t, inputPayload.Phone, payload.Phone, "Payload.Phone should match")
	})

	t.Run("Success_WithNilOptionalFields_ShouldCreateUserCreatedEventCorrectly", func(t *testing.T) {
		correlationID := types.MustNewUUID()
		traceID := types.MustNewUUID()
		var nilActorUserID types.NullableUUID
		var nilPreviousEventID types.NullableUUID
		var nilCausationID types.NullableUUID

		inputPayload := UserCreatedPayload{
			UserID: types.MustNewUUID(),
			Name:   "System Generated User",
			Email:  "systemgen@example.com",
			Phone:  "0000000000",
		}

		input := USerCreatedEventInput{
			CorrelationID:   correlationID,
			UserID:          nilActorUserID,
			TraceID:         traceID,
			PreviousEventID: nilPreviousEventID,
			CausationID:     nilCausationID,
			Payload:         inputPayload,
		}

		event, err := NewUserCreatedEvent(input)

		require.NoError(t, err, "NewUserCreatedEvent should not return error")
		require.NotNil(t, event, "The event should not be nil")

		// Header assertions
		assert.Equal(t, UserCreatedEventType, event.Header.EventType)
		assert.Equal(t, UserCreatedEventVersion, event.Header.SchemaVersion)
		assert.Equal(t, UserEventSource, event.Header.Source)

		// Context assertions
		assert.Equal(t, input.CorrelationID, event.Context.CorrelationID, "Context.CorrelationID should match input")
		assert.False(t, event.Context.UserID.Valid, "Context.UserID should be invalid (nil)")
		actualContextUserID, valid := event.Context.UserID.GetUUID()
		assert.False(t, valid, "Context.UserID GetUUID should return false for validity")
		assert.Equal(t, types.Nil, actualContextUserID, "Context.UserID UUID value should be Nil")

		// Metadata assertions
		assert.Equal(t, input.TraceID, event.Metadata.TraceID, "Metadata.TraceID should match input")
		assert.False(t, event.Metadata.PreviousEventID.Valid, "Metadata.PreviousEventID should be invalid (nil)")
		assert.False(t, event.Metadata.CausationID.Valid, "Metadata.CausationID should be invalid (nil)")

		// Payload assertions
		var payload UserCreatedPayload
		err = json.Unmarshal(event.Payload, &payload)
		require.NoError(t, err, "Payload unmarshalling should not return error")
		assert.Equal(t, inputPayload.UserID, payload.UserID, "Payload.UserID should match")
		assert.Equal(t, inputPayload.Name, payload.Name, "Payload.Name should match")
	})
}
