package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewEventHeader(t *testing.T) {
	t.Run("Success_ShouldCreateHeaderWithGeneratedIDAndCurrentTimestamp", func(t *testing.T) {
		eventType := EventType("test.event")
		eventVersion := EventVersion("v1")
		source := "TestService"

		beforeCall := time.Now().UTC()
		header, err := NewEventHeader(eventType, eventVersion, source)
		afterCall := time.Now().UTC()

		require.NoError(t, err, "NewEventHeader should not return an error on success")

		assert.NotEqual(t, types.Nil, header.EventID, "EventID should be a non-zero UUID")
		assert.NotEmpty(t, header.EventID.String(), "EventID should not be an empty string")
		_, uuidErr := types.ParseUUID(header.EventID.String())
		assert.NoError(t, uuidErr, "EventID should be a valid UUID")

		assert.Equal(t, eventType, header.EventType, "EventType should match the input")
		assert.Equal(t, eventVersion, header.SchemaVersion, "EventVersion should match the input")
		assert.Equal(t, source, header.Source, "Source should match the input")

		assert.True(t, (header.Timestamp.Equal(beforeCall) || header.Timestamp.After(beforeCall)) && (header.Timestamp.Equal(afterCall) || header.Timestamp.Before(afterCall)),
			"Timestamp should be between just before and just after the call. Got: %v, Before: %v, After: %v", header.Timestamp, beforeCall, afterCall)
		assert.Equal(t, time.UTC, header.Timestamp.Location(), "Timestamp should be in UTC")
	})
}

func TestNewEventContext(t *testing.T) {
	t.Run("WithValidUserID_ShouldCreateContextCorrectly", func(t *testing.T) {
		correlationID := types.MustNewUUID()
		userIDValue := types.MustNewUUID()
		userID := types.NewValidNullableUUID(userIDValue)

		context := NewEventContext(correlationID, userID)

		assert.Equal(t, correlationID, context.CorrelationID, "CorrelationID should match the input")
		assert.Equal(t, userID, context.UserID, "UserID should match the input")
		assert.True(t, context.UserID.Valid, "UserID should be valid")
		actualUserIDValue, _ := context.UserID.GetUUID()
		assert.Equal(t, userIDValue, actualUserIDValue, "UserID UUID value should match")
	})

	t.Run("WithNilUserID_ShouldCreateContextCorrectly", func(t *testing.T) {
		correlationID := types.MustNewUUID()
		var nilUserID types.NullableUUID

		context := NewEventContext(correlationID, nilUserID)

		assert.Equal(t, correlationID, context.CorrelationID, "CorrelationID should match the input")
		assert.Equal(t, nilUserID, context.UserID, "UserID should match the input (nil)")
		assert.False(t, context.UserID.Valid, "UserID should be invalid (nil)")
		actualUserIDValue, _ := context.UserID.GetUUID()
		assert.Equal(t, types.Nil, actualUserIDValue, "UserID UUID value should be zero when nil")
	})
}

func TestNewEventMetadata(t *testing.T) {
	t.Run("WithAllFields_ShouldCreateMetadataCorrectly", func(t *testing.T) {
		traceID := types.MustNewUUID()
		previousEventIDValue := types.MustNewUUID()
		previousEventID := types.NewValidNullableUUID(previousEventIDValue)
		causationIDValue := types.MustNewUUID()
		causationID := types.NewValidNullableUUID(causationIDValue)

		metadata := NewEventMetadata(traceID, previousEventID, causationID)

		assert.Equal(t, traceID, metadata.TraceID, "TraceID should match the input")
		assert.Equal(t, previousEventID, metadata.PreviousEventID, "PreviousEventID should match the input")
		assert.True(t, metadata.PreviousEventID.Valid, "PreviousEventID should be valid")
		actualPreviousIDValue, _ := metadata.PreviousEventID.GetUUID()
		assert.Equal(t, previousEventIDValue, actualPreviousIDValue, "PreviousEventID UUID value should match")

		assert.Equal(t, causationID, metadata.CausationID, "CausationID should match the input")
		assert.True(t, metadata.CausationID.Valid, "CausationID should be valid")
		actualCausationIDValue, _ := metadata.CausationID.GetUUID()
		assert.Equal(t, causationIDValue, actualCausationIDValue, "CausationID UUID value should match")
	})

	t.Run("WithOnlyTraceID_ShouldCreateMetadataCorrectly", func(t *testing.T) {
		traceID := types.MustNewUUID()
		var nilPreviousEventID types.NullableUUID
		var nilCausationID types.NullableUUID

		metadata := NewEventMetadata(traceID, nilPreviousEventID, nilCausationID)

		assert.Equal(t, traceID, metadata.TraceID, "TraceID should match the input")
		assert.False(t, metadata.PreviousEventID.Valid, "PreviousEventID should be invalid (nil)")
		assert.False(t, metadata.CausationID.Valid, "CausationID should be invalid (nil)")
	})
}

