package valkey

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

// Client wraps a Valkey (Redis) client for messaging operations.
type Client struct {
	RDB valkey.Client
}

// New creates a new Valkey client connection.
func New(cfg config.ValkeyConfig) (*Client, error) {
	rdb, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.Addr()},
		Password:    cfg.Password,
		SelectDB:    cfg.DB,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}
	if err := rdb.Do(context.Background(), rdb.B().Ping().Build()).Error(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	logger.Info().Str("host", cfg.Host).Int("port", cfg.Port).Msg("Valkey connected")

	return &Client{RDB: rdb}, nil
}

// Close closes the Valkey client connection.
func (c *Client) Close() error {
	c.RDB.Close()
	return nil
}

// Ping checks the Valkey connection is alive.
func (c *Client) Ping(ctx context.Context) error {
	return c.RDB.Do(ctx, c.RDB.B().Ping().Build()).Error()
}

// Publish publishes a message to a channel.
func (c *Client) Publish(ctx context.Context, channel string, message any) error {
	return c.RDB.Do(ctx, c.RDB.B().Publish().Channel(channel).Message(fmt.Sprint(message)).Build()).Error()
}

// Subscribe subscribes to one or more channels.
func (c *Client) Subscribe(ctx context.Context, channels ...string) (*Subscription, error) {
	msgCh := make(chan Message, 64)
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(msgCh)
		defer cancel()
		cmd := c.RDB.B().Subscribe().Channel(channels...).Build()
		err := c.RDB.Receive(ctx, cmd, func(msg valkey.PubSubMessage) {
			select {
			case msgCh <- Message{Channel: msg.Channel, Payload: msg.Message}:
			case <-ctx.Done():
			}
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error().Err(err).Msg("Subscription closed with error")
		}
	}()

	return &Subscription{ch: msgCh, cancel: cancel}, nil
}

// Set stores a key-value pair with expiration.
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.RDB.Do(ctx, c.RDB.B().Set().Key(key).Value(fmt.Sprint(value)).Ex(expiration).Build()).Error()
}

// Get retrieves a value by key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.RDB.Do(ctx, c.RDB.B().Get().Key(key).Build()).ToString()
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.RDB.Do(ctx, c.RDB.B().Del().Key(keys...).Build()).Error()
}

// Exists checks if one or more keys exist.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.RDB.Do(ctx, c.RDB.B().Exists().Key(keys...).Build()).AsInt64()
}

// HSet sets one or more field-value pairs in a hash.
func (c *Client) HSet(ctx context.Context, key string, values ...any) error {
	builder := c.RDB.B().Hset().Key(key).FieldValue()
	for i := 0; i < len(values); i += 2 {
		if i+1 < len(values) {
			field := fmt.Sprint(values[i])
			value := fmt.Sprint(values[i+1])
			builder = builder.FieldValue(field, value)
		}
	}
	return c.RDB.Do(ctx, builder.Build()).Error()
}

// HGetAll retrieves all field-value pairs in a hash.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.RDB.Do(ctx, c.RDB.B().Hgetall().Key(key).Build()).AsStrMap()
}