package types_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewPhone(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      types.Phone
		expectError   bool
		expectedCode  msg.ErrorCode
		errorContains string
	}{
		{
			name:        "Valid phone with country code and DDD",
			input:       "5562982870053",
			expected:    types.Phone("5562982870053"),
			expectError: false,
		},
		{
			name:        "Valid phone with DDD and 9 digits (should add 55)",
			input:       "62982870053",
			expected:    types.Phone("5562982870053"),
			expectError: false,
		},
		{
			name:        "Valid phone with formatting (spaces and parentheses)",
			input:       "(62) 98287-0053",
			expected:    types.Phone("5562982870053"),
			expectError: false,
		},
		{
			name:        "Valid phone with formatting (plus, spaces, hyphens)",
			input:       "+55 62 98287-0053",
			expected:    types.Phone("5562982870053"),
			expectError: false,
		},
		{
			name:          "Empty phone string",
			input:         "",
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "Phone number cannot be empty",
		},
		{
			name:          "Phone too short (after normalization and no 55 prefix)",
			input:         "6298287005", // 10 digits
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "Normalized phone number must have 13 digits",
		},
		{
			name:          "Phone too long (after normalization)",
			input:         "55629828700531", // 14 digits
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "Normalized phone number must have 13 digits",
		},
		{
			name:          "Phone 13 digits but wrong country code",
			input:         "5462982870053", // Starts with 54
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "Normalized 13-digit phone number must start with country code '55'",
		},
		{
			name:          "Phone 11 digits but starting with 55 (ambiguous)",
			input:         "55123456789", // 55 + 9 digits
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "11-digit number starting with country code '55' is ambiguous",
		},
		{
			name:          "Phone with non-numeric characters only",
			input:         "abc-def",
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "Normalized phone number must have 13 digits",
		},
		{
			name:          "Raw input too long",
			input:         strings.Repeat("1", types.MaxRawPhoneInputLength+1),
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "Raw phone input (length 31) exceeds maximum length",
		},
		{
			name:          "Normalized to empty and then prefixed",
			input:         "() -",
			expectError:   true,
			expectedCode:  msg.CodeInvalid,
			errorContains: "Normalized phone number must have 13 digits",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			phone, err := types.NewPhone(tc.input)

			if tc.expectError {
				require.Error(t, err, "NewPhone(%s) should have returned an error", tc.input)
				var msgErr *msg.MessageError
				if assert.ErrorAs(t, err, &msgErr, "NewPhone(%s): error should be a *msg.MessageError", tc.input) {
					if tc.expectedCode != "" {
						assert.Equal(t, tc.expectedCode, msgErr.Code, "NewPhone(%s): error code mismatch", tc.input)
					}
					if tc.errorContains != "" {
						assert.Contains(t, msgErr.Message, tc.errorContains, "NewPhone(%s): error message content mismatch", tc.input)
					}
				}
			} else {
				require.NoError(t, err, "NewPhone(%s) should not have returned an error", tc.input)
				assert.Equal(t, tc.expected, phone, "NewPhone(%s): Phone value mismatch", tc.input)
				assert.False(t, phone.IsEmpty(), "NewPhone(%s): Valid phone should not be empty", tc.input)
			}
		})
	}
}

func TestPhone_IsEmpty(t *testing.T) {
	assert.True(t, types.Phone("").IsEmpty(), "Phone(\"\").IsEmpty() should return true")

	pValid, err := types.NewPhone("5562982870053")
	require.NoError(t, err, "Setup for IsEmpty: NewPhone valid should not error")
	assert.False(t, pValid.IsEmpty(), "Valid Phone from NewPhone should not be empty")
}

func TestPhone_String(t *testing.T) {
	pValid, err := types.NewPhone("5562982870053")
	require.NoError(t, err, "Setup for String: NewPhone valid should not error")
	assert.Equal(t, "5562982870053", pValid.String(), "pValid.String() mismatch")

	assert.Equal(t, "", types.Phone("").String(), "Phone(\"\").String() should be empty string")
}

