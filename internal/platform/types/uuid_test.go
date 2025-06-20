package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewUUID(t *testing.T) {
	id, err := types.NewUUID()
	require.NoError(t, err, "NewUUID() should not return an error on success")
	assert.False(t, id.IsNil(), "NewUUID() should not return a nil UUID")
	assert.Equal(t, uuid.Version(7), uuid.UUID(id).Version(), "NewUUID() should generate a Version 7 UUID")
	_, parseErr := uuid.Parse(id.String())
	assert.NoError(t, parseErr, "NewUUID() should produce a valid UUID string")
}

func TestMustNewUUID(t *testing.T) {
	var id types.UUID
	assert.NotPanics(t, func() { id = types.MustNewUUID() }, "MustNewUUID() should not panic on success")
	assert.False(t, id.IsNil(), "MustNewUUID() result should not be nil")
	assert.Equal(t, uuid.Version(7), uuid.UUID(id).Version(), "MustNewUUID() should generate a Version 7 UUID")
}

func TestParseUUID(t *testing.T) {
	validV7GoogleUUID := uuid.Must(uuid.NewV7())
	validV7Str := validV7GoogleUUID.String()

	testCases := []struct {
		name            string
		input           string
		want            types.UUID
		wantErr         bool
		expectedErrCode msg.ErrorCode
		errContains     string
	}{
		{"valid uuid", validV7Str, types.UUID(validV7GoogleUUID), false, "", ""},
		{"nil uuid string", "00000000-0000-0000-0000-000000000000", types.Nil, false, "", ""},
		{"invalid uuid", "not-a-uuid", types.Nil, true, msg.CodeInvalid, "Invalid UUID string format: 'not-a-uuid'"},
		{"empty string", "", types.Nil, true, msg.CodeInvalid, "Invalid UUID string format: ''"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := types.ParseUUID(tc.input)
			if tc.wantErr {
				require.Error(t, err, fmt.Sprintf("ParseUUID('%s') should return an error", tc.input))
				var msgErr *msg.MessageError
				isMsgError := errors.As(err, &msgErr)
				require.True(t, isMsgError, "Error should be of type *msg.MessageError")
				assert.Equal(t, tc.expectedErrCode, msgErr.Code, "Error code mismatch")
				if tc.errContains != "" {
					assert.Contains(t, msgErr.Message, tc.errContains, "Error message content mismatch")
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("ParseUUID('%s') should not return an error", tc.input))
				assert.Equal(t, tc.want, got, "Parsed UUID mismatch")
				if tc.input == "00000000-0000-0000-0000-000000000000" {
					assert.True(t, got.IsNil(), "Parsed nil UUID string should result in types.Nil")
				}
			}
		})
	}
}

func TestMustParseUUID(t *testing.T) {
	validV7Str := uuid.Must(uuid.NewV7()).String()

	t.Run("valid uuid, no panic", func(t *testing.T) {
		var id types.UUID
		assert.NotPanics(t, func() { id = types.MustParseUUID(validV7Str) }, "MustParseUUID should not panic for valid UUID")
		assert.Equal(t, validV7Str, id.String(), "Parsed UUID string mismatch")
	})

	t.Run("invalid uuid, panics with MessageError", func(t *testing.T) {
		invalidInput := "this-will-panic"
		expectedPanicMsgPart := fmt.Sprintf("Invalid UUID string format: '%s'", invalidInput)

		defer func() {
			r := recover()
			require.NotNil(t, r, "MustParseUUID should panic for invalid UUID")
			err, ok := r.(error)
			require.True(t, ok, "Panic value should be an error")
			var msgErr *msg.MessageError
			isMsgError := errors.As(err, &msgErr)
			require.True(t, isMsgError, "Panic error should be of type *msg.MessageError")
			assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Panic error code mismatch")
			assert.Contains(t, msgErr.Message, expectedPanicMsgPart, "Panic error message content mismatch")
		}()
		_ = types.MustParseUUID(invalidInput)
	})
}

