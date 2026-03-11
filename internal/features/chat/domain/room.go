package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/pkg/uuidgen"
)

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

func (r *Room) Validate() error {
	if r.Name == "" {
		return ErrRoomNameRequired
	}
	if r.Type != "public" && r.Type != "private" && r.Type != "direct" {
		return ErrInvalidRoomType
	}
	return nil
}

type RoomMember struct {
	RoomID      uuid.UUID `json:"room_id"`
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

var (
	ErrRoomNameRequired = NewError("room name is required")
	ErrInvalidRoomType  = NewError("invalid room type")
)

func NewError(message string) error {
	return &ValidationError{Message: message}
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
