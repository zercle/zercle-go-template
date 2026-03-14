package domain

// MessageEvent represents a real-time message event.
type MessageEvent struct {
	Type      string `json:"type"`
	RoomID    string `json:"room_id"`
	MessageID string `json:"message_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// PresenceEvent represents a user presence status change.
type PresenceEvent struct {
	Type     string `json:"type"`
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Status   string `json:"status"`
	Username string `json:"username"`
}

// TypingEvent represents a typing indicator event.
type TypingEvent struct {
	Type     string `json:"type"`
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Typing   bool   `json:"typing"`
}
