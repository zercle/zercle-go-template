//go:build unit

package valkey_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/messaging/valkey"
)

func TestConfig_ValkeyAddr(t *testing.T) {
	cfg := &config.Config{Valkey: config.ValkeyConfig{Host: "127.0.0.1", Port: 6379}}
	require.Equal(t, "127.0.0.1:6379", cfg.ValkeyAddr())
}

func TestNewClient_Unreachable(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Valkey: config.ValkeyConfig{
			Host: "192.0.2.1", // TEST-NET-1, should not be reachable
			Port: 6379,
			DB:   0,
		},
		Log: config.LogConfig{Level: "info", Format: "json"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := valkey.NewClient(ctx, cfg)
	assert.Error(t, err, "expected unreachable valkey to error")
	assert.Nil(t, client, "client must be nil when creation fails")
	assert.Contains(t, err.Error(), "valkey", "error should mention valkey")
}
