package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

func (v *Validator) Validate(data any) error {
	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return msg.NewInternalError(err, map[string]any{"details": "unexpected validation error type"})
	}

	var fieldErrors []map[string]string
	for _, fieldError := range validationErrors {
		message := fmt.Sprintf("Validation failed on the '%s' rule.", fieldError.Tag())
		if fieldError.Param() != "" {
			message = fmt.Sprintf("Validation failed on the '%s' rule (param: %s).", fieldError.Tag(), fieldError.Param())
		}

		fieldErrors = append(fieldErrors, map[string]string{
			"field":   fieldError.Field(),
			"rule":    fieldError.Tag(),
			"param":   fieldError.Param(),
			"message": message,
		})
	}

	context := map[string]any{
		"errors": fieldErrors,
	}

	return msg.NewValidationError(err, context, "The provided data is invalid. Please check the context for details.")
}
