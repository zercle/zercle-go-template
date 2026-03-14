package uuidgen

import (
	"log/slog"

	"github.com/google/uuid"
)

// New generates a new UUIDv7.
func New() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		slog.Warn("Failed to generate UUIDv7, falling back to uuid.New()", "err", err)
		return uuid.New()
	}
	return id
}

// NewString generates a new UUIDv7 as a string.
func NewString() string {
	id, err := uuid.NewV7()
	if err != nil {
		slog.Warn("Failed to generate UUIDv7, falling back to uuid.NewString()", "err", err)
		return uuid.NewString()
	}
	return id.String()
}
