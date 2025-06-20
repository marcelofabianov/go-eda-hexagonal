package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewCreatedAt(t *testing.T) {
	ca := types.NewCreatedAt()
	assert.WithinDuration(t, time.Now(), ca.Time(), time.Second, "NewCreatedAt() time should be current")
}

func TestCreatedAt_Time(t *testing.T) {
	now := time.Now()
	ca := types.CreatedAt(now)
	assert.True(t, now.Equal(ca.Time()), "Time() should return the underlying time.Time")
}

func TestCreatedAt_JSONEncoding(t *testing.T) {
	specificTime := time.Date(2024, 5, 22, 10, 30, 0, 123456789, time.UTC)
	ca := types.CreatedAt(specificTime)

	t.Run("Marshal", func(t *testing.T) {
		jsonData, err := json.Marshal(ca)
		require.NoError(t, err, "json.Marshal for CreatedAt should not error")
		expectedJSON := `"` + specificTime.Format(time.RFC3339Nano) + `"`
		assert.JSONEq(t, expectedJSON, string(jsonData), "Marshalled JSON for CreatedAt is not as expected")
	})

	t.Run("Unmarshal valid", func(t *testing.T) {
		jsonInput := `"` + specificTime.Format(time.RFC3339Nano) + `"`
		var newCa types.CreatedAt
		err := json.Unmarshal([]byte(jsonInput), &newCa)
		require.NoError(t, err, "json.Unmarshal for valid CreatedAt JSON should not error")
		assert.True(t, newCa.Time().Equal(specificTime), "Unmarshalled CreatedAt time mismatch")
	})

	t.Run("Unmarshal invalid data", func(t *testing.T) {
		var newCa types.CreatedAt
		err := json.Unmarshal([]byte(`"invalid-date"`), &newCa)
		require.Error(t, err, "json.Unmarshal for invalid date string should error")
		var msgErr *msg.MessageError
		isMsgError := errors.As(err, &msgErr)
		require.True(t, isMsgError, "Error should be a *msg.MessageError for invalid data")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid for invalid data")
	})

	t.Run("Unmarshal null (expect error)", func(t *testing.T) {
		var newCa types.CreatedAt
		err := json.Unmarshal([]byte("null"), &newCa)
		require.Error(t, err, "json.Unmarshal for JSON null should error for CreatedAt")
		var msgErr *msg.MessageError
		isMsgError := errors.As(err, &msgErr)
		require.True(t, isMsgError, "Error should be a *msg.MessageError for JSON null")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid for JSON null")
		assert.Equal(t, "CreatedAt cannot be null (received JSON 'null').", msgErr.Message, "Error message for JSON null mismatch")
	})
}

func TestCreatedAt_DatabaseEncoding(t *testing.T) {
	originalTime := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Value", func(t *testing.T) {
		caVal := types.CreatedAt(originalTime)
		val, err := caVal.Value()
		require.NoError(t, err, "Value() for CreatedAt should not error")
		dbTime, ok := val.(time.Time)
		require.True(t, ok, "DB value should be time.Time")
		assert.True(t, dbTime.Equal(originalTime), "DB value time does not match original")
	})

	testCasesScan := []struct {
		name            string
		src             interface{}
		expectError     bool
		expectedCode    msg.ErrorCode
		expectedMessage string
	}{
		{"Scan time.Time", originalTime, false, "", ""},
		{"Scan string RFC3339Nano", originalTime.Format(time.RFC3339Nano), false, "", ""},
		{"Scan []byte RFC3339Nano", []byte(originalTime.Format(time.RFC3339Nano)), false, "", ""},
		{"Scan string common DB format with offset", originalTime.Format("2006-01-02 15:04:05.999999999-07"), false, "", ""},
		{"Scan string common DB format no offset", originalTime.UTC().Format("2006-01-02 15:04:05.999999999"), false, "", ""},
		{"Scan nil (expect error)", nil, true, msg.CodeInvalid, "Scanned nil value for non-nullable CreatedAt."},
		{"Scan incompatible type (int)", 12345, true, msg.CodeInvalid, "Incompatible type (int) for CreatedAt."},
		{"Scan incompatible type (string)", "not-a-time-string", true, msg.CodeInvalid, "Failed to convert string ('not-a-time-string') to CreatedAt."},
	}

	for _, tc := range testCasesScan {
		t.Run(tc.name, func(t *testing.T) {
			var newCa types.CreatedAt
			err := newCa.Scan(tc.src)
			if tc.expectError {
				require.Error(t, err, fmt.Sprintf("Scan for '%s' should error", tc.name))
				var msgErr *msg.MessageError
				isMsgError := errors.As(err, &msgErr)
				require.True(t, isMsgError, fmt.Sprintf("Error for '%s' should be a *msg.MessageError: %v", tc.name, err))
				assert.Equal(t, tc.expectedCode, msgErr.Code, fmt.Sprintf("Error code mismatch for '%s'", tc.name))
				if tc.expectedMessage != "" {
					assert.Equal(t, tc.expectedMessage, msgErr.Message, fmt.Sprintf("Error message mismatch for '%s'", tc.name))
				}
			} else {
				require.NoError(t, err, fmt.Sprintf("Scan for '%s' should not error", tc.name))
				assert.True(t, newCa.Time().UTC().Equal(originalTime.UTC()),
					fmt.Sprintf("Expected time instant %v (UTC), got %v (UTC) for '%s'", originalTime.UTC(), newCa.Time().UTC(), tc.name))
			}
		})
	}
}

