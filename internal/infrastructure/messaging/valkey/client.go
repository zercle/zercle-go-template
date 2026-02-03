package valkey

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/shared/logger"
)

type Client struct {
	RDB *redis.Client
}

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

func (c *Client) Close() error {
	return c.RDB.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.RDB.Ping(ctx).Err()
}

func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.RDB.Publish(ctx, channel, message).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.RDB.Subscribe(ctx, channels...)
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.RDB.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.RDB.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.RDB.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.RDB.Exists(ctx, keys...).Result()
}

func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.RDB.HSet(ctx, key, values...).Err()
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.RDB.HGetAll(ctx, key).Result()
}
