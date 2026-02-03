package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

const (
	ChannelRoomMessages = "room:%s:messages"
	ChannelRoomPresence = "room:%s:presence"
	ChannelUserTyping   = "room:%s:typing:%s"
)

type MessageEvent struct {
	Type      string `json:"type"`
	RoomID    string `json:"room_id"`
	MessageID string `json:"message_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type PresenceEvent struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	Online bool   `json:"online"`
	SeenAt int64  `json:"seen_at"`
}

type TypingEvent struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	RoomID    string `json:"room_id"`
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
}

type Service struct {
	client *valkey.Client
}

var _ PubSubServiceInterface = (*Service)(nil)

func New(client *valkey.Client) *Service {
	return &Service{client: client}
}

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

func (s *Service) PublishPresence(ctx context.Context, roomID string, event PresenceEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal presence event: %w", err)
	}

	channel := fmt.Sprintf(ChannelRoomPresence, roomID)
	return s.client.Publish(ctx, channel, data)
}

func (s *Service) PublishTyping(ctx context.Context, roomID, userID string, event TypingEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal typing event: %w", err)
	}

	channel := fmt.Sprintf(ChannelUserTyping, roomID, userID)
	return s.client.Publish(ctx, channel, data)
}

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
