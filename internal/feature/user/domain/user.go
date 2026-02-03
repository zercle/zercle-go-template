// Package domain contains the core business entities and domain logic for the user feature.
// This package has no external dependencies and defines the fundamental
// structures of the user domain.
package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user entity in the domain.
// It contains the core user data and business rules.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetID returns the user ID for interface compatibility.
func (u *User) GetID() string {
	return u.ID
}

// GetEmail returns the user email for interface compatibility.
func (u *User) GetEmail() string {
	return u.Email
}

// emailRegex is the regex pattern for email validation.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validate performs validation on the user entity.
// Returns nil if valid, otherwise returns an error describing the validation failure.
func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidID
	}
	if !IsValidEmail(u.Email) {
		return ErrInvalidEmail
	}
	if u.Name == "" {
		return ErrInvalidName
	}
	if len(u.Name) < 2 || len(u.Name) > 100 {
		return ErrInvalidNameLength
	}
	return nil
}

// SetPassword hashes and sets the user's password.
// The password must be at least 8 characters long.
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ErrPasswordHashFailed
	}

	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}

// VerifyPassword checks if the provided password matches the stored hash.
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// Update updates the user's fields and sets the updated timestamp.
// Only non-empty fields are updated.
func (u *User) Update(name string) {
	if name != "" {
		u.Name = name
	}
	u.UpdatedAt = time.Now()
}

// IsValidEmail validates an email address format.
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}

// NewUser creates a new user with the given email, name, and password.
// It generates a new UUID for the user and hashes the password.
func NewUser(email, name, password string) (*User, error) {
	if !IsValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if len(name) < 2 || len(name) > 100 {
		return nil, ErrInvalidNameLength
	}

	now := time.Now()
	user := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}

// Domain errors for the user entity.
// These are used for validation and business rule violations.
var (
	ErrInvalidID          = NewDomainError("INVALID_ID", "user ID is required")
	ErrInvalidEmail       = NewDomainError("INVALID_EMAIL", "email format is invalid")
	ErrInvalidName        = NewDomainError("INVALID_NAME", "name is required")
	ErrInvalidNameLength  = NewDomainError("INVALID_NAME_LENGTH", "name must be between 2 and 100 characters")
	ErrPasswordTooShort   = NewDomainError("PASSWORD_TOO_SHORT", "password must be at least 8 characters")
	ErrPasswordHashFailed = NewDomainError("PASSWORD_HASH_FAILED", "failed to hash password")
	ErrUserNotFound       = NewDomainError("USER_NOT_FOUND", "user not found")
	ErrDuplicateEmail     = NewDomainError("DUPLICATE_EMAIL", "email already exists")
)

// DomainError represents a domain-specific error.
type DomainError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}
