//go:build unit

package db_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/db"
)

// TestNewDB_NilConfig verifies that NewDB rejects a nil config without
// attempting to parse a DSN or open a connection.
func TestNewDB_NilConfig(t *testing.T) {
	t.Parallel()

	nop := zerolog.Nop()
	gormDB, err := db.NewDB(context.Background(), nil, &nop)
	require.Error(t, err, "expected an error for a nil config")
	assert.Nil(t, gormDB, "gorm.DB must be nil when config is nil")
	assert.Contains(t, err.Error(), "config is nil", "error should mention nil config")
}

// TestNewDB_NilLogger verifies that NewDB rejects a nil logger.
func TestNewDB_NilLogger(t *testing.T) {
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
			MaxIdleConns:   1,
			MaxConnIdle:    5 * time.Second,
			MaxConnLife:    10 * time.Second,
			ConnectTimeout: 1 * time.Second,
		},
	}

	gormDB, err := db.NewDB(context.Background(), cfg, nil)
	require.Error(t, err, "expected an error for a nil logger")
	assert.Nil(t, gormDB, "gorm.DB must be nil when logger is nil")
	assert.Contains(t, err.Error(), "logger is nil", "error should mention nil logger")
}

// TestNewDB_UnreachableHost verifies that NewDB fails when the configured host
// is unreachable and that no usable *gorm.DB is returned.
//
// Note: the underlying pgx stdlib driver used by gorm.io/driver/postgres may
// fail during gorm.Open (eager dial) or during the explicit PingContext call,
// depending on driver version and timeout interaction. Either failure path is
// acceptable — what matters is that the caller sees a wrapped error and a nil
// *gorm.DB so the connection pool is not leaked.
func TestNewDB_UnreachableHost(t *testing.T) {
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
			MaxIdleConns:   1,
			MaxConnIdle:    5 * time.Second,
			MaxConnLife:    10 * time.Second,
			ConnectTimeout: 1 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nop := zerolog.Nop()
	gormDB, err := db.NewDB(ctx, cfg, &nop)
	require.Error(t, err, "expected an error when the host is unreachable")
	assert.Nil(t, gormDB, "gorm.DB must be nil on connection failure")
	msg := err.Error()
	if !assert.True(t,
		strings.Contains(msg, "ping db") || strings.Contains(msg, "open gorm"),
		"error should describe either ping or open failure; got: %s", msg,
	) {
		return
	}
	// Make absolutely sure we never accidentally expose a non-nil *gorm.DB on
	// failure paths — the production contract is nil-or-valid.
	if gormDB != nil {
		assert.Fail(t, "gorm.DB should be nil on failure")
	}
}
