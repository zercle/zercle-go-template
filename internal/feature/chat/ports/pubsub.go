//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../mocks/pubsub.mock.go -package=mocks

package ports

import (
	"context"

	"github.com/zercle/zercle-go-template/internal/feature/chat/domain"
)

// PubSubService defines the interface for pub/sub operations.
type PubSubService interface {
	PublishMessage(ctx context.Context, roomID string, event domain.MessageEvent) error
	PublishPresence(ctx context.Context, roomID string, event domain.PresenceEvent) error
	PublishTyping(ctx context.Context, roomID, userID string, event domain.TypingEvent) error
}
