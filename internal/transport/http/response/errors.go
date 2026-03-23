// Package response provides error code constants and error mapping utilities.
package response

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	user "github.com/zercle/zercle-go-template/internal/feature/user"
	apperrors "github.com/zercle/zercle-go-template/pkg/errors"
)

// Error codes for API responses following JSend error specification.
const (
	CodeInternalError   = 5000 // Internal server error
	CodeDatabaseError   = 5001 // Database operation failed
	CodeNotFound        = 4004 // Resource not found
	CodeValidationError = 4000 // Validation failed
	CodeDuplicate       = 4009 // Duplicate resource
	CodeUnauthorized    = 4001 // Authentication required
	CodeForbidden       = 4003 // Access denied
)

// ErrorCode represents a mapped error code with HTTP status.
type ErrorCode struct {
	HTTPStatus int
	Code       int
	Message    string
}

// errorMappings maps domain errors to HTTP status and error codes.
var errorMappings = map[error]ErrorCode{
	user.ErrUserNotFound: {
		HTTPStatus: http.StatusNotFound,
		Code:       CodeNotFound,
		Message:    "User not found",
	},
	user.ErrDuplicateEmail: {
		HTTPStatus: http.StatusConflict,
		Code:       CodeDuplicate,
		Message:    "Email already exists",
	},
	user.ErrInvalidEmail: {
		HTTPStatus: http.StatusBadRequest,
		Code:       CodeValidationError,
		Message:    "Invalid email format",
	},
	user.ErrInvalidPassword: {
		HTTPStatus: http.StatusBadRequest,
		Code:       CodeValidationError,
		Message:    "Invalid password",
	},
}

// MapError maps domain errors to JSend error responses.
// If the error is not recognized, it returns a generic internal error response.
func MapError(c *echo.Context, err error) error {
	// Check if it's an AppError first
	if appErr, ok := apperrors.AsAppError(err); ok {
		return Fail(c, appErr.HTTPStatus(), map[string]string{
			"message": appErr.Message,
		})
	}

	// Check if we have a mapping for this error
	if mapping, ok := errorMappings[err]; ok {
		return Fail(c, mapping.HTTPStatus, map[string]string{
			"message": mapping.Message,
		})
	}

	// Check for wrapped errors using errors.Is
	for domainErr, mapping := range errorMappings {
		if errors.Is(err, domainErr) {
			return Fail(c, mapping.HTTPStatus, map[string]string{
				"message": mapping.Message,
			})
		}
	}

	// Default to internal server error
	return InternalError(c, "An unexpected error occurred")
}

// MapErrorWithCode maps domain errors to JSend error responses with custom codes.
// This variant is useful when you need to include additional error details.
func MapErrorWithCode(c *echo.Context, err error) error {
	// Check if it's an AppError first
	if appErr, ok := apperrors.AsAppError(err); ok {
		return Error(c, appErr.HTTPStatus(), appErr.Message, appErr.Code)
	}

	// Check if we have a mapping for this error
	if mapping, ok := errorMappings[err]; ok {
		return Error(c, mapping.HTTPStatus, mapping.Message, mapping.Code)
	}

	// Check for wrapped errors using errors.Is
	for domainErr, mapping := range errorMappings {
		if errors.Is(err, domainErr) {
			return Error(c, mapping.HTTPStatus, mapping.Message, mapping.Code)
		}
	}

	// Default to internal server error
	return InternalError(c, "An unexpected error occurred")
}
