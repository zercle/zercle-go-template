package chat

import "time"

// RoomResponse represents room data in API responses.
type RoomResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	OwnerID     string    `json:"owner_id"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MemberResponse represents room member data in API responses.
type MemberResponse struct {
	RoomID      string    `json:"room_id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// MessageResponse represents message data in API responses.
type MessageResponse struct {
	ID             string    `json:"id"`
	RoomID         string    `json:"room_id"`
	SenderID       string    `json:"sender_id"`
	SenderUsername string    `json:"sender_username"`
	Content        string    `json:"content"`
	MessageType    string    `json:"message_type"`
	ReplyTo        *string   `json:"reply_to"`
	CreatedAt      time.Time `json:"created_at"`
}

// MessageHistoryResponse represents a list of messages with pagination info.
type MessageHistoryResponse struct {
	Messages []*MessageResponse `json:"messages"`
	HasMore  bool               `json:"has_more"`
}

// RoomListResponse represents a paginated list of rooms.
type RoomListResponse struct {
	Rooms  []*RoomResponse `json:"rooms"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}
