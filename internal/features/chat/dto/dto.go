package dto

import "github.com/google/uuid"

// CreateRoomRequest represents the create room request body.
// swagger:model
type CreateRoomRequest struct {
	// The name of the room
	// Required: true
	// Min Length: 1
	// Max Length: 100
	Name string `json:"name" validate:"required,min=1,max=100"`
	// The description of the room
	Description string `json:"description"`
	// The type of room (public, private, direct)
	// Required: true
	// Enum: public, private, direct
	Type string `json:"type" validate:"required,oneof=public private direct"`
	// List of user IDs to add as initial members
	// Format: uuid
	MemberIDs []string `json:"member_ids"`
}

// UpdateRoomRequest represents the update room request body.
// swagger:model
type UpdateRoomRequest struct {
	// The new name of the room
	Name string `json:"name"`
	// The new description of the room
	Description string `json:"description"`
}

// RoomResponse represents a chat room in responses.
// swagger:model
type RoomResponse struct {
	// The unique room identifier
	// Format: uuid
	ID uuid.UUID `json:"id"`
	// The name of the room
	Name string `json:"name"`
	// The description of the room
	Description string `json:"description"`
	// The type of room (public, private, direct)
	Type string `json:"type"`
	// The ID of the room owner
	// Format: uuid
	OwnerID uuid.UUID `json:"owner_id"`
	// The number of members in the room
	MemberCount int `json:"member_count"`
	// When the room was created
	// Format: date-time
	CreatedAt string `json:"created_at"`
}

// ListRoomsResponse represents a list of rooms.
// swagger:model
type ListRoomsResponse struct {
	// The list of rooms
	Rooms []*RoomResponse `json:"rooms"`
	// Total number of rooms
	Total int `json:"total"`
}

// SendMessageRequest represents the send message request body.
// swagger:model
type SendMessageRequest struct {
	// The message content
	// Required: true
	Content string `json:"content" validate:"required"`
	// The type of message (text, image, file, etc.)
	MessageType string `json:"message_type"`
	// ID of the message being replied to (optional)
	// Format: uuid
	ReplyTo string `json:"reply_to"`
}

// MessageResponse represents a chat message in responses.
// swagger:model
type MessageResponse struct {
	// The unique message identifier
	// Format: uuid
	ID uuid.UUID `json:"id"`
	// The ID of the room this message belongs to
	// Format: uuid
	RoomID uuid.UUID `json:"room_id"`
	// The ID of the message sender
	// Format: uuid
	SenderID uuid.UUID `json:"sender_id"`
	// The username of the message sender
	SenderUsername string `json:"sender_username"`
	// The message content
	Content string `json:"content"`
	// The type of message
	MessageType string `json:"message_type"`
	// ID of the message being replied to (if any)
	ReplyTo string `json:"reply_to,omitempty"`
	// When the message was sent
	// Format: date-time
	CreatedAt string `json:"created_at"`
}

// GetMessagesResponse represents a paginated list of messages.
// swagger:model
type GetMessagesResponse struct {
	// The list of messages
	Messages []*MessageResponse `json:"messages"`
	// Whether there are more messages available
	HasMore bool `json:"has_more"`
}
