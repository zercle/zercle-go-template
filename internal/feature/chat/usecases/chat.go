package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/zercle/zercle-go-template/internal/feature/chat/domain"
	"github.com/zercle/zercle-go-template/internal/feature/chat/ports"
)

// Service implements the chat business logic.
type Service struct {
	roomRepo    ports.RoomRepository
	messageRepo ports.MessageRepository
	pubsub      ports.PubSubService
}

var _ ports.Service = (*Service)(nil)

// NewService creates a new chat service without pub/sub.
func NewService(roomRepo ports.RoomRepository, messageRepo ports.MessageRepository) *Service {
	return &Service{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		pubsub:      nil,
	}
}

// NewServiceWithPubSub creates a new chat service with pub/sub.
func NewServiceWithPubSub(roomRepo ports.RoomRepository, messageRepo ports.MessageRepository, ps ports.PubSubService) *Service {
	return &Service{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		pubsub:      ps,
	}
}

// CreateRoom creates a new chat room.
func (s *Service) CreateRoom(ctx context.Context, input ports.CreateRoomInput) (*domain.Room, error) {
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

// GetRoom retrieves a room by ID.
func (s *Service) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, domain.ErrRoomNotFound
	}
	return room, nil
}

// ListRooms lists rooms for a user.
func (s *Service) ListRooms(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.roomRepo.FindByUserID(ctx, userID, limit, offset)
}

// UpdateRoom updates a room.
func (s *Service) UpdateRoom(ctx context.Context, roomID uuid.UUID, name, description string) (*domain.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, domain.ErrRoomNotFound
	}

	room.Name = name
	room.Description = description

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// DeleteRoom deletes a room.
func (s *Service) DeleteRoom(ctx context.Context, roomID uuid.UUID) error {
	return s.roomRepo.Delete(ctx, roomID)
}

// JoinRoom adds a user to a room.
func (s *Service) JoinRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	isMember, err := s.roomRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return err
	}

	if isMember {
		return domain.ErrAlreadyJoined
	}

	return s.roomRepo.AddMember(ctx, roomID, userID, "member")
}

// LeaveRoom removes a user from a room.
func (s *Service) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	isMember, err := s.roomRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return domain.ErrNotMember
	}

	return s.roomRepo.RemoveMember(ctx, roomID, userID)
}

// GetRoomMembers returns all members of a room.
func (s *Service) GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error) {
	return s.roomRepo.GetMembers(ctx, roomID)
}

// SendMessage sends a message to a room.
func (s *Service) SendMessage(ctx context.Context, input ports.SendMessageInput) (*domain.Message, error) {
	isMember, err := s.roomRepo.IsMember(ctx, input.RoomID, input.SenderID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, domain.ErrNotMember
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
		event := domain.MessageEvent{
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

// GetMessageHistory retrieves message history for a room.
func (s *Service) GetMessageHistory(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.FindByRoomID(ctx, roomID, limit, offset, before)
}
