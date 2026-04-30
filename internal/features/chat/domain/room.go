package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// Room represents a chat room with members and metadata.
type Room struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	MemberCount int        `json:"member_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewRoom creates a new room with generated UUID and timestamps.
func NewRoom(name, description, roomType string, ownerID uuid.UUID) *Room {
	now := time.Now()
	return &Room{
		ID:          uuidgen.New(),
		Name:        name,
		Description: description,
		Type:        roomType,
		OwnerID:     ownerID,
		MemberCount: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate checks room name and type validity.
func (r *Room) Validate() error {
	if r.Name == "" {
		return ErrRoomNameRequired
	}
	if r.Type != "public" && r.Type != "private" && r.Type != "direct" {
		return ErrInvalidRoomType
	}
	return nil
}

// RoomMember represents a user's membership in a room with role and metadata.
type RoomMember struct {
	RoomID      uuid.UUID `json:"room_id"`
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// ErrRoomNameRequired is returned when room name is empty.
var (
	ErrRoomNameRequired = NewError("room name is required")
	ErrInvalidRoomType  = NewError("invalid room type")
)

// NewError creates a new ValidationError with the given message.
func NewError(message string) error {
	return &ValidationError{Message: message}
}

// ValidationError represents a room validation failure.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
