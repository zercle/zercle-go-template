package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/zercle/zercle-go-template/internal/errors"
	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
	"github.com/zercle/zercle-go-template/internal/features/chat/messaging"
)

type ChatService struct {
	roomRepo    domain.RoomRepository
	messageRepo domain.MessageRepository
	pubsub      messaging.PubSubServiceInterface
}

var _ ChatServiceInterface = (*ChatService)(nil)

func NewChatService(roomRepo domain.RoomRepository, messageRepo domain.MessageRepository) *ChatService {
	return &ChatService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		pubsub:      nil,
	}
}

func NewChatServiceWithPubSub(roomRepo domain.RoomRepository, messageRepo domain.MessageRepository, ps messaging.PubSubServiceInterface) *ChatService {
	return &ChatService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		pubsub:      ps,
	}
}

type CreateRoomInput struct {
	Name        string
	Description string
	Type        string
	OwnerID     uuid.UUID
	MemberIDs   []uuid.UUID
}

func (s *ChatService) CreateRoom(ctx context.Context, input CreateRoomInput) (*domain.Room, error) {
	room := domain.NewRoom(input.Name, input.Description, input.Type, input.OwnerID)
	if err := room.Validate(); err != nil {
		return nil, err
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	if err := s.roomRepo.AddMember(ctx, room.ID, input.OwnerID, "owner"); err != nil {
		return nil, err
	}

	for _, memberID := range input.MemberIDs {
		if memberID != input.OwnerID {
			if err := s.roomRepo.AddMember(ctx, room.ID, memberID, "member"); err != nil {
				return nil, err
			}
		}
	}

	return room, nil
}

func (s *ChatService) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, apperrors.ErrRoomNotFound
	}
	return room, nil
}

func (s *ChatService) ListRooms(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.roomRepo.FindByUserID(ctx, userID, limit, offset)
}

func (s *ChatService) UpdateRoom(ctx context.Context, roomID uuid.UUID, name, description string) (*domain.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, apperrors.ErrRoomNotFound
	}

	room.Name = name
	room.Description = description

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *ChatService) DeleteRoom(ctx context.Context, roomID uuid.UUID) error {
	return s.roomRepo.Delete(ctx, roomID)
}

func (s *ChatService) JoinRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	isMember, err := s.roomRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return err
	}

	if isMember {
		return apperrors.ErrAlreadyJoined
	}

	return s.roomRepo.AddMember(ctx, roomID, userID, "member")
}

func (s *ChatService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	isMember, err := s.roomRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return apperrors.ErrNotMember
	}

	return s.roomRepo.RemoveMember(ctx, roomID, userID)
}

func (s *ChatService) GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error) {
	return s.roomRepo.GetMembers(ctx, roomID)
}

type SendMessageInput struct {
	RoomID      uuid.UUID
	SenderID    uuid.UUID
	Content     string
	MessageType string
	ReplyTo     *uuid.UUID
}

func (s *ChatService) SendMessage(ctx context.Context, input SendMessageInput) (*domain.Message, error) {
	isMember, err := s.roomRepo.IsMember(ctx, input.RoomID, input.SenderID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, apperrors.ErrNotMember
	}

	if input.MessageType == "" {
		input.MessageType = "text"
	}

	message := domain.NewMessage(input.RoomID, input.SenderID, input.Content, input.MessageType)
	message.ReplyTo = input.ReplyTo

	if err := message.Validate(); err != nil {
		return nil, err
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	if s.pubsub != nil {
		event := messaging.MessageEvent{
			Type:      "message",
			RoomID:    message.RoomID.String(),
			MessageID: message.ID.String(),
			SenderID:  message.SenderID.String(),
			Content:   message.Content,
			Timestamp: message.CreatedAt.Unix(),
		}
		_ = s.pubsub.PublishMessage(ctx, message.RoomID.String(), event)
	}

	return message, nil
}

func (s *ChatService) GetMessageHistory(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.FindByRoomID(ctx, roomID, limit, offset, before)
}
