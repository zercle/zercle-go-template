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

// Error returns the human-readable message.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the causal error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// DomainError is a marker interface that domain sentinel errors may optionally
// implement so that mappers can recognise them generically.
type DomainError interface {
	error
	DomainError()
}

// Sentinel boundary errors. These are the shared error responses returned when
// a domain or infrastructure error cannot be mapped to a feature-specific
// sentinel.
var (
	ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "resource not found", HTTPStatus: http.StatusNotFound, GRPCCode: codes.NotFound}
	ErrInvalidInput = &AppError{Code: "INVALID_INPUT", Message: "invalid input", HTTPStatus: http.StatusBadRequest, GRPCCode: codes.InvalidArgument}
	ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized", HTTPStatus: http.StatusUnauthorized, GRPCCode: codes.Unauthenticated}
	ErrForbidden    = &AppError{Code: "FORBIDDEN", Message: "forbidden", HTTPStatus: http.StatusForbidden, GRPCCode: codes.PermissionDenied}
	ErrConflict     = &AppError{Code: "CONFLICT", Message: "conflict", HTTPStatus: http.StatusConflict, GRPCCode: codes.AlreadyExists}
	ErrInternal     = &AppError{Code: "INTERNAL", Message: "internal error", HTTPStatus: http.StatusInternalServerError, GRPCCode: codes.Internal}
)
