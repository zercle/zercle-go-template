package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

// Room represents a chat room.
type Room struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Type        string     `json:"type" db:"type"`
	OwnerID     uuid.UUID  `json:"owner_id" db:"owner_id"`
	MemberCount int        `json:"member_count" db:"member_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// NewRoom creates a new room instance.
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

// Validate validates the room data.
func (r *Room) Validate() error {
	if r.Name == "" {
		return ErrRoomNameRequired
	}
	if r.Type != "public" && r.Type != "private" && r.Type != "direct" {
		return ErrInvalidRoomType
	}
	return nil
}

// RoomMember represents a member of a chat room.
type RoomMember struct {
	RoomID      uuid.UUID `json:"room_id" db:"room_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Username    string    `json:"username" db:"username"`
	DisplayName string    `json:"display_name" db:"display_name"`
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`
	Role        string    `json:"role" db:"role"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}
