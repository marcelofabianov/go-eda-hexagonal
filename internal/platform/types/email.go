package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

const (
	MaxEmailLength = 254
)

var emailRegexPattern = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

type Email string

func validateEmail(emailStr string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(emailStr))

	if normalizedEmail == "" {
		return "", msg.NewValidationError(nil,
			map[string]any{"input_email": emailStr},
			"Email address cannot be empty.",
		)
	}
	if len(normalizedEmail) > MaxEmailLength {
		message := fmt.Sprintf("Email address (length %d) exceeds maximum length of %d characters.", len(normalizedEmail), MaxEmailLength)
		return "", msg.NewValidationError(nil,
			map[string]any{"length": len(normalizedEmail), "max_length": MaxEmailLength, "input_email": emailStr},
			message,
		)
	}
	if !emailRegexPattern.MatchString(normalizedEmail) {
		message := fmt.Sprintf("Email address '%s' has an invalid format.", emailStr)
		return "", msg.NewValidationError(nil,
			map[string]any{"input_email": emailStr},
			message,
		)
	}
	return normalizedEmail, nil
}

func NewEmail(emailStr string) (Email, error) {
	validatedEmail, err := validateEmail(emailStr)
	if err != nil {
		return "", err
	}
	return Email(validatedEmail), nil
}

func MustNewEmail(emailStr string) Email {
	email, err := NewEmail(emailStr)
	if err != nil {
		panic(err)
	}
	return email
}

func (e Email) String() string {
	return string(e)
}

func (e Email) IsEmpty() bool {
	return string(e) == ""
}

func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Email) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		message := fmt.Sprintf("Email must be a valid JSON string (received: %s).", string(data))
		return msg.NewValidationError(err,
			map[string]any{"input_json": string(data)},
			message,
		)
	}
	validatedEmail, err := validateEmail(s)
	if err != nil {
		return err
	}
	*e = Email(validatedEmail)
	return nil
}

func (e Email) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

func (e *Email) UnmarshalText(text []byte) error {
	emailStr := string(text)
	validatedEmail, err := validateEmail(emailStr)
	if err != nil {
		return err
	}
	*e = Email(validatedEmail)
	return nil
}

func (e Email) Value() (driver.Value, error) {
	return e.String(), nil
}

func (e *Email) Scan(src interface{}) error {
	if src == nil {
		return msg.NewValidationError(nil,
			map[string]any{"target_type": "Email"},
			"Scanned nil value for non-nullable Email type.",
		)
	}
	var emailStr string
	switch sval := src.(type) {
	case string:
		emailStr = sval
	case []byte:
		emailStr = string(sval)
	default:
		message := fmt.Sprintf("Incompatible type (%T) for Email. Expected string or []byte.", src)
		return msg.NewValidationError(nil,
			map[string]any{"received_type": fmt.Sprintf("%T", src)},
			message,
		)
	}

	validatedEmail, err := validateEmail(emailStr)
	if err != nil {
		if originalMsgErr, ok := err.(*msg.MessageError); ok {
			originalMsgErr.WithContext("scan_source_value", emailStr)
			return originalMsgErr
		}
		message := fmt.Sprintf("Failed to scan database value ('%s') to Email.", emailStr)
		return msg.NewValidationError(err,
			map[string]any{"scan_source_value": emailStr},
			message,
		)
	}
	*e = Email(validatedEmail)
	return nil
}
