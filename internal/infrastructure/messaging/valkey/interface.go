package valkey

import (
	"context"
	"time"
)

// Message represents a pub/sub message received from a channel.
type Message struct {
	Channel string
	Payload string
}

// Subscription represents a channel subscription that can receive messages.
type Subscription struct {
	ch     chan Message
	cancel context.CancelFunc
}

// Channel returns a read-only channel of messages received on this subscription.
func (s *Subscription) Channel() <-chan Message {
	return s.ch
}

// Close cancels the subscription and releases resources.
func (s *Subscription) Close() error {
	s.cancel()
	return nil
}

// PubSubClient defines the interface for Valkey pub/sub and cache operations.
type PubSubClient interface {
	Publish(ctx context.Context, channel string, message any) error
	Subscribe(ctx context.Context, channels ...string) (*Subscription, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	HSet(ctx context.Context, key string, values ...any) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	Ping(ctx context.Context) error
	Close() error
}