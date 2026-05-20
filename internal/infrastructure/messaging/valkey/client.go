package valkey

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Client wraps a Valkey client for pub/sub and cache operations.
type Client struct {
	RDB valkey.Client
}

// New creates a new Client connected to the Valkey server.
func New(cfg ValkeyConfig) (*Client, error) {
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

	slog.Default().Info("Valkey connected", "host", cfg.Host, "port", cfg.Port)

	return &Client{RDB: rdb}, nil
}

// Close closes the Valkey client connection.
func (c *Client) Close() error {
	c.RDB.Close()
	return nil
}

// Ping checks if the Valkey connection is alive.
func (c *Client) Ping(ctx context.Context) error {
	if err := c.RDB.Do(ctx, c.RDB.B().Ping().Build()).Error(); err != nil {
		return fmt.Errorf("failed to ping valkey: %w", err)
	}
	return nil
}

// Publish publishes a message to a channel.
func (c *Client) Publish(ctx context.Context, channel string, message any) error {
	if err := c.RDB.Do(ctx, c.RDB.B().Publish().Channel(channel).Message(fmt.Sprint(message)).Build()).Error(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// Subscribe subscribes to one or more channels and returns a Subscription.
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
			slog.Default().Error("subscription closed with error", "error", fmt.Errorf("failed to subscribe to channel: %w", err))
		}
	}()

	return &Subscription{ch: msgCh, cancel: cancel}, nil
}

// Set sets a key-value pair with an expiration time.
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := c.RDB.Do(ctx, c.RDB.B().Set().Key(key).Value(fmt.Sprint(value)).Ex(expiration).Build()).Error(); err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}
	return nil
}

// Get retrieves the value of a key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	result, err := c.RDB.Do(ctx, c.RDB.B().Get().Key(key).Build()).ToString()
	if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}
	return result, nil
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if err := c.RDB.Do(ctx, c.RDB.B().Del().Key(keys...).Build()).Error(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// Exists checks if one or more keys exist.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	result, err := c.RDB.Do(ctx, c.RDB.B().Exists().Key(keys...).Build()).AsInt64()
	if err != nil {
		return 0, fmt.Errorf("failed to check key existence: %w", err)
	}
	return result, nil
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
	if err := c.RDB.Do(ctx, builder.Build()).Error(); err != nil {
		return fmt.Errorf("failed to set hash fields: %w", err)
	}
	return nil
}

// HGetAll retrieves all field-value pairs in a hash.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result, err := c.RDB.Do(ctx, c.RDB.B().Hgetall().Key(key).Build()).AsStrMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash fields: %w", err)
	}
	return result, nil
}

// Config holds the configuration for connecting to Valkey.
//
//go:generate stringer -type=ValkeyConfig -trimprefix=Valkey
//nolint:revive
type ValkeyConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns the Valkey server address in host:port format.
func (v ValkeyConfig) Addr() string {
	return fmt.Sprintf("%s:%d", v.Host, v.Port)
}