func TestPhone_JSONMarshaling(t *testing.T) {
	validPhoneStr := "5562982870053"
	p, err := types.NewPhone(validPhoneStr)
	require.NoError(t, err, "JSONMarshaling setup: NewPhone(%s) failed", validPhoneStr)

	t.Run("Marshal valid Phone", func(t *testing.T) {
		jsonData, err := json.Marshal(p)
		require.NoError(t, err, "json.Marshal(Phone) should not error for valid Phone")
		assert.JSONEq(t, `"`+validPhoneStr+`"`, string(jsonData), "Marshaled JSON string mismatch")
	})

	t.Run("Unmarshal to valid Phone", func(t *testing.T) {
		jsonData := []byte(`"` + validPhoneStr + `"`)
		var unmarshaledPhone types.Phone
		err := json.Unmarshal(jsonData, &unmarshaledPhone)
		require.NoError(t, err, "json.Unmarshal should not error for valid Phone JSON")
		assert.Equal(t, p, unmarshaledPhone, "Unmarshaled Phone mismatch")
	})

	t.Run("Unmarshal non-string JSON", func(t *testing.T) {
		invalidJsonDataNonString := []byte(`123`)
		var invalidP1 types.Phone
		err := json.Unmarshal(invalidJsonDataNonString, &invalidP1)
		require.Error(t, err, "json.Unmarshal should error for non-string JSON")
		var msgErr1 *msg.MessageError
		require.ErrorAs(t, err, &msgErr1, "Error should be a *msg.MessageError for non-string JSON")
		assert.Equal(t, msg.CodeInvalid, msgErr1.Code, "Error code for non-string JSON mismatch")
		assert.Contains(t, msgErr1.Message, "Phone must be a valid JSON string", "Error message for non-string JSON mismatch")
	})

	t.Run("Unmarshal invalid Phone format string", func(t *testing.T) {
		invalidJsonDataInvalidPhone := []byte(`"invalid-phone-format"`)
		var invalidP2 types.Phone
		err := json.Unmarshal(invalidJsonDataInvalidPhone, &invalidP2)
		require.Error(t, err, "json.Unmarshal should error for invalid phone format string")
		var msgErr2 *msg.MessageError
		require.ErrorAs(t, err, &msgErr2, "Error should be a *msg.MessageError for invalid phone format")
		assert.Equal(t, msg.CodeInvalid, msgErr2.Code, "Error code for invalid phone format mismatch")
		assert.Contains(t, msgErr2.Message, "Normalized phone number must have 13 digits", "Error message for invalid phone format mismatch")
	})
}

func TestPhone_TextMarshaling(t *testing.T) {
	validPhoneStr := "5562982870053"
	p, err := types.NewPhone(validPhoneStr)
	require.NoError(t, err, "TextMarshaling setup: NewPhone(%s) failed", validPhoneStr)

	t.Run("MarshalText valid Phone", func(t *testing.T) {
		textData, err := p.MarshalText()
		require.NoError(t, err, "Phone.MarshalText() should not error for valid Phone")
		assert.Equal(t, validPhoneStr, string(textData), "Marshaled text mismatch")
	})

	t.Run("UnmarshalText to valid Phone", func(t *testing.T) {
		textData := []byte(validPhoneStr)
		var unmarshaledPhone types.Phone
		err := unmarshaledPhone.UnmarshalText(textData)
		require.NoError(t, err, "Phone.UnmarshalText() should not error for valid text")
		assert.Equal(t, p, unmarshaledPhone, "Unmarshaled Phone from text mismatch")
	})

	t.Run("UnmarshalText with invalid Phone format", func(t *testing.T) {
		invalidTextData := []byte("invalid-phone-format")
		var invalidP types.Phone
		err := invalidP.UnmarshalText(invalidTextData)
		require.Error(t, err, "Phone.UnmarshalText() should error for invalid phone format text")
		var msgErr *msg.MessageError
		require.ErrorAs(t, err, &msgErr, "Error should be a *msg.MessageError for invalid text format")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code for invalid text format mismatch")
	})
}

