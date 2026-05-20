package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/shared/errors"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// Message represents a chat message within a room.
type Message struct {
	ID             uuid.UUID  `json:"id"`
	RoomID         uuid.UUID  `json:"room_id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	SenderUsername string     `json:"sender_username"`
	Content        string     `json:"content"`
	MessageType    string     `json:"message_type"`
	ReplyTo        *uuid.UUID `json:"reply_to,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewMessage creates a new message with generated UUID and timestamps.
func NewMessage(roomID, senderID uuid.UUID, content, messageType string) *Message {
	now := time.Now()
	return &Message{
		ID:          uuidgen.New(),
		RoomID:      roomID,
		SenderID:    senderID,
		Content:     content,
		MessageType: messageType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate checks if the message content is not empty.
func (m *Message) Validate() error {
	if m.Content == "" {
		return ErrMessageContentRequired
	}
	return nil
}

// ErrMessageContentRequired is returned when message content is empty.
var ErrMessageContentRequired = errors.ErrMessageContentRequired
