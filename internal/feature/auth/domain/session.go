package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authentication session.
type Session struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}
