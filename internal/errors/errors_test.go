package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
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

// Tests for Wrap function.

func TestWrap(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
		want    bool // want non-nil error
	}{
		{
			name:    "wrap standard error",
			err:     errors.New("underlying error"),
			message: "context message",
			want:    true,
		},
		{
			name:    "wrap AppError",
			err:     ValidationError("invalid input"),
			message: "validation failed",
			want:    true,
		},
		{
			name:    "nil error returns nil",
			err:     nil,
			message: "some message",
			want:    false,
		},
		{
			name:    "wrap wrapped error",
			err:     Wrap(errors.New("original"), "first context"),
			message: "second context",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.err, tt.message)
			if tt.want && got == nil {
				t.Errorf("Wrap() = nil, want non-nil error")
			}
			if !tt.want && got != nil {
				t.Errorf("Wrap() = %v, want nil", got)
			}
			// Verify the error message contains the context
			if tt.want && got != nil {
				if !strings.Contains(got.Error(), tt.message) {
					t.Errorf("Wrap() error message = %v, want to contain %v", got.Error(), tt.message)
				}
			}
		})
	}
}

func TestWrap_Unwrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "context message")

	// Verify we can unwrap to get the original
	if !errors.Is(wrapped, original) {
		t.Error("Wrap() should support errors.Is for original error")
	}
}

func TestWrap_AppError(t *testing.T) {
	appErr := ValidationError("invalid email")
	wrapped := Wrap(appErr, "user validation failed")

	// Verify we can find AppError in the chain
	var found *AppError
	if !errors.As(wrapped, &found) {
		t.Error("Wrap() should support errors.As for AppError")
	}
	if found.Code != ErrCodeValidation {
		t.Errorf("Found AppError Code = %v, want %v", found.Code, ErrCodeValidation)
	}
}

// Tests for Is helper function.

func TestIs(t *testing.T) {
	sentinel := errors.New("sentinel error")

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "matches sentinel",
			err:    sentinel,
			target: sentinel,
			want:   true,
		},
		{
			name:   "does not match",
			err:    errors.New("different error"),
			target: sentinel,
			want:   false,
		},
		{
			name:   "wrapped sentinel",
			err:    fmt.Errorf("context: %w", sentinel),
			target: sentinel,
			want:   true,
		},
		{
			name:   "nil error",
			err:    nil,
			target: sentinel,
			want:   false,
		},
		{
			name:   "nil target",
			err:    sentinel,
			target: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIs_AppError(t *testing.T) {
	err1 := New(ErrCodeValidation, "error 1", http.StatusBadRequest)
	err2 := New(ErrCodeValidation, "error 2", http.StatusBadRequest)
	err3 := New(ErrCodeNotFound, "error 3", http.StatusNotFound)

	// Test matching by AppError
	if !Is(err1, err2) {
		t.Error("Is() should match AppErrors with same code")
	}
	if Is(err1, err3) {
		t.Error("Is() should not match AppErrors with different codes")
	}
}

// Tests for As helper function.

func TestAs(t *testing.T) {
	appErr := ValidationError("invalid input")
	wrapped := Wrap(appErr, "validation context")

	// Test using addressable variable
	var target *AppError
	if !As(wrapped, &target) {
		t.Error("As() should find AppError in wrapped error chain")
	}
	if target.Code != ErrCodeValidation {
		t.Errorf("Found AppError Code = %v, want %v", target.Code, ErrCodeValidation)
	}
}

func TestAs_NotFound(t *testing.T) {
	regularErr := errors.New("regular error")

	var target *AppError
	if As(regularErr, &target) {
		t.Error("As() should not find AppError in regular error")
	}
}

func TestAs_NilError(t *testing.T) {
	var target *AppError
	if As(nil, &target) {
		t.Error("As() should return false for nil error")
	}
}

// Test that Wrap handles nil gracefully.
func TestWrap_NilError(t *testing.T) {
	result := Wrap(nil, "some context")
	if result != nil {
		t.Errorf("Wrap(nil, _) = %v, want nil", result)
	}
}

// Test error chain with multiple wraps.
func TestWrap_MultipleWraps(t *testing.T) {
	original := errors.New("original")
	level1 := Wrap(original, "level 1")
	level2 := Wrap(level1, "level 2")
	level3 := Wrap(level2, "level 3")

	// Verify we can find original at any level
	if !errors.Is(level3, original) {
		t.Error("errors.Is should find original error in deep chain")
	}

	var found *AppError
	if errors.As(level3, &found) {
		t.Logf("Found AppError: %v", found)
	}
}
