package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

type ChatServiceInterface interface {
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
