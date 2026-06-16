//go:build unit

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
)

// TestNewPool_InvalidDSN verifies that NewPool returns a clear error without
// blocking when given an invalid DSN.
func TestNewPool_InvalidDSN(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		DB: config.DBConfig{
			Host:           "localhost",
			Port:           5432,
			Name:           "app",
			User:           "postgres",
			Password:       "postgres",
			SSLMode:        "invalid_ssl_mode",
			MaxConns:       10,
			MinConns:       2,
			MaxConnIdle:    30 * time.Second,
			MaxConnLife:    1 * time.Minute,
			ConnectTimeout: 1 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	require.Error(t, err, "expected an error for an invalid DSN")
	assert.Nil(t, pool, "pool must be nil when creation fails")
	assert.Contains(t, err.Error(), "parse db config", "error should describe config parsing failure")
}

// TestNewPool_UnreachableHost verifies that NewPool fails promptly when the
// configured host is unreachable, and that the pool is not leaked.
func TestNewPool_UnreachableHost(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		DB: config.DBConfig{
			Host:           "192.0.2.1", // TEST-NET-1, should not be reachable
			Port:           5432,
			Name:           "app",
			User:           "postgres",
			Password:       "postgres",
			SSLMode:        "disable",
			MaxConns:       2,
			MinConns:       1,
			MaxConnIdle:    5 * time.Second,
			MaxConnLife:    10 * time.Second,
			ConnectTimeout: 1 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	require.Error(t, err, "expected an error when pinging an unreachable host")
	assert.Nil(t, pool, "pool must be nil when ping fails")
	assert.Contains(t, err.Error(), "ping db", "error should describe ping failure")
}
