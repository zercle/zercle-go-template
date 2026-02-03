// Package domain provides shared domain contracts used across all features.
package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents standardized API error codes.
type ErrorCode int

// Error code constants for standardized API responses.
const (
	CodeNotFound     ErrorCode = 4004
	CodeDuplicate    ErrorCode = 4009
	CodeValidation   ErrorCode = 4000
	CodeUnauthorized ErrorCode = 4001
	CodeForbidden    ErrorCode = 4003
	CodeInternal     ErrorCode = 5000
	CodeBadRequest   ErrorCode = 4000
	CodeConflict     ErrorCode = 4009
)

// Sentinel errors for common conditions.
var (
	ErrInternalServer = errors.New("internal server error")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrBadRequest     = errors.New("bad request")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrValidation     = errors.New("validation error")
)

// AppError is the single application error type used across all features.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// NewAppError creates a new AppError with the given code and message.
func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// NewAppErrorWithError creates a new AppError wrapping an underlying error.
func NewAppErrorWithError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
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

// NotFoundError returns an AppError for a missing resource.
func NotFoundError(resource string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: fmt.Sprintf("%s not found", resource)}
}

// ConflictError returns an AppError for a resource conflict.
func ConflictError(resource string) *AppError {
	return &AppError{Code: http.StatusConflict, Message: fmt.Sprintf("%s already exists", resource)}
}

// UnauthorizedError returns an AppError for unauthorized access.
func UnauthorizedError(message string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: message}
}

// BadRequestError returns an AppError for bad requests.
func BadRequestError(message string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: message}
}

// InternalServerError returns an AppError for internal server errors.
func InternalServerError(message string) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: message}
}
