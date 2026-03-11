package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

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

func (m *Message) Validate() error {
	if m.Content == "" {
		return ErrMessageContentRequired
	}
	return nil
}

var ErrMessageContentRequired = NewError("message content is required")
