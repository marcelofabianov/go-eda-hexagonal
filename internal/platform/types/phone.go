package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

const (
	DefaultCountryCode     = "55"
	CountryCodeLength      = 2
	DDDLength              = 2
	LocalPhoneNumberLength = 9
	NormalizedPhoneLength  = CountryCodeLength + DDDLength + LocalPhoneNumberLength
	MaxRawPhoneInputLength = 30
)

var nonNumericRegexPattern = regexp.MustCompile(`\D+`)

type Phone string

func normalizePhone(phoneStr string) string {
	return nonNumericRegexPattern.ReplaceAllString(phoneStr, "")
}

func validateAndPrefixNormalizedPhone(normalizedNum string, originalInputForErrorContext string) (string, error) {
	numLen := len(normalizedNum)
	finalNum := normalizedNum

	if numLen == (DDDLength + LocalPhoneNumberLength) {
		if strings.HasPrefix(finalNum, DefaultCountryCode) {
			message := fmt.Sprintf("Invalid phone number format: 11-digit number starting with country code '%s' is ambiguous or incomplete.", DefaultCountryCode)
			return "", msg.NewValidationError(nil,
				map[string]any{"input_phone": originalInputForErrorContext, "normalized_phone": finalNum},
				message,
			)
		}
		finalNum = DefaultCountryCode + finalNum
		numLen = len(finalNum)
	}

	if numLen != NormalizedPhoneLength {
		message := fmt.Sprintf("Normalized phone number must have %d digits (e.g., 55DDNNNNNNNNN), got %d.", NormalizedPhoneLength, numLen)
		return "", msg.NewValidationError(nil,
			map[string]any{"input_phone": originalInputForErrorContext, "normalized_phone_after_prefix_attempt": finalNum, "expected_length": NormalizedPhoneLength, "actual_length": numLen},
			message,
		)
	}

	if !strings.HasPrefix(finalNum, DefaultCountryCode) {
		message := fmt.Sprintf("Normalized 13-digit phone number must start with country code '%s'.", DefaultCountryCode)
		return "", msg.NewValidationError(nil,
			map[string]any{"input_phone": originalInputForErrorContext, "normalized_phone": finalNum, "expected_prefix": DefaultCountryCode},
			message,
		)
	}

	return finalNum, nil
}

func NewPhone(phoneStr string) (Phone, error) {
	trimmedInput := strings.TrimSpace(phoneStr)
	if trimmedInput == "" {
		return "", msg.NewValidationError(nil,
			map[string]any{"input_phone": phoneStr},
			"Phone number cannot be empty.",
		)
	}

	if utf8.RuneCountInString(trimmedInput) > MaxRawPhoneInputLength {
		message := fmt.Sprintf("Raw phone input (length %d) exceeds maximum length of %d characters.", utf8.RuneCountInString(trimmedInput), MaxRawPhoneInputLength)
		return "", msg.NewValidationError(nil,
			map[string]any{"max_length": MaxRawPhoneInputLength, "input_phone": phoneStr},
			message,
		)
	}

	normalized := normalizePhone(trimmedInput)
	validatedNum, err := validateAndPrefixNormalizedPhone(normalized, phoneStr)
	if err != nil {
		return "", err
	}

	return Phone(validatedNum), nil
}

func MustNewPhone(phoneStr string) Phone {
	phone, err := NewPhone(phoneStr)
	if err != nil {
		panic(err)
	}
	return phone
}

func (p Phone) String() string {
	return string(p)
}

func (p Phone) IsEmpty() bool {
	return string(p) == ""
}

func (p Phone) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Phone) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		message := fmt.Sprintf("Phone must be a valid JSON string (received: %s).", string(data))
		return msg.NewValidationError(err,
			map[string]any{"input_json": string(data)},
			message,
		)
	}
	phone, err := NewPhone(s)
	if err != nil {
		return err
	}
	*p = phone
	return nil
}

func (p Phone) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Phone) UnmarshalText(text []byte) error {
	phone, err := NewPhone(string(text))
	if err != nil {
		return err
	}
	*p = phone
	return nil
}

func (p Phone) Value() (driver.Value, error) {
	if p.IsEmpty() {
		return nil, msg.NewValidationError(nil, nil,
			"Attempted to save an empty or invalid Phone value to the database.")
	}
	return p.String(), nil
}

func (p *Phone) Scan(src interface{}) error {
	if src == nil {
		return msg.NewValidationError(nil,
			map[string]any{"target_type": "Phone"},
			"Scanned nil value for non-nullable Phone type from database.",
		)
	}

	var phoneStr string
	switch sval := src.(type) {
	case string:
		phoneStr = sval
	case []byte:
		phoneStr = string(sval)
	default:
		message := fmt.Sprintf("Incompatible type (%T) for Phone scan. Expected string or []byte.", src)
		return msg.NewValidationError(nil,
			map[string]any{"received_type": fmt.Sprintf("%T", src)},
			message,
		)
	}

	normalizedFromDB := normalizePhone(phoneStr)
	validatedNum, err := validateAndPrefixNormalizedPhone(normalizedFromDB, phoneStr)
	if err != nil {
		if originalMsgErr, ok := err.(*msg.MessageError); ok {
			originalMsgErr.WithContext("scan_source_value_db", phoneStr)
			return originalMsgErr
		}
		message := fmt.Sprintf("Failed to scan database value ('%s') to Phone due to invalid format after normalization.", phoneStr)
		return msg.NewValidationError(err,
			map[string]any{"scan_source_value_db": phoneStr},
			message,
		)
	}
	*p = Phone(validatedNum)
	return nil
}
