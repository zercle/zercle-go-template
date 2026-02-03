// Package uid provides UUID generation utilities optimized for database B-tree indexes.
// Uses UUIDv7 for better performance compared to UUIDv4 due to time-ordered generation.
package uid

import (
	"github.com/google/uuid"
)

// New generates a new UUIDv7 identifier.
// UUIDv7 is preferred over UUIDv4 for database primary keys because:
//   - Time-ordered generation improves B-tree index locality
//   - Reduces index fragmentation in PostgreSQL
//   - Better sequential write performance
func New() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		// This should never happen, but if it does, fall back to V4
		return uuid.New()
	}
	return id
}

// Parse parses a UUID from its string representation.
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a UUID from its string representation.
// Panics if the string is not a valid UUID.
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// FromString is an alias for Parse.
func FromString(s string) (uuid.UUID, error) {
	return Parse(s)
}

// String returns the string representation of a UUID.
func String(id uuid.UUID) string {
	return id.String()
}

// IsValid checks if a string is a valid UUID.
func IsValid(s string) bool {
	_, err := Parse(s)
	return err == nil
}
