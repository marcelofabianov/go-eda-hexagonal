package types_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/types"
)

func TestNewEmail(t *testing.T) {
	validMaxLengthLocalPart := strings.Repeat("a", types.MaxEmailLength-len("@example.com"))
	validMaxLengthEmail := validMaxLengthLocalPart + "@example.com"

	invalidTooLongLocalPart := strings.Repeat("a", types.MaxEmailLength-len("@example.com")+1)
	invalidTooLongEmail := invalidTooLongLocalPart + "@example.com"

	testCases := []struct {
		name            string
		input           string
		expectedEmail   types.Email
		expectError     bool
		expectedErrCode msg.ErrorCode
		expectedMessage string
	}{
		{"valid simple", "test@example.com", types.Email("test@example.com"), false, "", ""},
		{"valid with subdomain", "test@sub.example.com", types.Email("test@sub.example.com"), false, "", ""},
		{"valid with long TLD", "test@example.co.uk", types.Email("test@example.co.uk"), false, "", ""},
		{"valid with numbers", "test123@example.com", types.Email("test123@example.com"), false, "", ""},
		{"valid with hyphen in domain", "test@example-domain.com", types.Email("test@example-domain.com"), false, "", ""},
		{"valid with dot in local part", "first.last@example.com", types.Email("first.last@example.com"), false, "", ""},
		{"valid with + alias", "test+alias@example.com", types.Email("test+alias@example.com"), false, "", ""},
		{"normalization lowercase", "Test@Example.COM", types.Email("test@example.com"), false, "", ""},
		{"normalization trim spaces", "  test@example.com  ", types.Email("test@example.com"), false, "", ""},
		{"valid at max length", validMaxLengthEmail, types.Email(strings.ToLower(validMaxLengthEmail)), false, "", ""},
		{"invalid empty", "", types.Email(""), true, msg.CodeInvalid, "Email address cannot be empty."},
		{"invalid spaces only", "   ", types.Email(""), true, msg.CodeInvalid, "Email address cannot be empty."},
		{"invalid no at sign", "testexample.com", types.Email(""), true, msg.CodeInvalid, "Email address 'testexample.com' has an invalid format."},
		{"invalid no local part", "@example.com", types.Email(""), true, msg.CodeInvalid, "Email address '@example.com' has an invalid format."},
		{"invalid no domain", "test@", types.Email(""), true, msg.CodeInvalid, "Email address 'test@' has an invalid format."},
		{"invalid exceeds length", invalidTooLongEmail, types.Email(""), true, msg.CodeInvalid, fmt.Sprintf("Email address (length %d) exceeds maximum length of %d characters.", len(invalidTooLongEmail), types.MaxEmailLength)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			email, err := types.NewEmail(tc.input)
			if tc.expectError {
				require.Error(t, err, fmt.Sprintf("NewEmail('%s') should return an error", tc.input))
				var msgErr *msg.MessageError
				isMsgError := errors.As(err, &msgErr)
				require.True(t, isMsgError, "Error should be of type *msg.MessageError")
				assert.Equal(t, tc.expectedErrCode, msgErr.Code, "Error code mismatch")
				if tc.expectedMessage != "" {
					assert.Equal(t, tc.expectedMessage, msgErr.Message, "Error message content mismatch")
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("NewEmail('%s') should not return an error", tc.input))
				assert.Equal(t, tc.expectedEmail, email, "Email result mismatch")
			}
		})
	}
}

func TestMustNewEmail(t *testing.T) {
	assert.NotPanics(t, func() {
		email := types.MustNewEmail("valid@example.com")
		assert.Equal(t, types.Email("valid@example.com"), email, "Email from MustNewEmail should match valid input")
	}, "MustNewEmail should not panic for valid email")

	defer func() {
		r := recover()
		require.NotNil(t, r, "MustNewEmail should panic for invalid email")
		err, ok := r.(error)
		require.True(t, ok, "Panic value should be an error")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Panic error should be a *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Panic error code mismatch")
		assert.Contains(t, msgErr.Message, "Email address", "Panic error message should indicate an email issue")
	}()
	_ = types.MustNewEmail("invalid")
}

func TestEmail_StringAndIsEmpty(t *testing.T) {
	email, _ := types.NewEmail("test@example.com")
	assert.Equal(t, "test@example.com", email.String(), "String() should return the correct email string")
	assert.False(t, email.IsEmpty(), "Non-empty email should return false for IsEmpty()")

	var emptyEmail types.Email
	assert.True(t, emptyEmail.IsEmpty(), "Zero-value Email should be empty")
	assert.Equal(t, "", emptyEmail.String(), "String() for zero-value Email should be empty string")
}

func TestEmail_TextEncoding(t *testing.T) {
	validEmail := types.Email("user@example.com")

	t.Run("MarshalText", func(t *testing.T) {
		text, err := validEmail.MarshalText()
		require.NoError(t, err, "MarshalText should not return an error")
		assert.Equal(t, []byte("user@example.com"), text, "MarshalText output mismatch")
	})

	t.Run("UnmarshalText valid (with normalization)", func(t *testing.T) {
		var email types.Email
		err := email.UnmarshalText([]byte("  USER@EXAMPLE.COM  "))
		require.NoError(t, err, "UnmarshalText for valid (normalizable) text should not error")
		assert.Equal(t, types.Email("user@example.com"), email, "UnmarshalText should normalize email")
	})

	t.Run("UnmarshalText invalid", func(t *testing.T) {
		var email types.Email
		err := email.UnmarshalText([]byte("invalid"))
		require.Error(t, err, "UnmarshalText should return error for invalid email text")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
	})
}

