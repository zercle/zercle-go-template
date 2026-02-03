// Package validator provides request validation using go-playground/validator.
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/zercle/zercle-go-template/internal/domain"
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

// ValidateStruct validates a struct based on its validation tags.
func (v *Validator) ValidateStruct(s any) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		return domain.ErrValidation
	}

	var appErrors domain.ValidationErrors
	for _, e := range validationErrors {
		appErrors = append(appErrors, domain.ValidationError{
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

var validationMessages = map[string]string{
	"required": "%s is required",
	"email":    "%s must be a valid email address",
	"min":      "%s must be at least %s characters",
	"max":      "%s must be at most %s characters",
	"len":      "%s must be exactly %s characters",
	"eq":       "%s must be equal to %s",
	"ne":       "%s must not be equal to %s",
	"gt":       "%s must be greater than %s",
	"gte":      "%s must be greater than or equal to %s",
	"lt":       "%s must be less than %s",
	"lte":      "%s must be less than or equal to %s",
	"eqfield":  "%s must be equal to %s",
	"uuid":     "%s must be a valid UUID",
	"url":      "%s must be a valid URL",
	"datetime": "%s must be a valid datetime in format %s",
}

// formatValidationError converts a validation error to a human-readable message.
func formatValidationError(e validator.FieldError) string {
	tmpl, ok := validationMessages[e.Tag()]
	if !ok {
		return fmt.Sprintf("%s failed validation: %s", e.Field(), e.Tag())
	}
	if e.Param() != "" {
		return fmt.Sprintf(tmpl, e.Field(), e.Param())
	}
	return fmt.Sprintf(tmpl, e.Field())
}

// HasErrors checks if there are any validation errors.
func (v *Validator) HasErrors(err error) bool {
	var validationErrors domain.ValidationErrors
	return errors.As(err, &validationErrors)
}
