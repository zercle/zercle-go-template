//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../mocks/repository.mock.go -package=mocks

package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/feature/chat/domain"
)

// RoomReader defines methods for reading room data.
type RoomReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Room, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error)
	GetMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error)
	IsMember(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
}

// RoomWriter defines methods for writing room data.
type RoomWriter interface {
	Create(ctx context.Context, room *domain.Room) error
	Update(ctx context.Context, room *domain.Room) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RoomMembershipManager defines methods for managing room membership.
type RoomMembershipManager interface {
	AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, roomID, userID uuid.UUID) error
}

// RoomRepository combines room reading, writing, and membership operations.
type RoomRepository interface {
	RoomReader
	RoomWriter
	RoomMembershipManager
}

// MessageReader defines methods for reading message data.
type MessageReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error)
	FindByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error)
}

// MessageWriter defines methods for writing message data.
type MessageWriter interface {
	Create(ctx context.Context, message *domain.Message) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MessageRepository combines message reading and writing operations.
type MessageRepository interface {
	MessageReader
	MessageWriter
}
