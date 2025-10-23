// Package apierrors provides API error types and constructors.
package apierrors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents an error code for API responses
type ErrorCode string

const (
	// ErrCodeNotFound represents a not found error.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeBadRequest represents a bad request error.
	ErrCodeBadRequest ErrorCode = "BAD_REQUEST"
	// ErrCodeUnauthorized represents an unauthorized error.
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// ErrCodeForbidden represents a forbidden error.
	ErrCodeForbidden ErrorCode = "FORBIDDEN"
	// ErrCodeConflict represents a conflict error.
	ErrCodeConflict ErrorCode = "CONFLICT"
	// ErrCodeInternalError represents an internal error.
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
	// ErrCodeValidation represents a validation error.
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeDatabaseError represents a database error.
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"
	// ErrCodeDuplicateEntry represents a duplicate entry error.
	ErrCodeDuplicateEntry ErrorCode = "DUPLICATE_ENTRY"
)

// APIError represents an error that can be returned via API
type APIError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	StatusCode int       `json:"status_code"`
	Err        error     `json:"-"`
}

// Error implements error interface
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, statusCode int, err error) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// NewNotFoundError creates a NOT_FOUND error
func NewNotFoundError(message string) *APIError {
	return NewAPIError(ErrCodeNotFound, message, http.StatusNotFound, nil)
}

// NewBadRequestError creates a BAD_REQUEST error
func NewBadRequestError(message string) *APIError {
	return NewAPIError(ErrCodeBadRequest, message, http.StatusBadRequest, nil)
}

// NewUnauthorizedError creates an UNAUTHORIZED error
func NewUnauthorizedError(message string) *APIError {
	return NewAPIError(ErrCodeUnauthorized, message, http.StatusUnauthorized, nil)
}

// NewForbiddenError creates a FORBIDDEN error
func NewForbiddenError(message string) *APIError {
	return NewAPIError(ErrCodeForbidden, message, http.StatusForbidden, nil)
}

// NewConflictError creates a CONFLICT error
func NewConflictError(message string) *APIError {
	return NewAPIError(ErrCodeConflict, message, http.StatusConflict, nil)
}

// NewDuplicateEntryError creates a DUPLICATE_ENTRY error
func NewDuplicateEntryError(message string) *APIError {
	return NewAPIError(ErrCodeDuplicateEntry, message, http.StatusConflict, nil)
}

// NewValidationError creates a VALIDATION_ERROR
func NewValidationError(message string, err error) *APIError {
	return NewAPIError(ErrCodeValidation, message, http.StatusBadRequest, err)
}

// NewDatabaseError creates a DATABASE_ERROR
func NewDatabaseError(message string, err error) *APIError {
	return NewAPIError(ErrCodeDatabaseError, message, http.StatusInternalServerError, err)
}

// NewInternalError creates an INTERNAL_ERROR
func NewInternalError(message string, err error) *APIError {
	return NewAPIError(ErrCodeInternalError, message, http.StatusInternalServerError, err)
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// AsAPIError converts an error to an APIError if possible
func AsAPIError(err error) *APIError {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return NewInternalError("An unexpected error occurred", err)
}
