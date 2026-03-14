package domain

import "errors"

// Domain-specific error variables.
var (
	ErrUsernameRequired   = errors.New("username is required")
	ErrEmailRequired      = errors.New("email is required")
	ErrPasswordRequired   = errors.New("password is required")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
)
