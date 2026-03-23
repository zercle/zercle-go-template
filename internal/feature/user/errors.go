package user

import "errors"

// Domain errors for the user feature.
var (
	// ErrUserNotFound indicates the requested user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrDuplicateEmail indicates an email is already registered.
	ErrDuplicateEmail = errors.New("email already registered")

	// ErrInvalidCredentials indicates the provided credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserInactive indicates the user account is inactive.
	ErrUserInactive = errors.New("user account is inactive")

	// ErrInvalidEmail indicates email format is invalid.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidPassword indicates password is empty or invalid.
	ErrInvalidPassword = errors.New("invalid password")
)

// DomainError is a sentinel error for user domain errors that supports errors.Is comparison.
type DomainError struct {
	Sentinel error
}

func (e *DomainError) Error() string {
	return e.Sentinel.Error()
}

func (e *DomainError) Is(target error) bool {
	return errors.Is(e.Sentinel, target)
}

// NewDomainError creates a new domain error wrapping a sentinel error.
func NewDomainError(sentinel error) *DomainError {
	return &DomainError{Sentinel: sentinel}
}
