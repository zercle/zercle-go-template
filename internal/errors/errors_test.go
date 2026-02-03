package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "error without details",
			err:      New(ErrCodeValidation, "validation failed", http.StatusBadRequest),
			expected: "[VALIDATION_ERROR] validation failed",
		},
		{
			name:     "error with details",
			err:      New(ErrCodeNotFound, "resource not found", http.StatusNotFound).WithDetails("user with id 123"),
			expected: "[NOT_FOUND] resource not found: user with id 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_WithCause(t *testing.T) {
	cause := errors.New("underlying error")
	appErr := New(ErrCodeInternal, "operation failed", http.StatusInternalServerError).WithCause(cause)

	if appErr.Cause != cause {
		t.Error("WithCause() should set the cause")
	}

	if !errors.Is(appErr, cause) {
		t.Error("Unwrap() should return the cause")
	}
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := New(ErrCodeValidation, "validation failed", http.StatusBadRequest).WithDetails("field: email")

	if appErr.Details != "field: email" {
		t.Errorf("WithDetails() Details = %v, want %v", appErr.Details, "field: email")
	}
}

func TestAppError_Is(t *testing.T) {
	tests := []struct {
		name string
		err1 *AppError
		err2 error
		want bool
	}{
		{
			name: "same code",
			err1: New(ErrCodeValidation, "error 1", http.StatusBadRequest),
			err2: New(ErrCodeValidation, "error 2", http.StatusBadRequest),
			want: true,
		},
		{
			name: "different code",
			err1: New(ErrCodeValidation, "error 1", http.StatusBadRequest),
			err2: New(ErrCodeNotFound, "error 2", http.StatusNotFound),
			want: false,
		},
		{
			name: "not an AppError",
			err1: New(ErrCodeValidation, "error 1", http.StatusBadRequest),
			err2: errors.New("regular error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err1.Is(tt.err2); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name           string
		constructor    func(string) *AppError
		message        string
		expectedCode   ErrorCode
		expectedStatus int
	}{
		{
			name:           "ValidationError",
			constructor:    ValidationError,
			message:        "field required",
			expectedCode:   ErrCodeValidation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "NotFoundError",
			constructor:    NotFoundError,
			message:        "user",
			expectedCode:   ErrCodeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ConflictError",
			constructor:    ConflictError,
			message:        "duplicate entry",
			expectedCode:   ErrCodeConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "InternalError",
			constructor:    InternalError,
			message:        "database error",
			expectedCode:   ErrCodeInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "UnauthorizedError",
			constructor:    UnauthorizedError,
			message:        "invalid token",
			expectedCode:   ErrCodeUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ForbiddenError",
			constructor:    ForbiddenError,
			message:        "access denied",
			expectedCode:   ErrCodeForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "BadRequestError",
			constructor:    BadRequestError,
			message:        "malformed request",
			expectedCode:   ErrCodeBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor(tt.message)
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.expectedCode)
			}
			if err.StatusCode != tt.expectedStatus {
				t.Errorf("StatusCode = %v, want %v", err.StatusCode, tt.expectedStatus)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "validation error",
			err:  ValidationError("invalid input"),
			want: true,
		},
		{
			name: "not found error",
			err:  NotFoundError("user"),
			want: false,
		},
		{
			name: "wrapped validation error",
			err:  errors.New("wrapped: " + ValidationError("invalid input").Error()),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.want {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found error",
			err:  NotFoundError("user"),
			want: true,
		},
		{
			name: "validation error",
			err:  ValidationError("invalid input"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "conflict error",
			err:  ConflictError("duplicate"),
			want: true,
		},
		{
			name: "validation error",
			err:  ValidationError("invalid input"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConflictError(tt.err); got != tt.want {
				t.Errorf("IsConflictError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "nil error",
			err:  nil,
			want: http.StatusOK,
		},
		{
			name: "validation error",
			err:  ValidationError("invalid"),
			want: http.StatusBadRequest,
		},
		{
			name: "not found error",
			err:  NotFoundError("user"),
			want: http.StatusNotFound,
		},
		{
			name: "internal error",
			err:  InternalError("oops"),
			want: http.StatusInternalServerError,
		},
		{
			name: "regular error",
			err:  errors.New("some error"),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStatusCode(tt.err); got != tt.want {
				t.Errorf("GetStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
