package user

import (
	"encoding/json"
	"fmt"

	"github.com/marcelofabianov/redtogreen/internal/platform/event"
	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

const (
	UserCreatedEventType    event.EventType    = "user.created"
	UserCreatedEventVersion event.EventVersion = "v1.0.0"
	UserEventSource         string             = "IdentityService"
)

type USerCreatedEventInput struct {
	CorrelationID   types.UUID
	UserID          types.NullableUUID // Author
	TraceID         types.UUID         // OTEL
	PreviousEventID types.NullableUUID
	CausationID     types.NullableUUID
	Payload         UserCreatedPayload
}

type UserCreatedPayload struct {
	UserID types.UUID `json:"userId"`
	Name   string     `json:"name"`
	Email  string     `json:"email"`
	Phone  string     `json:"phone"`
}

func NewUserCreatedEvent(input USerCreatedEventInput) (*event.Event, error) {
	payloadBytes, err := json.Marshal(input.Payload)
	if err != nil {
		return nil, msg.NewInternalError(err, map[string]any{"payload": fmt.Sprintf("%+v", input.Payload)})
	}

	evt, err := event.NewEvent(event.EventInput{
		EventType:       UserCreatedEventType,
		EventVersion:    UserCreatedEventVersion,
		Source:          UserEventSource,
		CorrelationID:   input.CorrelationID,
		UserID:          input.UserID,
		TraceID:         input.TraceID,
		PreviousEventID: input.PreviousEventID,
		CausationID:     input.CausationID,
		Payload:         payloadBytes,
	})
	if err != nil {
		return nil, err
	}

	return &evt, nil
}
