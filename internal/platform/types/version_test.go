package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewVersion(t *testing.T) {
	v := types.NewVersion()
	assert.Equal(t, types.Version(1), v, "NewVersion() should initialize with 1")
	assert.Equal(t, 1, v.Int(), "Int() of NewVersion() should be 1")
}

func TestVersion_Increment(t *testing.T) {
	v := types.Version(5)
	v.Increment()
	assert.Equal(t, types.Version(6), v, "Increment() should increase the version by 1")

	v = types.NewVersion()
	v.Increment()
	assert.Equal(t, types.Version(2), v, "Increment() after NewVersion() should result in 2")
}

func TestVersion_Int(t *testing.T) {
	v := types.Version(42)
	assert.Equal(t, 42, v.Int(), "Int() should return the correct int value")
}

func TestVersion_Previous(t *testing.T) {
	testCases := []struct {
		name           string
		initialVersion types.Version
		expectedPrev   types.Version
	}{
		{"version greater than 1", types.Version(5), types.Version(4)},
		{"version is 1", types.Version(1), types.Version(0)},
		{"version is 0", types.Version(0), types.Version(0)},
		{"version is negative", types.Version(-5), types.Version(0)},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			previous := tc.initialVersion.Previous()
			assert.Equal(t, tc.expectedPrev, previous, fmt.Sprintf("Previous() of %d should be %d", tc.initialVersion.Int(), tc.expectedPrev.Int()))
		})
	}

	t.Run("should not alter original version", func(t *testing.T) {
		originalVersion := types.Version(3)
		_ = originalVersion.Previous()
		assert.Equal(t, types.Version(3), originalVersion, "Previous() should not modify the original version")
	})
}

func TestVersion_JSONEncoding(t *testing.T) {
	originalVersion := types.Version(123)

	t.Run("MarshalJSON", func(t *testing.T) {
		jsonData, err := json.Marshal(originalVersion)
		require.NoError(t, err, "json.Marshal should not return an error")
		expectedJSON := fmt.Sprintf("%d", originalVersion.Int())
		assert.JSONEq(t, expectedJSON, string(jsonData), "Marshalled JSON should be the expected number string")
	})

	t.Run("UnmarshalJSON valid", func(t *testing.T) {
		var newVersion types.Version
		jsonData := []byte(fmt.Sprintf("%d", originalVersion.Int()))
		err := json.Unmarshal(jsonData, &newVersion)
		require.NoError(t, err, "json.Unmarshal should not return error for valid numeric JSON")
		assert.Equal(t, originalVersion, newVersion, "Version after Marshal/Unmarshal should be equal to original")
	})

	t.Run("UnmarshalJSON non-numeric json", func(t *testing.T) {
		var v types.Version
		err := json.Unmarshal([]byte("\"abc\""), &v)
		require.Error(t, err, "json.Unmarshal should return error for non-numeric JSON")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be of type *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
		assert.Equal(t, "Version must be a JSON number.", msgErr.Message, "Error message mismatch")
	})

	t.Run("UnmarshalJSON null json", func(t *testing.T) {
		var v types.Version
		err := json.Unmarshal([]byte("null"), &v)
		require.Error(t, err, "json.Unmarshal of 'null' into Version should now error due to explicit check")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error for 'null' JSON should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code for 'null' JSON should be CodeInvalid")
		assert.Equal(t, "Version cannot be null (received JSON 'null').", msgErr.Message, "Error message for 'null' JSON mismatch")
	})
}

func TestVersion_DatabaseEncoding(t *testing.T) {
	originalVersion := types.Version(77)

	t.Run("Value and Scan roundtrip (int64)", func(t *testing.T) {
		value, err := originalVersion.Value()
		require.NoError(t, err, "Value() should not return an error")
		assert.Equal(t, int64(77), value, "Value() should return an int64")

		var newVersion types.Version
		err = newVersion.Scan(value)
		require.NoError(t, err, "Scan() with int64 should not return an error")
		assert.Equal(t, originalVersion, newVersion, "Version after Value/Scan (int64) should be equal to original")
	})

	testCasesScan := []struct {
		name            string
		src             interface{}
		expectError     bool
		expectedCode    msg.ErrorCode
		expectedMessage string
	}{
		{"Scan []byte numeric string", []byte("88"), false, "", ""},
		{"Scan nil value (expect error)", nil, true, msg.CodeInvalid, "Scanned nil value for non-nullable Version."},
		{"Scan incompatible type (string)", "not a number string", true, msg.CodeInvalid, "Incompatible type (string) for Version. Expected int64 or []byte."},
		{"Scan non-numeric []byte", []byte("abc"), true, msg.CodeInvalid, "Failed to convert []byte ('abc') to int for Version."},
		{"Scan int64 value out of int range (overflow)", int64(math.MaxInt32 + 10), (int(int64(math.MaxInt32+10)) != math.MaxInt32+10), msg.CodeInvalid, fmt.Sprintf("Value %d from database is out of range for Version (int).", int64(math.MaxInt32+10))},
	}

	for _, tc := range testCasesScan {
		t.Run(tc.name, func(t *testing.T) {
			var v types.Version
			err := v.Scan(tc.src)
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
				assert.NoError(t, err, fmt.Sprintf("Scan for '%s' should not error", tc.name))
				if s, ok := tc.src.(int64); ok {
					assert.Equal(t, types.Version(s), v)
				} else if b, ok := tc.src.([]byte); ok {
					expectedInt, _ := strconv.Atoi(string(b))
					assert.Equal(t, types.Version(expectedInt), v)
				}
			}
		})
	}
}
