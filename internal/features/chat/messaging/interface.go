package messaging

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

type PubSubServiceInterface interface {
	PublishMessage(ctx context.Context, roomID string, event MessageEvent) error
	PublishPresence(ctx context.Context, roomID string, event PresenceEvent) error
	PublishTyping(ctx context.Context, roomID, userID string, event TypingEvent) error
}
