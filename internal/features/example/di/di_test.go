//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package di_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/config"
	"github.com/zercle/zercle-go-template/internal/features/example/di"
	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

// TestRegister_DepsMissing returns an error when required DI dependencies are
// not registered.
func TestRegister_DepsMissing(t *testing.T) {
	t.Parallel()

	injector := do.New()

	err := di.Register(injector)
	require.Error(t, err)
}

// TestRegister_WithStubs verifies the feature DI registers its providers when
// the required shared dependencies are present.
func TestRegister_WithStubs(t *testing.T) {
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
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
			BodyLimit:    "1M",
		},
		GRPC:   config.GRPCConfig{Host: "0.0.0.0", Port: 50051},
		OTel:   config.OTelConfig{Exporter: "none", ServiceName: "test"},
		Log:    config.LogConfig{Level: "info", Format: "json"},
		Valkey: config.ValkeyConfig{Host: "127.0.0.1", Port: 6379, DB: 0},
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)
	require.NoError(t, telemetry.Register(context.Background(), injector))

	err := di.Register(injector)
	require.Error(t, err, "expected error because infra providers are missing")
}
