package chat

// CreateRoomRequest represents a request to create a chat room.
type CreateRoomRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	MemberIDs   []string `json:"member_ids"`
}

// UpdateRoomRequest represents a request to update a chat room.
type UpdateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SendMessageRequest represents a request to send a message.
type SendMessageRequest struct {
	Content     string  `json:"content"`
	MessageType string  `json:"message_type"`
	ReplyTo     *string `json:"reply_to"`
}
