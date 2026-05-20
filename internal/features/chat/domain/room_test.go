package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestRoom_Validate_EmptyName(t *testing.T) {
	t.Parallel()
	room := &Room{
		ID:   uuid.New(),
		Name: "",
		Type: RoomTypePublic,
	}
	err := room.Validate()
	if !errors.Is(err, ErrRoomNameRequired) {
		t.Errorf("expected ErrRoomNameRequired, got %v", err)
	}
}

func TestRoom_Validate_InvalidType(t *testing.T) {
	t.Parallel()
	room := &Room{
		ID:   uuid.New(),
		Name: "test-room",
		Type: "invalid",
	}
	err := room.Validate()
	if !errors.Is(err, ErrInvalidRoomType) {
		t.Errorf("expected ErrInvalidRoomType, got %v", err)
	}
}

func TestRoom_Validate_PublicRoom(t *testing.T) {
	t.Parallel()
	room := &Room{
		ID:   uuid.New(),
		Name: "test-room",
		Type: RoomTypePublic,
	}
	if err := room.Validate(); err != nil {
		t.Errorf("expected nil for public room, got %v", err)
	}
}

func TestRoom_Validate_PrivateRoom(t *testing.T) {
	t.Parallel()
	room := &Room{
		ID:   uuid.New(),
		Name: "test-room",
		Type: RoomTypePrivate,
	}
	if err := room.Validate(); err != nil {
		t.Errorf("expected nil for private room, got %v", err)
	}
}

func TestRoom_Validate_DirectRoom(t *testing.T) {
	t.Parallel()
	room := &Room{
		ID:   uuid.New(),
		Name: "test-room",
		Type: RoomTypeDirect,
	}
	if err := room.Validate(); err != nil {
		t.Errorf("expected nil for direct room, got %v", err)
	}
}

func TestNewRoom(t *testing.T) {
	t.Parallel()
	ownerID := uuid.New()
	room := NewRoom("test-room", "description", RoomTypePublic, ownerID)

	if room.Name != "test-room" {
		t.Errorf("expected Name=test-room, got %s", room.Name)
	}
	if room.Description != "description" {
		t.Errorf("expected Description=description, got %s", room.Description)
	}
	if room.Type != RoomTypePublic {
		t.Errorf("expected Type=public, got %s", room.Type)
	}
	if room.OwnerID != ownerID {
		t.Errorf("expected OwnerID=%s, got %s", ownerID, room.OwnerID)
	}
	if room.MemberCount != 1 {
		t.Errorf("expected MemberCount=1, got %d", room.MemberCount)
	}
	if room.ID == uuid.Nil {
		t.Error("expected generated UUID, got Nil")
	}
	if room.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if room.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}