func TestUUID_StringAndIsNil(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		id := mustNewTestUUID(t)
		str := id.String()
		_, err := uuid.Parse(str)
		require.NoError(t, err, "String() should return a valid UUID format")
	})

	t.Run("IsNil", func(t *testing.T) {
		assert.True(t, types.Nil.IsNil(), "types.Nil.IsNil() should be true")
		id := mustNewTestUUID(t)
		assert.False(t, id.IsNil(), "A non-nil UUID IsNil() should be false")
	})
}

func TestUUID_TextEncoding(t *testing.T) {
	originalID := mustNewTestUUID(t)

	t.Run("MarshalText and UnmarshalText roundtrip", func(t *testing.T) {
		text, err := originalID.MarshalText()
		require.NoError(t, err, "MarshalText should not return an error")

		var newID types.UUID
		err = newID.UnmarshalText(text)
		require.NoError(t, err, "UnmarshalText should not return an error for valid text")
		assert.Equal(t, originalID, newID, "ID after MarshalText/UnmarshalText roundtrip should be equal to original")
	})

	t.Run("UnmarshalText with invalid text", func(t *testing.T) {
		var id types.UUID
		invalidText := []byte("invalid-text-for-uuid")
		err := id.UnmarshalText(invalidText)
		require.Error(t, err, "UnmarshalText should return an error for invalid text")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be of type *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
		assert.Contains(t, msgErr.Message, "Invalid text representation for UUID", "Error message content mismatch")
	})
}

func TestUUID_JSONEncoding(t *testing.T) {
	originalID := mustNewTestUUID(t)
	validNu := types.NewValidNullableUUID(originalID)

	t.Run("Marshal and Unmarshal roundtrip via JSON", func(t *testing.T) {
		jsonData, err := json.Marshal(validNu)
		require.NoError(t, err, "json.Marshal for NullableUUID(valid) should not error")

		expectedJSON := fmt.Sprintf("\"%s\"", originalID.String())
		assert.JSONEq(t, expectedJSON, string(jsonData), "Marshalled JSON string mismatch")

		var newID types.UUID
		err = json.Unmarshal(jsonData, &newID)
		require.NoError(t, err, "json.Unmarshal for valid UUID JSON string should not error")
		assert.Equal(t, originalID, newID, "ID after JSON roundtrip should match original")
	})

	t.Run("UnmarshalJSON with invalid JSON UUID string", func(t *testing.T) {
		var id types.UUID
		err := json.Unmarshal([]byte("\"invalid-uuid-string\""), &id)
		require.Error(t, err, "json.Unmarshal for invalid UUID string should error")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be of type *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
	})
}

func TestUUID_DatabaseEncoding(t *testing.T) {
	originalID := mustNewTestUUID(t)

	t.Run("Value and Scan roundtrip", func(t *testing.T) {
		value, err := originalID.Value()
		require.NoError(t, err, "Value() should not return an error")
		if !originalID.IsNil() {
			assert.NotNil(t, value, "Value() for non-nil UUID should not be nil")
		}

		var newID types.UUID
		err = newID.Scan(value)
		require.NoError(t, err, "Scan() should not return an error for valid value")
		assert.Equal(t, originalID, newID, "ID after Value/Scan roundtrip should match original")
	})

	t.Run("Scan nil value", func(t *testing.T) {
		var id types.UUID
		err := id.Scan(nil)
		require.NoError(t, err, "Scan(nil) should not error (it sets to Nil UUID via google/uuid.Scan)")
		assert.True(t, id.IsNil(), "Scan(nil) should result in a Nil UUID")
	})

	t.Run("Value for Nil UUID", func(t *testing.T) {
		value, err := types.Nil.Value()
		require.NoError(t, err, "Value() for types.Nil should not error")
		assert.Nil(t, value, "Value() for types.Nil should return nil driver.Value")
	})

	t.Run("Scan incompatible type", func(t *testing.T) {
		var id types.UUID
		err := id.Scan(12345)
		require.Error(t, err, "Scan() should return an error for incompatible type")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be of type *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid for incompatible type")
		assert.Contains(t, msgErr.Message, "Failed to scan database value", "Error message content mismatch")
	})
}
