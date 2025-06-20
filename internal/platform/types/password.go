package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

const (
	MinPasswordLength = 10
)

var (
	hasNumberRegex = regexp.MustCompile(`[0-9]`)
	hasUpperRegex  = regexp.MustCompile(`[A-Z]`)
	hasLowerRegex  = regexp.MustCompile(`[a-z]`)
	hasSymbolRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

type Password string

func validatePassword(passwordStr string) error {
	if passwordStr == "" {
		return msg.NewValidationError(nil,
			map[string]any{"field": "password"},
			"Password cannot be empty.",
		)
	}

	passwordRuneLen := utf8.RuneCountInString(passwordStr)
	if passwordRuneLen < MinPasswordLength {
		message := fmt.Sprintf("Password must be at least %d characters long.", MinPasswordLength)
		return msg.NewValidationError(nil,
			map[string]any{"field": "password", "min_length": MinPasswordLength, "actual_length": passwordRuneLen},
			message,
		)
	}

	if !hasNumberRegex.MatchString(passwordStr) {
		return msg.NewValidationError(nil,
			map[string]any{"field": "password", "rule_violation": "missing_numeric"},
			"Password must contain at least one numeric character (0-9).",
		)
	}

	if !hasUpperRegex.MatchString(passwordStr) {
		return msg.NewValidationError(nil,
			map[string]any{"field": "password", "rule_violation": "missing_uppercase"},
			"Password must contain at least one uppercase letter (A-Z).",
		)
	}

	if !hasLowerRegex.MatchString(passwordStr) {
		return msg.NewValidationError(nil,
			map[string]any{"field": "password", "rule_violation": "missing_lowercase"},
			"Password must contain at least one lowercase letter (a-z).",
		)
	}

	if !hasSymbolRegex.MatchString(passwordStr) {
		return msg.NewValidationError(nil,
			map[string]any{"field": "password", "rule_violation": "missing_symbol"},
			"Password must contain at least one symbol.",
		)
	}

	return nil
}

func NewPassword(passwordStr string) (Password, error) {
	if err := validatePassword(passwordStr); err != nil {
		return "", err
	}
	return Password(passwordStr), nil
}

func MustNewPassword(passwordStr string) Password {
	p, err := NewPassword(passwordStr)
	if err != nil {
		panic(err)
	}
	return p
}

func (p Password) String() string {
	return string(p)
}

func (p Password) IsEmpty() bool {
	return string(p) == ""
}

func (p Password) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Password) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return msg.NewMessageError(
			err,
			"Password must be a valid JSON string.",
			msg.CodeInvalid,
			nil,
		)
	}

	newP, err := NewPassword(s)
	if err != nil {
		return err
	}
	*p = newP
	return nil
}

func (p Password) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Password) UnmarshalText(text []byte) error {
	s := string(text)
	newP, err := NewPassword(s)
	if err != nil {
		return err
	}
	*p = newP
	return nil
}

func (p Password) Value() (driver.Value, error) {
	if p.IsEmpty() {
		return nil, msg.NewValidationError(nil,
			map[string]any{"field": "password"},
			"Attempted to save an empty Password value to the database.",
		)
	}
	return p.String(), nil
}

func (p *Password) Scan(src interface{}) error {
	if src == nil {
		*p = ""
		return msg.NewValidationError(nil,
			map[string]any{"target_type": "Password"},
			"Scanned nil value for non-nullable Password type from database.",
		)
	}

	var passwordStr string
	switch sval := src.(type) {
	case string:
		passwordStr = sval
	case []byte:
		passwordStr = string(sval)
	default:
		message := fmt.Sprintf("Incompatible type (%T) for Password scan.", src)
		return msg.NewValidationError(nil,
			map[string]any{"received_type": fmt.Sprintf("%T", src)},
			message,
		)
	}

	newP, err := NewPassword(passwordStr)
	if err != nil {
		return err
	}
	*p = newP
	return nil
}
