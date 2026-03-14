package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// Message represents a chat message.
type Message struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	RoomID         uuid.UUID  `json:"room_id" db:"room_id"`
	SenderID       uuid.UUID  `json:"sender_id" db:"sender_id"`
	SenderUsername string     `json:"sender_username" db:"sender_username"`
	Content        string     `json:"content" db:"content"`
	MessageType    string     `json:"message_type" db:"message_type"`
	ReplyTo        *uuid.UUID `json:"reply_to,omitempty" db:"reply_to"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// NewMessage creates a new message instance.
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

// Validate validates the message data.
func (m *Message) Validate() error {
	if m.Content == "" {
		return ErrMessageContentRequired
	}
	return nil
}
