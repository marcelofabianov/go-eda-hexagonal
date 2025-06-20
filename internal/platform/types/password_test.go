package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewPassword(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Success: valid password that meets all criteria",
			input:       "ValidPass123!",
			expectError: false,
		},
		{
			name:          "Failure: password is too short",
			input:         "Vp1!",
			expectError:   true,
			errorContains: "must be at least 10 characters long",
		},
		{
			name:          "Failure: password has no number",
			input:         "ValidPassword!",
			expectError:   true,
			errorContains: "must contain at least one numeric character",
		},
		{
			name:          "Failure: password has no uppercase letter",
			input:         "validpass123!",
			expectError:   true,
			errorContains: "must contain at least one uppercase letter",
		},
		{
			name:          "Failure: password has no lowercase letter",
			input:         "VALIDPASS123!",
			expectError:   true,
			errorContains: "must contain at least one lowercase letter",
		},
		{
			name:          "Failure: password has no symbol",
			input:         "ValidPass123",
			expectError:   true,
			errorContains: "must contain at least one symbol",
		},
		{
			name:          "Failure: password is empty",
			input:         "",
			expectError:   true,
			errorContains: "Password cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := types.NewPassword(tc.input)

			if tc.expectError {
				require.Error(t, err, "Expected an error for input '%s', but got none", tc.input)
				var msgErr *msg.MessageError
				require.ErrorAs(t, err, &msgErr, "Error should be a *msg.MessageError")
				assert.Equal(t, msg.CodeInvalid, msgErr.Code)
				assert.Contains(t, msgErr.Message, tc.errorContains, "Error message should contain the expected validation rule violation")
			} else {
				require.NoError(t, err, "Expected no error for valid input '%s', but got one", tc.input)
				assert.Equal(t, tc.input, p.String(), "The created password should match the input string")
				assert.False(t, p.IsEmpty(), "A valid password should not be considered empty")
			}
		})
	}
}

func TestMustNewPassword(t *testing.T) {
	t.Run("Success: should create password without panic", func(t *testing.T) {
		validPassword := "ValidPass123!"
		assert.NotPanics(t, func() {
			p := types.MustNewPassword(validPassword)
			assert.Equal(t, validPassword, p.String(), "Password string representation should match the input")
		}, "MustNewPassword should not panic for a valid password")
	})

	t.Run("Failure: should panic for invalid password", func(t *testing.T) {
		invalidPassword := "invalid"
		assert.Panics(t, func() {
			types.MustNewPassword(invalidPassword)
		}, "MustNewPassword should panic when given an invalid password")
	})
}

func TestPassword_UnmarshalJSON(t *testing.T) {
	t.Run("Success: unmarshal valid password from JSON", func(t *testing.T) {
		var p types.Password
		validJSON := []byte(`"ValidPass123!"`)
		err := json.Unmarshal(validJSON, &p)

		require.NoError(t, err, "Unmarshalling a valid password from JSON should not produce an error")
		assert.Equal(t, "ValidPass123!", p.String(), "Password value should be correct after unmarshalling")
	})

	t.Run("Failure: unmarshal password that fails validation", func(t *testing.T) {
		var p types.Password
		invalidPasswordJSON := []byte(`"short"`)
		err := json.Unmarshal(invalidPasswordJSON, &p)

		var msgErr *msg.MessageError
		require.ErrorAs(t, err, &msgErr, "Expected a validation error of type *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code)
		assert.Contains(t, msgErr.Message, "must be at least 10 characters long")
	})

	t.Run("Failure: unmarshal JSON of the wrong type", func(t *testing.T) {
		var p types.Password

		invalidJSON := []byte(`123`)
		err := json.Unmarshal(invalidJSON, &p)

		var msgErr *msg.MessageError
		require.ErrorAs(t, err, &msgErr, "Expected an error of type *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code)
		assert.Equal(t, "Password must be a valid JSON string.", msgErr.Message)

		var unmarshalTypeErr *json.UnmarshalTypeError
		assert.ErrorAs(t, err, &unmarshalTypeErr, "The error chain should contain the original json.UnmarshalTypeError")
	})
}
