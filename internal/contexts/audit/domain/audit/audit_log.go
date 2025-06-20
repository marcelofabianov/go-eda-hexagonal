package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/marcelofabianov/redtogreen/internal/platform/event"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

type NewAuditLogInput struct {
	EventID      types.UUID
	EventType    event.EventType
	EventVersion event.EventVersion
	EventContext event.EventContext
	TraceID      types.UUID
	UserAuthorID types.NullableUUID
	Payload      json.RawMessage
}

type AuditLog struct {
	ID           types.UUID         `db:"id"`
	EventID      types.UUID         `db:"event_id"`
	EventType    event.EventType    `db:"event_type"`
	EventVersion event.EventVersion `db:"event_version"`
	EventContext json.RawMessage    `db:"event_context"`
	TraceID      types.UUID         `db:"trace_id"`
	UserAuthorID types.NullableUUID `db:"user_author_id"`
	Payload      json.RawMessage    `db:"payload"`
	CreatedAt    time.Time          `db:"created_at"`
}

func NewAuditLog(input NewAuditLogInput) (*AuditLog, error) {
	if input.EventID.IsNil() {
		return nil, errors.New("audit log validation: event ID cannot be nil")
	}
	if input.EventType == "" {
		return nil, errors.New("audit log validation: event type cannot be empty")
	}
	if input.TraceID.IsNil() {
		return nil, errors.New("audit log validation: trace id cannot be nil")
	}

	id, err := types.NewUUID()
	if err != nil {
		return nil, err
	}

	eventContextBytes, err := json.Marshal(input.EventContext)
	if err != nil {
		return nil, fmt.Errorf("audit log creation: failed to marshal event context: %w", err)
	}

	return &AuditLog{
		ID:           id,
		EventID:      input.EventID,
		EventType:    input.EventType,
		EventVersion: input.EventVersion,
		EventContext: eventContextBytes,
		TraceID:      input.TraceID,
		UserAuthorID: input.UserAuthorID,
		Payload:      input.Payload,
		CreatedAt:    time.Now().UTC(),
	}, nil
}
