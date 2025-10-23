package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system (Domain Entity).
type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
