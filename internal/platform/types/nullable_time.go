package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

type NullableTime struct {
	sql.NullTime
}

func NewNullableTime(t time.Time, valid bool) NullableTime {
	return NullableTime{sql.NullTime{Time: t, Valid: valid}}
}

func NewValidNullableTime(t time.Time) NullableTime {
	return NullableTime{sql.NullTime{Time: t, Valid: true}}
}

func NewNullTime() NullableTime {
	return NullableTime{sql.NullTime{Time: time.Time{}, Valid: false}}
}

func (nt NullableTime) TimeOrZero() time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func (nt *NullableTime) Set(t time.Time) {
	nt.Time = t
	nt.Valid = !t.IsZero()
}

func (nt *NullableTime) SetNull() {
	nt.Time = time.Time{}
	nt.Valid = false
}

func (nt NullableTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(nt.Time)
}

func (nt *NullableTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		nt.Time = time.Time{}
		return nil
	}
	var tempTime time.Time
	if err := json.Unmarshal(data, &tempTime); err != nil {
		nt.Valid = false
		message := fmt.Sprintf("NullableTime must be a valid JSON timestamp or 'null'; received '%s'.", string(data))
		return msg.NewValidationError(err,
			map[string]any{"input_json": string(data), "target_type": "NullableTime"},
			message,
		)
	}
	nt.Time = tempTime
	nt.Valid = true
	return nil
}

func (nt NullableTime) IsNullable() bool {
	return !nt.Valid
}

type DeletedAt struct {
	NullableTime
}

func NewDeletedAtNow() DeletedAt {
	return DeletedAt{NullableTime: NewNullableTime(time.Now(), true)}
}

func NewNilDeletedAt() DeletedAt {
	return DeletedAt{NullableTime: NewNullableTime(time.Time{}, false)}
}

func (da *DeletedAt) SetNow() {
	da.Time = time.Now()
	da.Valid = true
}

func (da *DeletedAt) IsNullable() bool {
	return !da.Valid
}

type ArchivedAt struct {
	NullableTime
}

func NewArchivedAtNow() ArchivedAt {
	return ArchivedAt{NullableTime: NewNullableTime(time.Now(), true)}
}

func NewNilArchivedAt() ArchivedAt {
	return ArchivedAt{NullableTime: NewNullableTime(time.Time{}, false)}
}

func (aa *ArchivedAt) SetNow() {
	aa.Time = time.Now()
	aa.Valid = true
}

func (aa *ArchivedAt) IsNullable() bool {
	return !aa.Valid
}
