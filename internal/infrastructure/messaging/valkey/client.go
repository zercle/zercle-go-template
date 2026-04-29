package valkey

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

// Client wraps a Valkey (Redis) client for messaging operations.
type Client struct {
	RDB *redis.Client
}

// New creates a new Valkey client connection.
func New(cfg config.ValkeyConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	logger.Info().Str("host", cfg.Host).Int("port", cfg.Port).Msg("Valkey connected")

	return &Client{RDB: rdb}, nil
}

// Close closes the Valkey client connection.
func (c *Client) Close() error {
	return c.RDB.Close()
}

// Ping checks the Valkey connection is alive.
func (c *Client) Ping(ctx context.Context) error {
	return c.RDB.Ping(ctx).Err()
}

// Publish publishes a message to a channel.
func (c *Client) Publish(ctx context.Context, channel string, message any) error {
	return c.RDB.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to one or more channels.
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.RDB.Subscribe(ctx, channels...)
}

// Set stores a key-value pair with expiration.
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.RDB.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.RDB.Get(ctx, key).Result()
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.RDB.Del(ctx, keys...).Err()
}

// Exists checks if one or more keys exist.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.RDB.Exists(ctx, keys...).Result()
}

// HSet sets one or more field-value pairs in a hash.
func (c *Client) HSet(ctx context.Context, key string, values ...any) error {
	return c.RDB.HSet(ctx, key, values...).Err()
}

// HGetAll retrieves all field-value pairs in a hash.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.RDB.HGetAll(ctx, key).Result()
}
