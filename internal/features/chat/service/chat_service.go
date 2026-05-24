package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/zercle/zercle-go-template/internal/features/chat/domain"
	"github.com/zercle/zercle-go-template/internal/features/chat/messaging"
	apperrors "github.com/zercle/zercle-go-template/internal/shared/errors"
)

const (
	// MessageTypeText is the standard text message type.
	MessageTypeText = "text"
	// DefaultRoomPageSize is the default number of rooms to return per page.
	DefaultRoomPageSize = 20
	// MaxRoomPageSize defines the maximum number of rooms returned per page.
	MaxRoomPageSize = 100
	// DefaultMessagePageSize is the default number of messages to return per page.
	DefaultMessagePageSize = 50
	// MaxMessagePageSize defines the maximum number of messages returned per page.
	MaxMessagePageSize = 100
)

// ChatService implements ChatServiceInterface with room and message operations.
type ChatService struct {
	roomRepo    domain.RoomRepository
	messageRepo domain.MessageRepository
	pubsub      messaging.PubSubServiceInterface
	logger      *zerolog.Logger
}

var _ ChatServiceInterface = (*ChatService)(nil)

// NewChatService creates a ChatService without pub/sub support.
func NewChatService(roomRepo domain.RoomRepository, messageRepo domain.MessageRepository, logger *zerolog.Logger) *ChatService {
	if logger == nil {
		l := zerolog.Nop()
		logger = &l
	}
	return &ChatService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		pubsub:      nil,
		logger:      logger,
	}
}

// NewChatServiceWithPubSub creates a ChatService with pub/sub support for real-time events.
func NewChatServiceWithPubSub(roomRepo domain.RoomRepository, messageRepo domain.MessageRepository, ps messaging.PubSubServiceInterface, logger *zerolog.Logger) *ChatService {
	if logger == nil {
		l := zerolog.Nop()
		logger = &l
	}
	return &ChatService{
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		pubsub:      ps,
		logger:      logger,
	}
}

// CreateRoomInput contains the required fields to create a new chat room.
type CreateRoomInput struct {
	Name        string
	Description string
	Type        string
	OwnerID     uuid.UUID
	MemberIDs   []uuid.UUID
}

// CreateRoom creates a new chat room with the owner as first member.
func (s *ChatService) CreateRoom(ctx context.Context, input CreateRoomInput) (*domain.Room, error) {
	room := domain.NewRoom(input.Name, input.Description, input.Type, input.OwnerID)
	if err := room.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate room: %w", err)
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	if err := s.roomRepo.AddMember(ctx, room.ID, input.OwnerID, "owner"); err != nil {
		return nil, fmt.Errorf("failed to add room owner: %w", err)
	}

	for _, memberID := range input.MemberIDs {
		if memberID != input.OwnerID {
			if err := s.roomRepo.AddMember(ctx, room.ID, memberID, "member"); err != nil {
				return nil, fmt.Errorf("failed to add room member: %w", err)
			}
		}
	}

	return room, nil
}

// GetRoom retrieves a room by ID or returns ErrRoomNotFound.
func (s *ChatService) GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, apperrors.ErrRoomNotFound
	}
	return room, nil
}

// ListRooms retrieves paginated rooms for a user with enforced limits.
func (s *ChatService) ListRooms(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Room, int, error) {
	if limit <= 0 {
		limit = DefaultRoomPageSize
	}
	if limit > MaxRoomPageSize {
		limit = MaxRoomPageSize
	}

	rooms, total, err := s.roomRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find rooms by user ID: %w", err)
	}
	return rooms, total, nil
}

// UpdateRoom updates room name and description.
func (s *ChatService) UpdateRoom(ctx context.Context, roomID uuid.UUID, name, description string) (*domain.Room, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, apperrors.ErrRoomNotFound
	}

	room.Name = name
	room.Description = description

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	return room, nil
}

// DeleteRoom deletes a room by ID.
func (s *ChatService) DeleteRoom(ctx context.Context, roomID uuid.UUID) error {
	if err := s.roomRepo.Delete(ctx, roomID); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}
	return nil
}

// JoinRoom adds a user to a room; returns ErrAlreadyJoined if already a member.
func (s *ChatService) JoinRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	isMember, err := s.roomRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}

	if isMember {
		return apperrors.ErrAlreadyJoined
	}

	if err := s.roomRepo.AddMember(ctx, roomID, userID, "member"); err != nil {
		return fmt.Errorf("failed to add member to room: %w", err)
	}
	return nil
}

// LeaveRoom removes a user from a room; returns ErrNotMember if not a member.
func (s *ChatService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	isMember, err := s.roomRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to check room membership: %w", err)
	}

	if !isMember {
		return apperrors.ErrNotMember
	}

	if err := s.roomRepo.RemoveMember(ctx, roomID, userID); err != nil {
		return fmt.Errorf("failed to remove member from room: %w", err)
	}
	return nil
}

// GetRoomMembers retrieves all members of a room.
func (s *ChatService) GetRoomMembers(ctx context.Context, roomID uuid.UUID) ([]*domain.RoomMember, error) {
	members, err := s.roomRepo.GetMembers(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}
	return members, nil
}

// SendMessageInput contains the required fields to send a message.
type SendMessageInput struct {
	RoomID      uuid.UUID
	SenderID    uuid.UUID
	Content     string
	MessageType string
	ReplyTo     *uuid.UUID
}

// SendMessage sends a message to a room and publishes to pub/sub if configured.
func (s *ChatService) SendMessage(ctx context.Context, input SendMessageInput) (*domain.Message, error) {
	isMember, err := s.roomRepo.IsMember(ctx, input.RoomID, input.SenderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room membership: %w", err)
	}

	if !isMember {
		return nil, apperrors.ErrNotMember
	}

	if input.MessageType == "" {
		input.MessageType = MessageTypeText
	}

	message := domain.NewMessage(input.RoomID, input.SenderID, input.Content, input.MessageType)
	message.ReplyTo = input.ReplyTo

	if err := message.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate message: %w", err)
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
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
		if err := s.pubsub.PublishMessage(ctx, message.RoomID.String(), event); err != nil {
			s.logger.Error().
				Err(err).
				Str("room_id", message.RoomID.String()).
				Msg("failed to publish message event")
		}
	}

	return message, nil
}

// GetMessageHistory retrieves paginated message history for a room.
func (s *ChatService) GetMessageHistory(ctx context.Context, roomID uuid.UUID, limit, offset int, before *uuid.UUID) ([]*domain.Message, bool, error) {
	if limit <= 0 {
		limit = DefaultMessagePageSize
	}
	if limit > MaxMessagePageSize {
		limit = MaxMessagePageSize
	}

	messages, hasMore, err := s.messageRepo.FindByRoomID(ctx, roomID, limit, offset, before)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find messages by room ID: %w", err)
	}
	return messages, hasMore, nil
}
