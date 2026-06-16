//go:build unit

package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/app"
	"github.com/zercle/zercle-go-template/internal/config"
)

// TestBuild_DatabaseUnreachable verifies that Build returns an error and shuts
// down cleanly when the configured database is unreachable, without panicking.
func TestBuild_DatabaseUnreachable(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		App: config.AppConfig{
			Name:            "zercle-go-template",
			Environment:     "test",
			Host:            "0.0.0.0",
			Port:            8080,
			ShutdownTimeout: 5 * time.Second,
		},
		HTTP: config.HTTPConfig{
			Host:               "0.0.0.0",
			Port:               8080,
			ReadTimeout:        15 * time.Second,
			WriteTimeout:       15 * time.Second,
			IdleTimeout:        60 * time.Second,
			BodyLimit:          "1M",
			HealthProbeTimeout: 5 * time.Second,
		},
		GRPC: config.GRPCConfig{Host: "0.0.0.0", Port: 50051},
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
		Valkey: config.ValkeyConfig{
			Host: "127.0.0.1",
			Port: 6379,
			DB:   0,
		},
		OTel: config.OTelConfig{Exporter: "none", ServiceName: "test"},
		Log:  config.LogConfig{Level: "info", Format: "json"},
		Example: config.ExampleConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
			MaxNameLength:   255,
		},
	}

	require.NoError(t, cfg.Validate())

	application, injector, err := app.Build(context.Background(), cfg)
	assert.Error(t, err, "expected database unreachable error")
	assert.Nil(t, application, "application must be nil when Build fails")
	assert.NotNil(t, injector, "injector must be returned for shutdown")

	assert.NotNil(t, injector, "injector must be returned for shutdown")

	report := injector.Shutdown()
	assert.True(t, report == nil || report.Succeed, "expected clean injector shutdown")
}
