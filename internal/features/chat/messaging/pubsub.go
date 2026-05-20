package messaging

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bytedance/sonic"

	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
)

// ChannelRoomMessages is the Redis channel pattern for room messages.
const (
	ChannelRoomMessages = "room:%s:messages"
	ChannelRoomPresence = "room:%s:presence"
	ChannelUserTyping   = "room:%s:typing:%s"
)

// MessageEvent represents a chat message event published to subscribers.
type MessageEvent struct {
	Type      string `json:"type"`
	RoomID    string `json:"room_id"`
	MessageID string `json:"message_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// PresenceEvent represents a user presence change event.
type PresenceEvent struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	Online bool   `json:"online"`
	SeenAt int64  `json:"seen_at"`
}

// TypingEvent represents a user typing indicator event.
type TypingEvent struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	RoomID    string `json:"room_id"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

// Service provides pub/sub functionality for chat events.
type Service struct {
	client *valkey.Client
}

var _ PubSubServiceInterface = (*Service)(nil)

// New creates a new Service with the given Valkey client.
func New(client *valkey.Client) *Service {
	return &Service{client: client}
}

// PublishMessage publishes a message event to the room channel.
func (s *Service) PublishMessage(ctx context.Context, roomID string, event MessageEvent) error {
	data, err := sonic.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal message event: %w", err)
	}

	channel := fmt.Sprintf(ChannelRoomMessages, roomID)
	if err := s.client.Publish(ctx, channel, data); err != nil {
		slog.Default().Error("failed to publish message", "error", err, "room", roomID)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	slog.Default().Debug("message published", "room", roomID, "message_id", event.MessageID)
	return nil
}

// PublishPresence publishes a presence event to the room channel.
func (s *Service) PublishPresence(ctx context.Context, roomID string, event PresenceEvent) error {
	data, err := sonic.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal presence event: %w", err)
	}

	channel := fmt.Sprintf(ChannelRoomPresence, roomID)
	if err := s.client.Publish(ctx, channel, data); err != nil {
		return fmt.Errorf("failed to publish presence event: %w", err)
	}
	return nil
}

// PublishTyping publishes a typing event to the user channel.
func (s *Service) PublishTyping(ctx context.Context, roomID, userID string, event TypingEvent) error {
	data, err := sonic.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal typing event: %w", err)
	}

	channel := fmt.Sprintf(ChannelUserTyping, roomID, userID)
	if err := s.client.Publish(ctx, channel, data); err != nil {
		return fmt.Errorf("failed to publish typing event: %w", err)
	}
	return nil
}

// SubscribeToRoom subscribes to message and presence events for a room.
func (s *Service) SubscribeToRoom(ctx context.Context, roomID string, handler func(eventType string, data []byte)) {
	msgChannel := fmt.Sprintf(ChannelRoomMessages, roomID)
	presenceChannel := fmt.Sprintf(ChannelRoomPresence, roomID)

	pubsub, err := s.client.Subscribe(ctx, msgChannel, presenceChannel)
	if err != nil {
		slog.Default().Error("failed to subscribe to room", "error", err)
		return
	}
	ch := pubsub.Channel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = pubsub.Close()
				return
			case msg := <-ch:
				var event MessageEvent
				if err := sonic.Unmarshal([]byte(msg.Payload), &event); err != nil {
					slog.Default().Error("failed to unmarshal event", "error", err)
					continue
				}
				handler(event.Type, []byte(msg.Payload))
			}
		}
	}()
}
