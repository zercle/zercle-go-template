package domain

import (
	"context"

	"github.com/google/uuid"
)

type RoomReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Room, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Room, int, error)
	GetMembers(ctx context.Context, roomID uuid.UUID) ([]*RoomMember, error)
	IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
}

type RoomWriter interface {
	Create(ctx context.Context, room *Room) error
	Update(ctx context.Context, room *Room) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type RoomMembershipManager interface {
	AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error
}

type RoomRepository interface {
	RoomReader
	RoomWriter
	RoomMembershipManager
}

type MessageReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*Message, bool, error)
}

type MessageWriter interface {
	Create(ctx context.Context, message *Message) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type MessageRepository interface {
	MessageReader
	MessageWriter
}
