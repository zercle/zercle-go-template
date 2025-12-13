package domain

import "github.com/google/uuid"

// UserID is a type-safe wrapper for user identifiers.
// This enables cross-feature references (e.g., Post → User) without creating import cycles.
type UserID uuid.UUID

// UserRef provides a common interface for user references across features.
type UserRef interface {
	GetID() UserID
}

// NewUserID creates a new UserID from a uuid.UUID.
func NewUserID(id uuid.UUID) UserID {
	return UserID(id)
}

// String converts UserID to string.
func (id UserID) String() string {
	return uuid.UUID(id).String()
}

// UUID converts UserID to uuid.UUID.
func (id UserID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

// MarshalText implements encoding.TextMarshaler.
func (id UserID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *UserID) UnmarshalText(text []byte) error {
	var uuidVal uuid.UUID
	if err := uuidVal.UnmarshalText(text); err != nil {
		return err
	}
	*id = UserID(uuidVal)
	return nil
}
