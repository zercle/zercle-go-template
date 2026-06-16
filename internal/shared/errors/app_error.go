// Package errors provides a typed, shared boundary error type and mappers to
// HTTP and gRPC status codes.
package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// AppError is the shared boundary error type. It carries enough metadata for
// transport-agnostic handlers to translate domain failures into HTTP or gRPC
// responses without string matching.
type AppError struct {
	// Code is a stable, machine-readable error code (e.g. NOT_FOUND).
	Code string
	// Message is a human-readable description.
	Message string
	// HTTPStatus is the HTTP status code that should be returned.
	HTTPStatus int
	// GRPCCode is the gRPC status code that should be returned.
	GRPCCode codes.Code
	// Cause is the underlying error, if any, preserved for observability.
	Cause error
}

// Error returns the human-readable message, falling back to the machine-readable
// Code when Message is empty so the error string is never empty.
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// Unwrap returns the causal error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Sentinel boundary errors. These are the shared error responses returned when
// a domain or infrastructure error cannot be mapped to a feature-specific
// sentinel.
var (
	ErrNotFound         = &AppError{Code: "NOT_FOUND", Message: "resource not found", HTTPStatus: http.StatusNotFound, GRPCCode: codes.NotFound}
	ErrInvalidInput     = &AppError{Code: "INVALID_INPUT", Message: "invalid input", HTTPStatus: http.StatusBadRequest, GRPCCode: codes.InvalidArgument}
	ErrUnauthorized     = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized", HTTPStatus: http.StatusUnauthorized, GRPCCode: codes.Unauthenticated}
	ErrForbidden        = &AppError{Code: "FORBIDDEN", Message: "forbidden", HTTPStatus: http.StatusForbidden, GRPCCode: codes.PermissionDenied}
	ErrConflict         = &AppError{Code: "CONFLICT", Message: "conflict", HTTPStatus: http.StatusConflict, GRPCCode: codes.AlreadyExists}
	ErrCanceled         = &AppError{Code: "CANCELED", Message: "request canceled", HTTPStatus: 499, GRPCCode: codes.Canceled}
	ErrDeadlineExceeded = &AppError{Code: "DEADLINE_EXCEEDED", Message: "deadline exceeded", HTTPStatus: http.StatusGatewayTimeout, GRPCCode: codes.DeadlineExceeded}
	ErrInternal         = &AppError{Code: "INTERNAL", Message: "internal error", HTTPStatus: http.StatusInternalServerError, GRPCCode: codes.Internal}
)
