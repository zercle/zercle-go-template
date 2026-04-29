package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

// ChannelRoomMessages is the template for room message channels.
// ChannelRoomMessages is the template for room message channels.
const (
	ChannelRoomMessages = "room:%s:messages"
	ChannelRoomPresence = "room:%s:presence"
	ChannelUserTyping   = "room:%s:typing:%s"
)

// MessageEvent represents a chat message published to a room.
// MessageEvent represents a chat message published to a room.
type MessageEvent struct {
	Type      string `json:"type"`
	RoomID    string `json:"room_id"`
	MessageID string `json:"message_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// PresenceEvent represents a user presence update in a room.
// PresenceEvent represents a user presence update in a room.
type PresenceEvent struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	Online bool   `json:"online"`
	SeenAt int64  `json:"seen_at"`
}

// TypingEvent represents a user typing indicator in a room.
// TypingEvent represents a user typing indicator in a room.
type TypingEvent struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	RoomID    string `json:"room_id"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

// Service implements PubSubServiceInterface using Valkey.
// Service implements PubSubServiceInterface using Valkey.
type Service struct {
	client *valkey.Client
}

var _ PubSubServiceInterface = (*Service)(nil)

// New creates a new PubSub Service with the given Valkey client.
// New creates a new PubSub Service with the given Valkey client.
func New(client *valkey.Client) *Service {
	return &Service{client: client}
}

// PublishMessage publishes a message event to a room channel.
func (s *Service) PublishMessage(ctx context.Context, roomID string, event MessageEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal message event: %w", err)
	}

	channel := fmt.Sprintf(ChannelRoomMessages, roomID)
	if err := s.client.Publish(ctx, channel, data); err != nil {
		logger.Error().Err(err).Str("room", roomID).Msg("Failed to publish message")
		return err
	}

	logger.Debug().Str("room", roomID).Str("message_id", event.MessageID).Msg("Message published")
	return nil
}

// PublishPresence publishes a presence event to a room channel.
func (s *Service) PublishPresence(ctx context.Context, roomID string, event PresenceEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal presence event: %w", err)
	}

	channel := fmt.Sprintf(ChannelRoomPresence, roomID)
	return s.client.Publish(ctx, channel, data)
}

// PublishTyping publishes a typing event to a user's typing channel.
func (s *Service) PublishTyping(ctx context.Context, roomID, userID string, event TypingEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal typing event: %w", err)
	}

	channel := fmt.Sprintf(ChannelUserTyping, roomID, userID)
	return s.client.Publish(ctx, channel, data)
}

// SubscribeToRoom subscribes to message and presence channels for a room.
// SubscribeToRoom subscribes to message and presence channels for a room.
func (s *Service) SubscribeToRoom(ctx context.Context, roomID string, handler func(eventType string, data []byte)) {
	msgChannel := fmt.Sprintf(ChannelRoomMessages, roomID)
	presenceChannel := fmt.Sprintf(ChannelRoomPresence, roomID)

	pubsub := s.client.Subscribe(ctx, msgChannel, presenceChannel)
	ch := pubsub.Channel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = pubsub.Close()
				return
			case msg := <-ch:
				var event MessageEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					logger.Error().Err(err).Msg("Failed to unmarshal event")
					continue
				}
				handler(event.Type, []byte(msg.Payload))
			}
		}
	}()
}