func TestEmail_JSONEncoding(t *testing.T) {
	validEmail := types.Email("user@domain.com")

	t.Run("MarshalJSON", func(t *testing.T) {
		jsonData, err := json.Marshal(validEmail)
		require.NoError(t, err, "json.Marshal should not return an error")
		expectedJSON := fmt.Sprintf("\"%s\"", validEmail)
		assert.JSONEq(t, expectedJSON, string(jsonData), "Marshalled JSON string mismatch")
	})

	t.Run("UnmarshalJSON valid (with normalization)", func(t *testing.T) {
		jsonData := []byte(fmt.Sprintf("\"%s\"", "  USER@DOMAIN.COM  "))
		var newEmail types.Email
		err := json.Unmarshal(jsonData, &newEmail)
		require.NoError(t, err, "json.Unmarshal for valid (normalizable) JSON string should not error")
		assert.Equal(t, types.Email("user@domain.com"), newEmail, "Unmarshalled email should be normalized")
	})

	t.Run("UnmarshalJSON invalid string content", func(t *testing.T) {
		var email types.Email
		err := json.Unmarshal([]byte("\"invalid\""), &email)
		require.Error(t, err, "json.Unmarshal should error for invalid email content in JSON string")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
	})

	t.Run("UnmarshalJSON non-string", func(t *testing.T) {
		var email types.Email
		err := json.Unmarshal([]byte("123"), &email)
		require.Error(t, err, "json.Unmarshal should error for non-string JSON")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
	})

	t.Run("UnmarshalJSON null (expect error)", func(t *testing.T) {
		var email types.Email
		err := json.Unmarshal([]byte("null"), &email)
		require.Error(t, err, "json.Unmarshal for JSON null should error for Email type")
		var msgErr *msg.MessageError
		require.True(t, errors.As(err, &msgErr), "Error should be *msg.MessageError")
		assert.Equal(t, msg.CodeInvalid, msgErr.Code, "Error code should be CodeInvalid")
	})
}

func TestEmail_DatabaseEncoding(t *testing.T) {
	validNormalizedEmail := types.Email("dbuser@example.com")

	t.Run("Value", func(t *testing.T) {
		value, err := validNormalizedEmail.Value()
		require.NoError(t, err, "Value() should not return an error")
		assert.Equal(t, "dbuser@example.com", value, "Value() output mismatch")
	})

	testCasesScan := []struct {
		name            string
		src             interface{}
		expected        types.Email
		expectError     bool
		expectedErrCode msg.ErrorCode
		expectedMessage string
	}{
		{"Scan valid string (normalized)", "dbuser@example.com", types.Email("dbuser@example.com"), false, "", ""},
		{"Scan valid string (needs normalization)", "  DBUser@Example.COM  ", types.Email("dbuser@example.com"), false, "", ""},
		{"Scan valid []byte", []byte("db_byte@example.com"), types.Email("db_byte@example.com"), false, "", ""},
		{"Scan invalid string", "invalid-db", types.Email(""), true, msg.CodeInvalid, "Email address 'invalid-db' has an invalid format."},
		{"Scan nil (expect error)", nil, types.Email(""), true, msg.CodeInvalid, "Scanned nil value for non-nullable Email type."},
		{"Scan incompatible type", 123, types.Email(""), true, msg.CodeInvalid, "Incompatible type (int) for Email. Expected string or []byte."},
		{"Scan empty string (expect error)", "", types.Email(""), true, msg.CodeInvalid, "Email address cannot be empty."},
		{"Scan empty []byte (expect error)", []byte(""), types.Email(""), true, msg.CodeInvalid, "Email address cannot be empty."},
	}

	for _, tc := range testCasesScan {
		t.Run(tc.name, func(t *testing.T) {
			var email types.Email
			err := email.Scan(tc.src)
			if tc.expectError {
				require.Error(t, err, fmt.Sprintf("Scan for '%s' should error", tc.name))
				var msgErr *msg.MessageError
				isMsgError := errors.As(err, &msgErr)
				require.True(t, isMsgError, fmt.Sprintf("Error for '%s' should be a *msg.MessageError: %v", tc.name, err))
				assert.Equal(t, tc.expectedErrCode, msgErr.Code, fmt.Sprintf("Error code mismatch for '%s'", tc.name))
				if tc.expectedMessage != "" {
					assert.Equal(t, tc.expectedMessage, msgErr.Message, fmt.Sprintf("Error message mismatch for '%s'", tc.name))
				}
			} else {
				assert.NoError(t, err, fmt.Sprintf("Scan for '%s' should not error", tc.name))
				assert.Equal(t, tc.expected, email, fmt.Sprintf("Scan for '%s' result mismatch", tc.name))
			}
		})
	}
}
