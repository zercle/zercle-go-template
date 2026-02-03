package user

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrDuplicateEmail is returned when an email already exists.
	ErrDuplicateEmail = errors.New("email already exists")
	// ErrInvalidEmail is returned when an email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidPassword is returned when a password is invalid.
	ErrInvalidPassword = errors.New("invalid password")
)
