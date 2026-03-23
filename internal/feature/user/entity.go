// Package user provides domain entities and contracts for the user feature.
// This package follows clean/hexagonal architecture principles where the domain
// layer defines contracts (interfaces) that infrastructure components implement.
package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserID is a typed identifier for User entities.
type UserID string

// UserStatus represents the current state of a user.
type UserStatus string

const (
	// UserStatusActive indicates a user that can log in and use the system.
	UserStatusActive UserStatus = "active"
	// UserStatusInactive indicates a user that cannot log in but data is preserved.
	UserStatusInactive UserStatus = "inactive"
	// UserStatusSuspended indicates a user that has been temporarily blocked.
	UserStatusSuspended UserStatus = "suspended"
)

// IsValid checks if the status is a valid value.
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended:
		return true
	default:
		return false
	}
}

// User is the domain entity representing a system user.
type User struct {
	ID           UserID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser creates a new User with generated ID and timestamps.
func NewUser(email, passwordHash, firstName, lastName string) (*User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if passwordHash == "" {
		return nil, errors.New("password hash is required")
	}
	if firstName == "" {
		return nil, errors.New("first name is required")
	}
	if lastName == "" {
		return nil, errors.New("last name is required")
	}

	now := time.Now().UTC()
	return &User{
		ID:           UserID(uuid.New().String()),
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// NewUserWithID creates a User with specified values (for reconstruction from database).
func NewUserWithID(id UserID, email, passwordHash, firstName, lastName string, status UserStatus, createdAt, updatedAt time.Time) (*User, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if firstName == "" {
		return nil, errors.New("first name is required")
	}
	if lastName == "" {
		return nil, errors.New("last name is required")
	}
	if !status.IsValid() {
		return nil, errors.New("invalid user status")
	}

	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       status,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

// Activate sets user status to active.
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now().UTC()
}

// Deactivate sets user status to inactive.
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now().UTC()
}

// Suspend sets user status to suspended.
func (u *User) Suspend() {
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now().UTC()
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}
