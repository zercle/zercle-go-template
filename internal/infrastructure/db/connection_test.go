//go:build integration

// Package db provides database connection management for the application.
// This file contains integration tests for database connectivity.
package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"zercle-go-template/internal/config"
	"zercle-go-template/internal/logger"
)

// getTestConfig returns a config for integration tests using environment variables.
func getTestConfig() *config.Config {
	port, _ := strconv.Atoi(getEnvOrDefault("DB_PORT", "5432"))

	return &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "test",
		},
		Database: config.DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     port,
			Database: getEnvOrDefault("DB_NAME", "zercle_template_test"),
			Username: getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		Log: config.LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestNew_Success tests successful database connection creation.
func TestNew_Success(t *testing.T) {
	cfg := getTestConfig()
	log := logger.NewNop()

	db, err := New(cfg, log)
	require.NoError(t, err, "should create database connection without error")
	require.NotNil(t, db, "should return non-nil database instance")

	defer func() {
		if db != nil {
			err := db.Close()
			assert.NoError(t, err, "should close database without error")
		}
	}()
}

// TestNew_ConnectionFailure tests database connection failure scenarios.
func TestNew_ConnectionFailure(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "empty host",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host: "",
					Port: 5432,
				},
			},
		},
		{
			name: "invalid port",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: -1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.NewNop()

			db, err := New(tt.config, log)

			// Should return nil or error for invalid configs
			if err != nil {
				assert.Nil(t, db)
			} else {
				assert.Nil(t, db)
			}
		})
	}
}

// TestPing verifies database ping functionality.
func TestPing(t *testing.T) {
	cfg := getTestConfig()
	log := logger.NewNop()

	db, err := New(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer db.Close()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "successful ping",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "ping with timeout",
			ctx:     testTimeoutContext(t, 5*time.Second),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Ping(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "ping should succeed")
			}
		})
	}
}

// TestPing_NoConnection tests ping when database is not connected.
func TestPing_NoConnection(t *testing.T) {
	db := &DB{}

	err := db.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestQuerier verifies that querier is accessible.
func TestQuerier(t *testing.T) {
	cfg := getTestConfig()
	log := logger.NewNop()

	db, err := New(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer db.Close()

	querier := db.Querier()
	assert.NotNil(t, querier, "querier should not be nil")
}

// TestStats verifies connection pool statistics.
func TestStats(t *testing.T) {
	cfg := getTestConfig()
	log := logger.NewNop()

	db, err := New(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer db.Close()

	stats := db.Stats()
	require.NotNil(t, stats, "stats should not be nil")

	// Verify stats have reasonable values
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(0), "total connections should be >= 0")
	assert.GreaterOrEqual(t, stats.AcquiredConns(), int32(0), "acquired connections should be >= 0")
}

// TestStats_NoConnection tests stats when database is not connected.
func TestStats_NoConnection(t *testing.T) {
	db := &DB{}

	stats := db.Stats()
	assert.Nil(t, stats, "stats should be nil when not connected")
}

// TestConnectionPool verifies connection pool behavior under load.
func TestConnectionPool(t *testing.T) {
	cfg := getTestConfig()
	log := logger.NewNop()

	db, err := New(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer db.Close()

	// Perform multiple concurrent pings to test connection pool
	ctx := context.Background()
	numGoroutines := 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			done <- db.Ping(ctx)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err, "concurrent ping should succeed")
	}

	// Verify connection stats after concurrent operations
	stats := db.Stats()
	require.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalConns(), int32(0))
}

// TestClose_Idempotent tests that Close can be called multiple times safely.
func TestClose_Idempotent(t *testing.T) {
	cfg := getTestConfig()
	log := logger.NewNop()

	db, err := New(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, db)

	// First close should succeed
	err = db.Close()
	assert.NoError(t, err)

	// Second close should also succeed (idempotent)
	err = db.Close()
	assert.NoError(t, err)
}

// TestDatabaseDSN verifies the DatabaseDSN function through integration test.
func TestDatabaseDSN(t *testing.T) {
	cfg := getTestConfig()

	dsn := cfg.DatabaseDSN()
	require.NotEmpty(t, dsn, "DSN should not be empty")

	// Verify DSN contains expected components
	assert.Contains(t, dsn, fmt.Sprintf("host=%s", cfg.Database.Host))
	assert.Contains(t, dsn, fmt.Sprintf("port=%d", cfg.Database.Port))
	assert.Contains(t, dsn, fmt.Sprintf("dbname=%s", cfg.Database.Database))
	assert.Contains(t, dsn, fmt.Sprintf("user=%s", cfg.Database.Username))
}

// testTimeoutContext creates a context with timeout for testing.
func testTimeoutContext(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}
