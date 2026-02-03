package uuidgen

import (
	"log/slog"

	"github.com/google/uuid"
)

func New() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	} else {
		slog.Warn("Failed to generate UUIDv7, falling back to uuid.New()", "err", err)
		return uuid.New()
	}
}

func NewString() string {
	if id, err := uuid.NewV7(); err == nil {
		return id.String()
	} else {
		slog.Warn("Failed to generate UUIDv7, falling back to uuid.NewString()", "err", err)
		return uuid.NewString()
	}
}
