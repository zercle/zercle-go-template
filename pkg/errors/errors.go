// Package errors provides custom error types for the application.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// Sentinel errors for common error conditions.
var (
	ErrInternalServer = errors.New("internal server error")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrValidation     = errors.New("validation error")
)

// AppError represents an application error with HTTP status code.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As checks.
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code for this error.
func (e *AppError) HTTPStatus() int {
	return e.Code
}

// NewAppError creates a new AppError with the given status code and message.
func NewAppError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorWithError creates a new AppError with the given status code, message, and underlying error.
func NewAppErrorWithError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents a validation error with field-level details.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors holds multiple validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	if len(e) == 1 {
		return fmt.Sprintf("validation failed: %s - %s", e[0].Field, e[0].Message)
	}
	return fmt.Sprintf("validation failed: %d errors", len(e))
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// NotFoundError creates an error for resources that were not found.
func NotFoundError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// ConflictError creates an error for resource conflicts (e.g., duplicate entries).
func ConflictError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: fmt.Sprintf("%s already exists", resource),
	}
}

// UnauthorizedError creates an error for authentication failures.
func UnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// ForbiddenError creates an error for authorization failures.
func ForbiddenError(message string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

// BadRequestError creates an error for invalid request data.
func BadRequestError(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

// BadRequestErrorWithError creates an error for invalid request data with an underlying error.
func BadRequestErrorWithError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

// InternalServerError creates an error for internal server errors.
func InternalServerError(message string) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

// InternalServerErrorWithError creates an internal server error with an underlying error.
func InternalServerErrorWithError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError attempts to convert an error to an AppError.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// MapStdlibError maps standard library errors to AppError.
func MapStdlibError(err error) *AppError {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	if appErr, ok := AsAppError(err); ok {
		return appErr
	}

	// Map standard library and other errors
	switch {
	case errors.Is(err, errors.ErrUnsupported),
		errors.Is(err, ErrBadRequest):
		return BadRequestError(err.Error())

	case errors.Is(err, os.ErrNotExist),
		errors.Is(err, ErrNotFound):
		return NotFoundError("resource")

	case errors.Is(err, ErrConflict):
		return ConflictError("resource")

	case errors.Is(err, ErrUnauthorized):
		return UnauthorizedError("authentication required")

	case errors.Is(err, ErrForbidden):
		return ForbiddenError("access denied")

	default:
		return InternalServerErrorWithError("an error occurred", err)
	}
}

// WriteHTTPError writes an HTTP error response based on the error type.
func WriteHTTPError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.HTTPStatus())
		if encodeErr := json.NewEncoder(w).Encode(map[string]any{
			"error":   appErr.Message,
			"code":    appErr.Code,
			"details": appErr.Details,
		}); encodeErr != nil {
			// Log encode error but don't expose it to client
			fmt.Fprintf(os.Stderr, "failed to encode error response: %v\n", encodeErr)
		}
		return
	}

	// Unknown error - return internal server error
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{
		"error": "internal server error",
	}); encodeErr != nil {
		// Log encode error but don't expose it to client
		fmt.Fprintf(os.Stderr, "failed to encode error response: %v\n", encodeErr)
	}
}
