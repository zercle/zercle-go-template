package uuidgen

import (
	"testing"

	"github.com/google/uuid"
)

func TestNew_ReturnsNonNilUUID(t *testing.T) {
	t.Parallel()
	id := New()
	if id == uuid.Nil {
		t.Errorf("expected non-nil UUID, got Nil")
	}
}

func TestNewString_ReturnsNonEmptyString(t *testing.T) {
	t.Parallel()
	s := NewString()
	if s == "" {
		t.Error("expected non-empty string, got empty")
	}
}

func TestNew_ReturnsV7(t *testing.T) {
	t.Parallel()
	id := New()
	version := id.Version()
	if version != 7 {
		t.Errorf("expected UUID version 7, got %d", version)
	}
}
