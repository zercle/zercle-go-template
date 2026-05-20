package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestMessage_Validate_EmptyContent(t *testing.T) {
	t.Parallel()
	msg := &Message{
		ID:      uuid.New(),
		Content: "",
	}
	err := msg.Validate()
	if !errors.Is(err, ErrMessageContentRequired) {
		t.Errorf("expected ErrMessageContentRequired, got %v", err)
	}
}

func TestMessage_Validate_ValidContent(t *testing.T) {
	t.Parallel()
	msg := &Message{
		ID:      uuid.New(),
		Content: "hello world",
	}
	if err := msg.Validate(); err != nil {
		t.Errorf("expected nil for valid message, got %v", err)
	}
}

func TestNewMessage(t *testing.T) {
	t.Parallel()
	roomID := uuid.New()
	senderID := uuid.New()
	msg := NewMessage(roomID, senderID, "test content", "text")

	if msg.RoomID != roomID {
		t.Errorf("expected RoomID=%s, got %s", roomID, msg.RoomID)
	}
	if msg.SenderID != senderID {
		t.Errorf("expected SenderID=%s, got %s", senderID, msg.SenderID)
	}
	if msg.Content != "test content" {
		t.Errorf("expected Content=test content, got %s", msg.Content)
	}
	if msg.MessageType != "text" {
		t.Errorf("expected MessageType=text, got %s", msg.MessageType)
	}
	if msg.ID == uuid.Nil {
		t.Error("expected generated UUID, got Nil")
	}
	if msg.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if msg.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}
