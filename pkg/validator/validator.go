// Package validator provides input validation helpers for HTTP request DTOs.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError holds field-level validation failures.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate runs struct-level validation and returns a slice of field errors.
// Returns nil if the struct is valid.
func Validate(s any) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errs []ValidationError
	for _, fe := range err.(validator.ValidationErrors) {
		errs = append(errs, ValidationError{
			Field:   strings.ToLower(fe.Field()),
			Message: fieldMessage(fe),
		})
	}
	return errs
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("minimum length is %s", fe.Param())
	case "max":
		return fmt.Sprintf("maximum length is %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
