package domain

import (
	"context"

	"github.com/google/uuid"
)

// RoomReader defines read operations for Room entities.
type RoomReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Room, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Room, int, error)
	GetMembers(ctx context.Context, roomID uuid.UUID) ([]*RoomMember, error)
	IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
}

// RoomWriter defines write operations for Room entities.
type RoomWriter interface {
	Create(ctx context.Context, room *Room) error
	Update(ctx context.Context, room *Room) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RoomMembershipManager defines room membership operations.
type RoomMembershipManager interface {
	AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error
}

// RoomRepository combines read, write, and membership operations for Room entities.
type RoomRepository interface {
	RoomReader
	RoomWriter
	RoomMembershipManager
}

// MessageReader defines read operations for Message entities.
type MessageReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*Message, bool, error)
}

// MessageWriter defines write operations for Message entities.
type MessageWriter interface {
	Create(ctx context.Context, message *Message) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MessageRepository combines read and write operations for Message entities.
type MessageRepository interface {
	MessageReader
	MessageWriter
}
