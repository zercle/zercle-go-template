// Package errors provides custom error types for the application.
// It defines domain-specific errors with proper error codes and HTTP status mappings.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents a machine-readable error code.
type ErrorCode string

// Predefined error codes for different error types.
const (
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
)

// AppError is the base application error type.
// It provides structured error information with code, message, and optional details.
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	Cause      error     `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error chaining.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithCause adds an underlying cause to the error.
func (e *AppError) WithCause(err error) *AppError {
	e.Cause = err
	return e
}

// WithDetails adds detailed information to the error.
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// Is implements error matching for use with errors.Is.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New creates a new AppError with the given code, message, and status.
func New(code ErrorCode, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: status,
	}
}

// Predefined error constructors for common error types.

// ValidationError creates a validation error with the given message.
func ValidationError(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

// NotFoundError creates a not found error for the given resource.
func NotFoundError(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// ConflictError creates a conflict error with the given message.
func ConflictError(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

// InternalError creates an internal server error.
func InternalError(message string) *AppError {
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

// UnauthorizedError creates an unauthorized error.
func UnauthorizedError(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// ForbiddenError creates a forbidden error.
func ForbiddenError(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

// BadRequestError creates a bad request error.
func BadRequestError(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

// Error checking helpers.

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeValidation
	}
	return false
}

// IsNotFoundError checks if an error is a not found error.
func IsNotFoundError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeNotFound
	}
	return false
}

// IsConflictError checks if an error is a conflict error.
func IsConflictError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == ErrCodeConflict
	}
	return false
}

// GetStatusCode extracts the HTTP status code from an error.
// Returns 500 if the error is not an AppError.
func GetStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
