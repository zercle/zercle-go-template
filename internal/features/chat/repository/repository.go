package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// RoomReader defines read operations for Room entities.
type RoomReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Room, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error)
	GetMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error)
	IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
}

// RoomWriter defines write operations for Room entities.
type RoomWriter interface {
	Create(ctx context.Context, room *domain.Room) error
	Update(ctx context.Context, room *domain.Room) error
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
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error)
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error)
}

// MessageWriter defines write operations for Message entities.
type MessageWriter interface {
	Create(ctx context.Context, message *domain.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MessageRepository combines read and write operations for Message entities.
type MessageRepository interface {
	MessageReader
	MessageWriter
}
