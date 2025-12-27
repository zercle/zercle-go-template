package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity
type User struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	FullName  string    `json:"full_name" db:"full_name"`
	Phone     string    `json:"phone,omitempty" db:"phone"`
	ID        uuid.UUID `json:"id" db:"id"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}
