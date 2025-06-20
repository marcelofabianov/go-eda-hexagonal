package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

const (
	LAYOUT_RFC3339_NANO_AUDIT               = time.RFC3339Nano
	LAYOUT_DB_WITH_OFFSET_COLON_AUDIT       = "2006-01-02 15:04:05.999999999-07:00"
	LAYOUT_DB_WITH_OFFSET_NO_COLON_AUDIT    = "2006-01-02 15:04:05.999999999Z0700"
	LAYOUT_DB_SPACE_NANO_OFFSET_SHORT_AUDIT = "2006-01-02 15:04:05.999999999-07"
	LAYOUT_DB_NO_OFFSET_AUDIT               = "2006-01-02 15:04:05.999999999"
	LAYOUT_RFC3339_NO_NANO_AUDIT            = time.RFC3339
	LAYOUT_DB_SIMPLE_AUDIT                  = "2006-01-02 15:04:05"
)

var commonAuditTimeLayouts = []string{
	LAYOUT_RFC3339_NANO_AUDIT,
	LAYOUT_DB_WITH_OFFSET_COLON_AUDIT,
	LAYOUT_DB_WITH_OFFSET_NO_COLON_AUDIT,
	LAYOUT_DB_SPACE_NANO_OFFSET_SHORT_AUDIT,
	LAYOUT_DB_NO_OFFSET_AUDIT,
	LAYOUT_RFC3339_NO_NANO_AUDIT,
	LAYOUT_DB_SIMPLE_AUDIT,
}

func parseAuditTimeMultipleLayouts(timeStr string) (time.Time, error) {
	var lastErr error
	for _, layout := range commonAuditTimeLayouts {
		parsedTime, err := time.Parse(layout, timeStr)
		if err == nil {
			return parsedTime, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("could not parse time '%s' with any known layouts: %w", timeStr, lastErr)
}

type CreatedAt time.Time

func NewCreatedAt() CreatedAt { return CreatedAt(time.Now()) }

func (ca CreatedAt) Time() time.Time { return time.Time(ca) }

func (ca CreatedAt) MarshalJSON() ([]byte, error) { return json.Marshal(time.Time(ca)) }
func (ca *CreatedAt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return msg.NewValidationError(nil,
			map[string]any{"input_json": "null", "target_type": "CreatedAt"},
			"CreatedAt cannot be null (received JSON 'null').",
		)
	}
	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return msg.NewValidationError(err,
			map[string]any{"input_json": string(data), "target_type": "CreatedAt"},
			"CreatedAt must be a valid JSON timestamp.",
		)
	}
	*ca = CreatedAt(t)
	return nil
}
func (ca CreatedAt) Value() (driver.Value, error) { return time.Time(ca), nil }
func (ca *CreatedAt) Scan(src interface{}) error {
	if src == nil {
		return msg.NewValidationError(nil,
			map[string]any{"target_type": "CreatedAt", "received_value": "nil_from_db"},
			"Scanned nil value for non-nullable CreatedAt.",
		)
	}
	var parsedTime time.Time
	var err error
	switch s := src.(type) {
	case time.Time:
		*ca = CreatedAt(s)
		return nil
	case []byte:
		strVal := string(s)
		parsedTime, err = parseAuditTimeMultipleLayouts(strVal)
		if err != nil {
			message := fmt.Sprintf("Failed to convert []byte ('%s') to CreatedAt.", strVal)
			return msg.NewValidationError(err,
				map[string]any{"input_bytes": strVal, "target_type": "CreatedAt"},
				message,
			)
		}
		*ca = CreatedAt(parsedTime)
		return nil
	case string:
		parsedTime, err = parseAuditTimeMultipleLayouts(s)
		if err != nil {
			message := fmt.Sprintf("Failed to convert string ('%s') to CreatedAt.", s)
			return msg.NewValidationError(err,
				map[string]any{"input_string": s, "target_type": "CreatedAt"},
				message,
			)
		}
		*ca = CreatedAt(parsedTime)
		return nil
	default:
		message := fmt.Sprintf("Incompatible type (%T) for CreatedAt.", src)
		return msg.NewValidationError(nil,
			map[string]any{"received_type": fmt.Sprintf("%T", src), "target_type": "CreatedAt"},
			message,
		)
	}
}

type UpdatedAt time.Time

func NewUpdatedAt() UpdatedAt        { return UpdatedAt(time.Now()) }
func (ua UpdatedAt) Time() time.Time { return time.Time(ua) }

func (ua UpdatedAt) MarshalJSON() ([]byte, error) { return json.Marshal(time.Time(ua)) }
func (ua *UpdatedAt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return msg.NewValidationError(nil,
			map[string]any{"input_json": "null", "target_type": "UpdatedAt"},
			"UpdatedAt cannot be null (received JSON 'null').",
		)
	}
	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return msg.NewValidationError(err,
			map[string]any{"input_json": string(data), "target_type": "UpdatedAt"},
			"UpdatedAt must be a valid JSON timestamp.",
		)
	}
	*ua = UpdatedAt(t)
	return nil
}
func (ua UpdatedAt) Value() (driver.Value, error) { return time.Time(ua), nil }
func (ua *UpdatedAt) Scan(src interface{}) error {
	if src == nil {
		return msg.NewValidationError(nil,
			map[string]any{"target_type": "UpdatedAt", "received_value": "nil_from_db"},
			"Scanned nil value for non-nullable UpdatedAt.",
		)
	}
	var parsedTime time.Time
	var err error
	switch s := src.(type) {
	case time.Time:
		*ua = UpdatedAt(s)
		return nil
	case []byte:
		strVal := string(s)
		parsedTime, err = parseAuditTimeMultipleLayouts(strVal)
		if err != nil {
			message := fmt.Sprintf("Failed to convert []byte ('%s') to UpdatedAt.", strVal)
			return msg.NewValidationError(err,
				map[string]any{"input_bytes": strVal, "target_type": "UpdatedAt"},
				message,
			)
		}
		*ua = UpdatedAt(parsedTime)
		return nil
	case string:
		parsedTime, err = parseAuditTimeMultipleLayouts(s)
		if err != nil {
			message := fmt.Sprintf("Failed to convert string ('%s') to UpdatedAt.", s)
			return msg.NewValidationError(err,
				map[string]any{"input_string": s, "target_type": "UpdatedAt"},
				message,
			)
		}
		*ua = UpdatedAt(parsedTime)
		return nil
	default:
		message := fmt.Sprintf("Incompatible type (%T) for UpdatedAt.", src)
		return msg.NewValidationError(nil,
			map[string]any{"received_type": fmt.Sprintf("%T", src), "target_type": "UpdatedAt"},
			message,
		)
	}
}
