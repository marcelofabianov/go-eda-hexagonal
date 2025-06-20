package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

type EventType string

type EventVersion string

type EventHeader struct {
	EventID       types.UUID   `json:"eventId"`
	EventType     EventType    `json:"eventType"`
	Timestamp     time.Time    `json:"timestamp"`
	Source        string       `json:"source"`
	SchemaVersion EventVersion `json:"jsonSchemaVersion"`
}

func NewEventHeader(eventType EventType, eventVersion EventVersion, source string) (EventHeader, error) {
	eventID, err := types.NewUUID()
	if err != nil {
		return EventHeader{}, fmt.Errorf("failed to generate event header id: %w", err)
	}

	return EventHeader{
		EventID:       eventID,
		EventType:     eventType,
		Timestamp:     time.Now().UTC(),
		Source:        source,
		SchemaVersion: eventVersion,
	}, nil
}

type EventContext struct {
	CorrelationID types.UUID         `json:"correlationId"`
	UserID        types.NullableUUID `json:"userId,omitempty"`
}

func NewEventContext(correlationID types.UUID, userID types.NullableUUID) EventContext {
	return EventContext{
		CorrelationID: correlationID,
		UserID:        userID,
	}
}

type EventMetadata struct {
	TraceID         types.UUID         `json:"traceId"`
	PreviousEventID types.NullableUUID `json:"previousEventId,omitempty"`
	CausationID     types.NullableUUID `json:"causationId,omitempty"`
}

func NewEventMetadata(traceID types.UUID, previousEventID types.NullableUUID, causationID types.NullableUUID) EventMetadata {
	return EventMetadata{
		TraceID:         traceID,
		PreviousEventID: previousEventID,
		CausationID:     causationID,
	}
}

type EventInput struct {
	EventType       EventType
	EventVersion    EventVersion
	Source          string
	CorrelationID   types.UUID
	UserID          types.NullableUUID
	TraceID         types.UUID
	PreviousEventID types.NullableUUID
	CausationID     types.NullableUUID
	Payload         json.RawMessage
}

type Event struct {
	Header   EventHeader     `json:"header"`
	Context  EventContext    `json:"context"`
	Metadata EventMetadata   `json:"metadata"`
	Payload  json.RawMessage `json:"payload"`
}

func NewEvent(input EventInput) (Event, error) {
	if input.EventType == "" {
		return Event{}, errors.New("event type cannot be empty")
	}
	if input.Source == "" {
		return Event{}, errors.New("event source cannot be empty")
	}
	if input.EventVersion == "" {
		return Event{}, errors.New("event version cannot be empty")
	}
	if input.CorrelationID.IsNil() {
		return Event{}, errors.New("correlation ID cannot be empty")
	}
	if input.TraceID.IsNil() {
		return Event{}, errors.New("trace ID cannot be empty")
	}
	if input.Payload == nil {
		return Event{}, errors.New("event payload cannot be empty")
	}
	if !json.Valid(input.Payload) {
		return Event{}, errors.New("event payload must be valid JSON")
	}

	header, err := NewEventHeader(input.EventType, input.EventVersion, input.Source)
	if err != nil {
		return Event{}, fmt.Errorf("failed to create event: %w", err)
	}

	context := NewEventContext(input.CorrelationID, input.UserID)
	metadata := NewEventMetadata(input.TraceID, input.PreviousEventID, input.CausationID)

	return Event{
		Header:   header,
		Context:  context,
		Metadata: metadata,
		Payload:  input.Payload,
	}, nil
}
