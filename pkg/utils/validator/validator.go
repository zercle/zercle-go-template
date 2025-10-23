// Package validator provides validation utilities using go-playground/validator.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct and returns validation errors in a user-friendly format.
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// formatValidationError formats validator errors into a readable error message.
func formatValidationError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	errors := make([]string, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		errors = append(errors, formatFieldError(fieldErr))
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
}

// formatFieldError formats a single field error into a readable message.
func formatFieldError(fieldErr validator.FieldError) string {
	field := strings.ToLower(fieldErr.Field())

	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, fieldErr.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fieldErr.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fieldErr.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fieldErr.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fieldErr.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fieldErr.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	default:
		return fmt.Sprintf("%s failed validation for '%s'", field, fieldErr.Tag())
	}
}

// GetValidator returns the underlying validator instance for custom registrations.
func GetValidator() *validator.Validate {
	return validate
}
