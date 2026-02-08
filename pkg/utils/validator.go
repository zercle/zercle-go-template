// Package utils provides utility functions for the application.
package utils

import (
	"errors"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// Validator provides validation functionality using go-playground/validator.
type Validator struct {
	validate *validator.Validate
}

var (
	instance *Validator
	once     sync.Once
)

// GetValidator returns the singleton validator instance.
func GetValidator() *Validator {
	once.Do(func() {
		v := &Validator{
			validate: validator.New(),
		}
		// Register custom validators
		_ = v.validate.RegisterValidation("alphanumspace", validateAlphaNumSpace)
		instance = v
	})
	return instance
}

// ValidateStruct validates a struct using struct tags.
func (v *Validator) ValidateStruct(s any) error {
	return v.validate.Struct(s)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validate.Var(field, tag)
}

// ValidationErrors converts validator errors to a map of field->message.
func (v *Validator) ValidationErrors(err error) map[string]string {
	result := make(map[string]string)

	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			message := getErrorMessage(e)
			result[field] = message
		}
	}

	return result
}

// getErrorMessage returns a human-readable error message for a validation error.
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short, minimum length is " + e.Param()
	case "max":
		return "Value is too long, maximum length is " + e.Param()
	case "len":
		return "Value must be exactly " + e.Param() + " characters"
	case "alphanumspace":
		return "Value can only contain letters, numbers, and spaces"
	default:
		return "Invalid value"
	}
}

// validateAlphaNumSpace validates that a string contains only alphanumeric characters and spaces.
func validateAlphaNumSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != ' ' {
			return false
		}
	}
	return true
}

// BindAndValidate binds request data and validates it.
// Returns validation errors as a map if validation fails.
func BindAndValidate(c interface{ BodyParser(obj any) error }, obj any) (map[string]string, error) {
	if err := c.BodyParser(obj); err != nil {
		return nil, err
	}

	v := GetValidator()
	if err := v.ValidateStruct(obj); err != nil {
		return v.ValidationErrors(err), nil
	}

	return nil, nil
}

// InitFiberValidator initializes the validator for Fiber.
// This is a no-op since we use the validator directly, but kept for consistency.
func InitFiberValidator() {
	// The validator is initialized lazily in GetValidator()
	// This function exists for symmetry with the old Gin-based InitGinValidator
	_ = GetValidator()
}

// InitGinValidator initializes the Gin binding validator with our custom validator.
// Deprecated: Use InitFiberValidator instead. Kept for backward compatibility.
func InitGinValidator() {
	InitFiberValidator()
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	var validationErrors validator.ValidationErrors
	return errors.As(err, &validationErrors)
}
