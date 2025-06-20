package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

type Version int

func NewVersion() Version {
	return Version(1)
}

func (v *Version) Increment() {
	*v++
}

func (v Version) Int() int {
	return int(v)
}

func (v Version) Previous() Version {
	if v <= 1 {
		return Version(0)
	}
	return v - 1
}

func (v Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(v))
}

func (v *Version) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return msg.NewValidationError(nil,
			map[string]any{"input_json": "null", "target_type": "Version"},
			"Version cannot be null (received JSON 'null').",
		)
	}
	var i int
	if err := json.Unmarshal(data, &i); err != nil {
		return msg.NewValidationError(err,
			map[string]any{"input_json": string(data), "target_type": "Version"},
			"Version must be a JSON number.",
		)
	}
	*v = Version(i)
	return nil
}

func (v Version) Value() (driver.Value, error) {
	return int64(v), nil
}

func (v *Version) Scan(src interface{}) error {
	if src == nil {
		return msg.NewValidationError(nil,
			map[string]any{"target_type": "Version"},
			"Scanned nil value for non-nullable Version.",
		)
	}
	var intVal int64
	switch s := src.(type) {
	case int64:
		const maxInt = int(^uint(0) >> 1)
		const minInt = -maxInt - 1
		if s > int64(maxInt) || s < int64(minInt) {
			message := fmt.Sprintf("Value %d from database is out of range for Version (int).", s)
			return msg.NewValidationError(nil,
				map[string]any{"source_value": s},
				message,
			)
		}
		intVal = s
	case []byte:
		parsed, err := strconv.ParseInt(string(s), 10, 32)
		if err != nil {
			message := fmt.Sprintf("Failed to convert []byte ('%s') to int for Version.", string(s))
			return msg.NewValidationError(err,
				map[string]any{"input_bytes": string(s)},
				message,
			)
		}
		intVal = parsed
	default:
		message := fmt.Sprintf("Incompatible type (%T) for Version. Expected int64 or []byte.", src)
		return msg.NewValidationError(nil,
			map[string]any{"received_type": fmt.Sprintf("%T", src)},
			message,
		)
	}
	*v = Version(intVal)
	return nil
}
