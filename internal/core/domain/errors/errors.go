// Package domerrors defines domain-specific errors.
package domerrors

import "errors"

var (
	// ErrNotFound is returned when a record is not found in the database.
	ErrNotFound = errors.New("record not found")
	// ErrDuplicate is returned when a unique constraint is violated.
	ErrDuplicate = errors.New("record already exists")
	// ErrInvalidCreds is returned when authentication fails.
	ErrInvalidCreds = errors.New("invalid credentials")
	// ErrUnauthorized is returned when authorization fails.
	ErrUnauthorized = errors.New("unauthorized action")
	// ErrInternalServer is returned when an unexpected error occurs.
	ErrInternalServer = errors.New("internal server error")
)
