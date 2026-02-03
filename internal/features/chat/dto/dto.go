package dto

import "github.com/google/uuid"

type CreateRoomRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Description string   `json:"description"`
	Type        string   `json:"type" validate:"required,oneof=public private direct"`
	MemberIDs   []string `json:"member_ids"`
}

type UpdateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoomResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	OwnerID     uuid.UUID `json:"owner_id"`
	MemberCount int       `json:"member_count"`
	CreatedAt   string    `json:"created_at"`
}

type ListRoomsResponse struct {
	Rooms []*RoomResponse `json:"rooms"`
	Total int             `json:"total"`
}

type SendMessageRequest struct {
	Content     string `json:"content" validate:"required"`
	MessageType string `json:"message_type"`
	ReplyTo     string `json:"reply_to"`
}

type MessageResponse struct {
	ID             uuid.UUID `json:"id"`
	RoomID         uuid.UUID `json:"room_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	SenderUsername string    `json:"sender_username"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	ReplyTo        string    `json:"reply_to,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

type GetMessagesResponse struct {
	Messages []*MessageResponse `json:"messages"`
	HasMore  bool               `json:"has_more"`
}