func TestNewEvent(t *testing.T) {
	// Existing success tests
	t.Run("Success_ShouldCreateEventWithAllFieldsPopulated", func(t *testing.T) {
		correlationID := types.MustNewUUID()
		userIDValue := types.MustNewUUID()
		userID := types.NewValidNullableUUID(userIDValue)
		traceID := types.MustNewUUID()
		previousEventIDValue := types.MustNewUUID()
		previousEventID := types.NewValidNullableUUID(previousEventIDValue)
		causationIDValue := types.MustNewUUID()
		causationID := types.NewValidNullableUUID(causationIDValue)

		payloadStruct := struct{ Data string }{Data: "test-payload-content"}
		payloadBytes, err := json.Marshal(payloadStruct)
		require.NoError(t, err, "json.Marshal should not fail for test payload")

		input := EventInput{
			EventType:       EventType("sample.event.type"),
			EventVersion:    EventVersion("v1.0.1"),
			Source:          "OrderService",
			CorrelationID:   correlationID,
			UserID:          userID,
			TraceID:         traceID,
			PreviousEventID: previousEventID,
			CausationID:     causationID,
			Payload:         payloadBytes,
		}

		beforeCall := time.Now().UTC()
		event, err := NewEvent(input)
		afterCall := time.Now().UTC()

		require.NoError(t, err, "NewEvent should not return an error on success")

		assert.NotEqual(t, types.Nil, event.Header.EventID, "Header.EventID should be a non-zero UUID")
		_, uuidErr := types.ParseUUID(event.Header.EventID.String())
		assert.NoError(t, uuidErr, "Header.EventID should be a valid UUID")
		assert.Equal(t, input.EventType, event.Header.EventType, "Header.EventType should match input")
		assert.Equal(t, input.EventVersion, event.Header.SchemaVersion, "Header.EventVersion should match input")
		assert.Equal(t, input.Source, event.Header.Source, "Header.Source should match input")
		assert.True(t, (event.Header.Timestamp.Equal(beforeCall) || event.Header.Timestamp.After(beforeCall)) && (event.Header.Timestamp.Equal(afterCall) || event.Header.Timestamp.Before(afterCall)),
			"Header.Timestamp should be between just before and just after the call. Got: %v, Before: %v, After: %v", event.Header.Timestamp, beforeCall, afterCall)
		assert.Equal(t, time.UTC, event.Header.Timestamp.Location(), "Header.Timestamp should be in UTC")

		assert.Equal(t, input.CorrelationID, event.Context.CorrelationID, "Context.CorrelationID should match input")
		assert.Equal(t, input.UserID, event.Context.UserID, "Context.UserID should match input")
		assert.True(t, event.Context.UserID.Valid, "Context.UserID should be valid")
		actualEventUserIDValue, _ := event.Context.UserID.GetUUID()
		assert.Equal(t, userIDValue, actualEventUserIDValue, "Context.UserID UUID value should match")

		assert.Equal(t, input.TraceID, event.Metadata.TraceID, "Metadata.TraceID should match input")
		assert.Equal(t, input.PreviousEventID, event.Metadata.PreviousEventID, "Metadata.PreviousEventID should match input")
		assert.True(t, event.Metadata.PreviousEventID.Valid, "Metadata.PreviousEventID should be valid")
		actualPreviousEventIDValue, _ := event.Metadata.PreviousEventID.GetUUID()
		assert.Equal(t, previousEventIDValue, actualPreviousEventIDValue, "Metadata.PreviousEventID UUID value should match")

		assert.Equal(t, input.CausationID, event.Metadata.CausationID, "Metadata.CausationID should match input")
		assert.True(t, event.Metadata.CausationID.Valid, "Metadata.CausationID should be valid")
		actualCausationIDValue, _ := event.Metadata.CausationID.GetUUID()
		assert.Equal(t, causationIDValue, actualCausationIDValue, "Metadata.CausationID UUID value should match")

		assert.Equal(t, payloadBytes, []byte(event.Payload), "Payload should match input")
	})

	t.Run("Success_WithNilOptionalFields_ShouldCreateEventCorrectly", func(t *testing.T) {
		correlationID := types.MustNewUUID()
		traceID := types.MustNewUUID()
		var nilUserID types.NullableUUID
		var nilPreviousEventID types.NullableUUID
		var nilCausationID types.NullableUUID

		payloadStruct := struct{ Info string }{Info: "system-event-payload"}
		payloadBytes, err := json.Marshal(payloadStruct)
		require.NoError(t, err, "json.Marshal should not fail for test payload")

		input := EventInput{
			EventType:       EventType("system.event.type"),
			EventVersion:    EventVersion("v2.0.0"),
			Source:          "SystemMonitor",
			CorrelationID:   correlationID,
			UserID:          nilUserID,
			TraceID:         traceID,
			PreviousEventID: nilPreviousEventID,
			CausationID:     nilCausationID,
			Payload:         payloadBytes,
		}

		event, err := NewEvent(input)
		require.NoError(t, err, "NewEvent should not return an error with nil optional fields")
		require.NotNil(t, event, "Event should not be nil")

		assert.NotEqual(t, types.Nil, event.Header.EventID, "Header.EventID should be a non-zero UUID")
		assert.Equal(t, input.EventType, event.Header.EventType)
		assert.Equal(t, input.Source, event.Header.Source)

		assert.Equal(t, input.CorrelationID, event.Context.CorrelationID, "Context.CorrelationID should match input")
		assert.False(t, event.Context.UserID.Valid, "Context.UserID should be invalid (nil)")
		actualEventUserIDValue, _ := event.Context.UserID.GetUUID()
		assert.Equal(t, types.Nil, actualEventUserIDValue, "Context.UserID UUID value should be zero when nil")

		assert.Equal(t, input.TraceID, event.Metadata.TraceID, "Metadata.TraceID should match input")
		assert.False(t, event.Metadata.PreviousEventID.Valid, "Metadata.PreviousEventID should be invalid (nil)")
		assert.False(t, event.Metadata.CausationID.Valid, "Metadata.CausationID should be invalid (nil)")

		assert.Equal(t, payloadBytes, []byte(event.Payload))
	})

	// NEW TESTS FOR VALIDATION FAILURES
	t.Run("Failure_EmptyEventType", func(t *testing.T) {
		input := EventInput{
			EventType:     "", // Empty
			EventVersion:  EventVersion("v1"),
			Source:        "TestService",
			CorrelationID: types.MustNewUUID(),
			UserID:        types.NewNullUUID(),
			TraceID:       types.MustNewUUID(),
			Payload:       json.RawMessage(`{"key":"value"}`),
		}
		_, err := NewEvent(input)
		require.Error(t, err)
		assert.EqualError(t, err, "event type cannot be empty")
	})

	t.Run("Failure_EmptySource", func(t *testing.T) {
		input := EventInput{
			EventType:     EventType("test.event"),
			EventVersion:  EventVersion("v1"),
			Source:        "", // Empty
			CorrelationID: types.MustNewUUID(),
			UserID:        types.NewNullUUID(),
			TraceID:       types.MustNewUUID(),
			Payload:       json.RawMessage(`{"key":"value"}`),
		}
		_, err := NewEvent(input)
		require.Error(t, err)
		assert.EqualError(t, err, "event source cannot be empty")
	})

	t.Run("Failure_EmptyEventVersion", func(t *testing.T) {
		input := EventInput{
			EventType:     EventType("test.event"),
			EventVersion:  "", // Empty
			Source:        "TestService",
			CorrelationID: types.MustNewUUID(),
			UserID:        types.NewNullUUID(),
			TraceID:       types.MustNewUUID(),
			Payload:       json.RawMessage(`{"key":"value"}`),
		}
		_, err := NewEvent(input)
		require.Error(t, err)
		assert.EqualError(t, err, "event version cannot be empty")
	})

	t.Run("Failure_NilCorrelationID", func(t *testing.T) {
		input := EventInput{
			EventType:     EventType("test.event"),
			EventVersion:  EventVersion("v1"),
			Source:        "TestService",
			CorrelationID: types.Nil, // Nil
			UserID:        types.NewNullUUID(),
			TraceID:       types.MustNewUUID(),
			Payload:       json.RawMessage(`{"key":"value"}`),
		}
		_, err := NewEvent(input)
		require.Error(t, err)
		assert.EqualError(t, err, "correlation ID cannot be empty")
	})

	t.Run("Failure_NilTraceID", func(t *testing.T) {
		input := EventInput{
			EventType:     EventType("test.event"),
			EventVersion:  EventVersion("v1"),
			Source:        "TestService",
			CorrelationID: types.MustNewUUID(),
			UserID:        types.NewNullUUID(),
			TraceID:       types.Nil, // Nil
			Payload:       json.RawMessage(`{"key":"value"}`),
		}
		_, err := NewEvent(input)
		require.Error(t, err)
		assert.EqualError(t, err, "trace ID cannot be empty")
	})

	t.Run("Failure_EmptyPayload", func(t *testing.T) {
		input := EventInput{
			EventType:     EventType("test.event"),
			EventVersion:  EventVersion("v1"),
			Source:        "TestService",
			CorrelationID: types.MustNewUUID(),
			UserID:        types.NewNullUUID(),
			TraceID:       types.MustNewUUID(),
			Payload:       json.RawMessage(nil), // Nil payload
		}
		_, err := NewEvent(input)
		require.Error(t, err)
		assert.EqualError(t, err, "event payload cannot be empty")
	})

	t.Run("Failure_InvalidJSONPayload", func(t *testing.T) {
		input := EventInput{
			EventType:     EventType("test.event"),
			EventVersion:  EventVersion("v1"),
			Source:        "TestService",
			CorrelationID: types.MustNewUUID(),
			UserID:        types.NewNullUUID(),
			TraceID:       types.MustNewUUID(),
			Payload:       json.RawMessage(`{"key":"value`), // Invalid JSON
		}
		_, err := NewEvent(input)
		require.Error(t, err)
		assert.EqualError(t, err, "event payload must be valid JSON")
	})
}
