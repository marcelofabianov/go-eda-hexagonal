package types

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

type UUID uuid.UUID

var Nil UUID

func NewUUID() (UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Nil, msg.NewInternalError(err, map[string]any{"operation": "generate_v7_uuid"})
	}
	return UUID(id), nil
}

func MustNewUUID() UUID {
	id, err := NewUUID()
	if err != nil {
		panic(err)
	}
	return id
}

func ParseUUID(s string) (UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		message := fmt.Sprintf("Invalid UUID string format: '%s'.", s)
		return Nil, msg.NewValidationError(err,
			map[string]any{"input_string": s},
			message,
		)
	}
	return UUID(id), nil
}

func MustParseUUID(s string) UUID {
	id, err := ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

func (u UUID) IsNil() bool {
	return uuid.UUID(u) == uuid.Nil
}

func (u UUID) MarshalText() ([]byte, error) {
	return uuid.UUID(u).MarshalText()
}

func (u *UUID) UnmarshalText(text []byte) error {
	var underlyingUUID uuid.UUID
	if err := underlyingUUID.UnmarshalText(text); err != nil {
		message := fmt.Sprintf("Invalid text representation for UUID: '%s'.", string(text))
		return msg.NewValidationError(err,
			map[string]any{"input_text": string(text)},
			message,
		)
	}
	*u = UUID(underlyingUUID)
	return nil
}

func (u UUID) Value() (driver.Value, error) {
	if u.IsNil() {
		return nil, nil
	}
	return uuid.UUID(u).Value()
}

func (u *UUID) Scan(src interface{}) error {
	var underlyingUUID uuid.UUID
	if err := underlyingUUID.Scan(src); err != nil {
		message := fmt.Sprintf("Failed to scan database value of type %T into UUID.", src)
		return msg.NewValidationError(err,
			map[string]any{"source_type": fmt.Sprintf("%T", src)},
			message,
		)
	}
	*u = UUID(underlyingUUID)
	return nil
}
