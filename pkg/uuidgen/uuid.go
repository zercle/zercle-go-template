package uuidgen

import (
	"log/slog"

	"github.com/google/uuid"
)

// New generates a new UUIDv7 (or falls back to uuid.New).
func New() uuid.UUID {
	id, err := uuid.NewV7()
	if err == nil {
		return id
	}
	slog.Warn("Failed to generate UUIDv7, falling back to uuid.New()", "err", err)
	return uuid.New()
}

// NewString generates a new UUIDv7 as a string (or falls back to uuid.NewString).
func NewString() string {
	id, err := uuid.NewV7()
	if err == nil {
		return id.String()
	}
	slog.Warn("Failed to generate UUIDv7, falling back to uuid.NewString()", "err", err)
	return uuid.NewString()
}