func TestPhone_SQLDriver(t *testing.T) {
	validPhoneStr := "5562982870053"
	validPhone, err := types.NewPhone(validPhoneStr)
	require.NoError(t, err, "SQLDriver setup: NewPhone(%s) failed", validPhoneStr)

	t.Run("Value success", func(t *testing.T) {
		val, err := validPhone.Value()
		require.NoError(t, err, "validPhone.Value() should not error")
		assert.Equal(t, validPhoneStr, val.(string), "Value() returned string mismatch")
	})

	t.Run("Value on uninitialized (empty) Phone (should error)", func(t *testing.T) {
		p := types.Phone("") // Invalid state, NewPhone would prevent this.
		val, err := p.Value()
		require.Error(t, err, "Phone(\"\").Value() should error")
		assert.Nil(t, val, "Value should be nil on error")
		var msgErr *msg.MessageError
		require.ErrorAs(t, err, &msgErr, "Error from Phone(\"\").Value() should be *msg.MessageError")
		assert.Contains(t, msgErr.Message, "Attempted to save an empty or invalid Phone value", "Error message for empty Phone value mismatch")
	})

	t.Run("Scan success for string", func(t *testing.T) {
		var p types.Phone
		err := p.Scan(validPhoneStr)
		require.NoError(t, err, "p.Scan(%s) should not error", validPhoneStr)
		assert.Equal(t, validPhone, p, "Scanned Phone from string mismatch")
	})

	t.Run("Scan success for byte slice", func(t *testing.T) {
		byteSlice := []byte(validPhoneStr)
		var p types.Phone
		err := p.Scan(byteSlice)
		require.NoError(t, err, "p.Scan([]byte(%s)) should not error", validPhoneStr)
		assert.Equal(t, validPhone, p, "Scanned Phone from byte slice mismatch")
	})

	t.Run("Scan success with normalization and prefixing from DB string", func(t *testing.T) {
		phoneStrWithFormatting := "(62) 98287-0053"
		expectedPhone, _ := types.NewPhone(phoneStrWithFormatting) // Use NewPhone to get expected normalized form
		var p types.Phone
		err := p.Scan(phoneStrWithFormatting)
		require.NoError(t, err, "p.Scan(%s) with formatting should not error", phoneStrWithFormatting)
		assert.Equal(t, expectedPhone, p, "Scanned Phone with formatting mismatch")
	})

	t.Run("Scan nil (should error as Phone is non-nullable)", func(t *testing.T) {
		var p types.Phone
		err := p.Scan(nil)
		require.Error(t, err, "p.Scan(nil) should error")
		var msgErr *msg.MessageError
		require.ErrorAs(t, err, &msgErr, "Error from p.Scan(nil) should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code for Scan(nil) mismatch")
		assert.Contains(t, msgErr.Message, "Scanned nil value for non-nullable Phone type", "Error message for Scan(nil) mismatch")
	})

	t.Run("Scan invalid type (int)", func(t *testing.T) {
		var p types.Phone
		err := p.Scan(12345)
		require.Error(t, err, "p.Scan(int) should error")
		var msgErr *msg.MessageError
		require.ErrorAs(t, err, &msgErr, "Error from p.Scan(int) should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code for Scan(int) mismatch")
		assert.Contains(t, msgErr.Message, "Incompatible type (int) for Phone scan", "Error message for Scan(int) mismatch")
	})

	t.Run("Scan invalid phone format string from DB", func(t *testing.T) {
		invalidDBString := "12345"
		var p types.Phone
		err := p.Scan(invalidDBString)
		require.Error(t, err, "p.Scan(%s) from DB with invalid format should error", invalidDBString)
		var msgErr *msg.MessageError
		require.ErrorAs(t, err, &msgErr, "Error from Scan(invalid format) should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code for Scan(invalid format) mismatch")
		assert.Contains(t, msgErr.Message, "Normalized phone number must have 13 digits", "Error message for Scan(invalid format) mismatch")
	})
}

func TestMustNewPhone(t *testing.T) {
	t.Run("Valid phone does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			phone := types.MustNewPhone("5562982870053")
			assert.Equal(t, types.Phone("5562982870053"), phone, "MustNewPhone with valid input returned incorrect value")
		}, "MustNewPhone with valid input should not panic")
	})

	t.Run("Invalid phone panics with expected message", func(t *testing.T) {
		assert.Panics(t, func() {
			types.MustNewPhone("invalid")
		}, "MustNewPhone with invalid input should panic")

		defer func() {
			r := recover()
			require.NotNil(t, r, "Expected a panic for invalid phone")
			err, ok := r.(error)
			require.True(t, ok, "Panic value should be an error")
			assert.Contains(t, err.Error(), "Normalized phone number must have 13 digits")
		}()
		types.MustNewPhone("invalid")
	})
}
