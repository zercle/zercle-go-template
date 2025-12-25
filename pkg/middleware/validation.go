package middleware

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/zercle/zercle-go-template/pkg/response"
)

// CustomValidator wraps go-playground/validator to integrate with Echo's validation interface.
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new validator instance for Echo request validation.
func NewCustomValidator() *CustomValidator {
	v := validator.New()
	return &CustomValidator{validator: v}
}

// Validate validates a struct using go-playground/validator rules.
// Returns validation errors if the struct fails validation.
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// ValidationErrors converts go-playground validation errors into JSend-compatible FieldError slices.
// Returns an empty slice if the error is not a validation error.
func ValidationErrors(err error) []response.FieldError {
	var fields []response.FieldError

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, e := range validationErrors {
			fields = append(fields, response.FieldError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return fields
}

// getErrorMessage translates validation error tags into user-friendly messages.
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + e.Param() + " characters"
	case "max":
		return "Must be at most " + e.Param() + " characters"
	case "gte":
		return "Must be greater than or equal to " + e.Param()
	case "lte":
		return "Must be less than or equal to " + e.Param()
	default:
		return "Failed validation"
	}
}

// Validation creates a pass-through validation middleware for Echo.
// Actual validation occurs via Echo's Validator interface (CustomValidator).
func Validation() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
