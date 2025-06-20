package types

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

type NullableUUID struct {
	uuid.NullUUID
}

func NewNullableUUID(id UUID, valid bool) NullableUUID {
	return NullableUUID{uuid.NullUUID{UUID: uuid.UUID(id), Valid: valid}}
}

func NewValidNullableUUID(id UUID) NullableUUID {
	return NullableUUID{uuid.NullUUID{UUID: uuid.UUID(id), Valid: true}}
}

func NewNullUUID() NullableUUID {
	return NullableUUID{uuid.NullUUID{UUID: uuid.Nil, Valid: false}}
}

func (nu NullableUUID) IsValid() bool {
	return nu.Valid && nu.UUID != uuid.Nil
}

func (nu NullableUUID) GetUUID() (UUID, bool) {
	if !nu.Valid {
		return Nil, false
	}
	return UUID(nu.UUID), true
}

func (nu NullableUUID) MarshalJSON() ([]byte, error) {
	if !nu.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(UUID(nu.UUID))
}

func (nu *NullableUUID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nu.Valid = false
		nu.UUID = uuid.Nil
		return nil
	}
	var tempUUID UUID
	if err := json.Unmarshal(data, &tempUUID); err != nil {
		nu.Valid = false
		message := fmt.Sprintf("NullableUUID must be a valid JSON UUID string or 'null'; received '%s'.", string(data))
		return msg.NewValidationError(err,
			map[string]any{"input_json": string(data), "target_type": "NullableUUID"},
			message,
		)
	}
	nu.UUID = uuid.UUID(tempUUID)
	nu.Valid = true
	return nil
}