func TestNewUpdatedAt(t *testing.T) {
	ua := types.NewUpdatedAt()
	assert.WithinDuration(t, time.Now(), ua.Time(), time.Second, "NewUpdatedAt() time should be current")
}

func TestUpdatedAt_Time(t *testing.T) {
	now := time.Now()
	ua := types.UpdatedAt(now)
	assert.True(t, now.Equal(ua.Time()), "Time() should return the underlying time.Time")
}

func TestUpdatedAt_JSONEncoding(t *testing.T) {
	specificTime := time.Date(2024, 5, 23, 11, 0, 0, 987654321, time.UTC)
	ua := types.UpdatedAt(specificTime)

	t.Run("Marshal", func(t *testing.T) {
		jsonData, err := json.Marshal(ua)
		require.NoError(t, err, "json.Marshal for UpdatedAt should not error")
		expectedJSON := `"` + specificTime.Format(time.RFC3339Nano) + `"`
		assert.JSONEq(t, expectedJSON, string(jsonData), "Marshalled JSON for UpdatedAt is not as expected")
	})

	t.Run("Unmarshal valid", func(t *testing.T) {
		jsonInput := `"` + specificTime.Format(time.RFC3339Nano) + `"`
		var newUa types.UpdatedAt
		err := json.Unmarshal([]byte(jsonInput), &newUa)
		require.NoError(t, err, "json.Unmarshal for valid UpdatedAt JSON should not error")
		assert.True(t, newUa.Time().Equal(specificTime), "Unmarshalled UpdatedAt time mismatch")
	})

	t.Run("Unmarshal invalid data", func(t *testing.T) {
		var newUa types.UpdatedAt
		err := json.Unmarshal([]byte(`"invalid-data"`), &newUa)
		require.Error(t, err, "json.Unmarshal for invalid date string should error")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be a *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
	})

	t.Run("Unmarshal null (expect error)", func(t *testing.T) {
		var newUa types.UpdatedAt
		err := json.Unmarshal([]byte("null"), &newUa)
		require.Error(t, err, "json.Unmarshal for JSON null should error for UpdatedAt")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be a *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
		assert.Equal(t, "UpdatedAt cannot be null (received JSON 'null').", msgErr.Message, "Error message for JSON null mismatch")
	})
}

func TestUpdatedAt_DatabaseEncoding(t *testing.T) {
	originalTime := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Value", func(t *testing.T) {
		uaVal := types.UpdatedAt(originalTime)
		val, err := uaVal.Value()
		require.NoError(t, err, "Value() for UpdatedAt should not error")
		dbTime, ok := val.(time.Time)
		require.True(t, ok, "DB value should be time.Time")
		assert.True(t, dbTime.Equal(originalTime), "DB value time does not match original")
	})

	testCasesScan := []struct {
		name            string
		src             interface{}
		expectError     bool
		expectedCode    msg.ErrorCode
		expectedMessage string
	}{
		{"Scan time.Time", originalTime, false, "", ""},
		{"Scan string RFC3339Nano", originalTime.Format(time.RFC3339Nano), false, "", ""},
		{"Scan []byte RFC3339Nano", []byte(originalTime.Format(time.RFC3339Nano)), false, "", ""},
		{"Scan string common DB format with offset", originalTime.Format("2006-01-02 15:04:05.999999999-07"), false, "", ""},
		{"Scan string common DB format no offset", originalTime.UTC().Format("2006-01-02 15:04:05.999999999"), false, "", ""},
		{"Scan nil (expect error)", nil, true, msg.CodeInvalid, "Scanned nil value for non-nullable UpdatedAt."},
		{"Scan incompatible type (int)", 12345, true, msg.CodeInvalid, "Incompatible type (int) for UpdatedAt."},
		{"Scan incompatible type (string)", "not-a-time-string", true, msg.CodeInvalid, "Failed to convert string ('not-a-time-string') to UpdatedAt."},
	}

	for _, tc := range testCasesScan {
		t.Run(tc.name, func(t *testing.T) {
			var newUa types.UpdatedAt
			err := newUa.Scan(tc.src)
			if tc.expectError {
				require.Error(t, err, fmt.Sprintf("Scan for '%s' should error", tc.name))
				var msgErr *msg.MessageError
				isMsgError := errors.As(err, &msgErr)
				require.True(t, isMsgError, fmt.Sprintf("Error for '%s' should be a *msg.MessageError: %v", tc.name, err))
				assert.Equal(t, tc.expectedCode, msgErr.Code, fmt.Sprintf("Error code mismatch for '%s'", tc.name))
				if tc.expectedMessage != "" {
					assert.Equal(t, tc.expectedMessage, msgErr.Message, fmt.Sprintf("Error message mismatch for '%s'", tc.name))
				}
			} else {
				require.NoError(t, err, fmt.Sprintf("Scan for '%s' should not error", tc.name))
				assert.True(t, newUa.Time().UTC().Equal(originalTime.UTC()),
					fmt.Sprintf("Expected time instant %v (UTC), got %v (UTC) for '%s'", originalTime.UTC(), newUa.Time().UTC(), tc.name))
			}
		})
	}
}
