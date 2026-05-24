package uuidgen

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// New generates a new UUIDv7 (or falls back to uuid.New).
func New() uuid.UUID {
	id, err := uuid.NewV7()
	if err == nil {
		return id
	}
	log.Warn().Err(err).Msg("Failed to generate UUIDv7, falling back to uuid.New()")
	return uuid.New()
}

// NewString generates a new UUIDv7 as a string (or falls back to uuid.NewString).
func NewString() string {
	id, err := uuid.NewV7()
	if err == nil {
		return id.String()
	}
	log.Warn().Err(err).Msg("Failed to generate UUIDv7, falling back to uuid.NewString()")
	return uuid.NewString()
}
