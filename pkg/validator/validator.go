// Package validator provides request validation using go-playground/validator.
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	apperrors "github.com/zercle/zercle-go-template/pkg/errors"
)

// Validator wraps the go-playground validator.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	v := validator.New()

	// Register function to get JSON tag name for struct field
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

// Validate validates a struct based on its validation tags.
func (v *Validator) ValidateStruct(s any) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		return apperrors.ErrValidation
	}

	var appErrors apperrors.ValidationErrors
	for _, e := range validationErrors {
		appErrors = append(appErrors, apperrors.ValidationError{
			Field:   e.Field(),
			Message: formatValidationError(e),
		})
	}

	return appErrors
}

// ValidateVar validates a single variable against a tag.
func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validate.Var(field, tag)
}

// formatValidationError converts a validation error to a human-readable message.
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", e.Field(), e.Param())
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", e.Field(), e.Param())
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", e.Field(), e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", e.Field(), e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", e.Field(), e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", e.Field(), e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", e.Field(), e.Param())
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", e.Field(), e.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", e.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", e.Field())
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime in format %s", e.Field(), e.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", e.Field(), e.Tag())
	}
}

// HasErrors checks if there are any validation errors.
func (v *Validator) HasErrors(err error) bool {
	var validationErrors apperrors.ValidationErrors
	return errors.As(err, &validationErrors)
}
