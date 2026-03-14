//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../mocks/service.mock.go -package=mocks

package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/feature/chat/domain"
)

// CreateRoomInput holds the data needed to create a room.
type CreateRoomInput struct {
	Name        string
	Description string
	Type        string
	OwnerID     uuid.UUID
	MemberIDs   []uuid.UUID
}

// SendMessageInput holds the data needed to send a message.
type SendMessageInput struct {
	RoomID      uuid.UUID
	SenderID    uuid.UUID
	Content     string
	MessageType string
	ReplyTo     *uuid.UUID
}

// Service defines the interface for chat operations.
type Service interface {
	CreateRoom(ctx context.Context, input CreateRoomInput) (*domain.Room, error)
	GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
	ListRooms(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error)
	UpdateRoom(ctx context.Context, roomID uuid.UUID, name, description string) (*domain.Room, error)
	DeleteRoom(ctx context.Context, roomID uuid.UUID) error
	JoinRoom(ctx context.Context, roomID, userID uuid.UUID) error
	LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error
	GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error)
	SendMessage(ctx context.Context, input SendMessageInput) (*domain.Message, error)
	GetMessageHistory(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error)
}
