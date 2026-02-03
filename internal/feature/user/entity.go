// Package user provides domain entities and contracts for the user feature.
// This package follows clean/hexagonal architecture principles where the domain
// layer defines contracts (interfaces) that infrastructure components implement.
package user

import (
	"errors"
	"time"

	"github.com/zercle/zercle-go-template/pkg/uid"
)

// ID is a typed identifier for User entities.
type ID string

// Status represents the current state of a user.
type Status string

const (
	// StatusActive indicates a user that can log in and use the system.
	StatusActive Status = "active"
	// StatusInactive indicates a user that cannot log in but data is preserved.
	StatusInactive Status = "inactive"
	// StatusSuspended indicates a user that has been temporarily blocked.
	StatusSuspended Status = "suspended"
)

// IsValid checks if the status is a valid value.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusSuspended:
		return true
	default:
		return false
	}
}

// User is the domain entity representing a system user.
type User struct {
	ID           ID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// New creates a new User with generated ID and timestamps.
func New(email, passwordHash, firstName, lastName string) (*User, error) {
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
		ID:           ID(uid.New().String()),
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// NewWithID creates a User with specified values (for reconstruction from database).
func NewWithID(id ID, email, passwordHash, firstName, lastName string, status Status, createdAt, updatedAt time.Time) (*User, error) {
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
	u.Status = StatusActive
	u.UpdatedAt = time.Now().UTC()
}

// Deactivate sets user status to inactive.
func (u *User) Deactivate() {
	u.Status = StatusInactive
	u.UpdatedAt = time.Now().UTC()
}

// Suspend sets user status to suspended.
func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now().UTC()
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}
