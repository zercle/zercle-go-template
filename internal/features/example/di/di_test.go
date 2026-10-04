//go:build unit

// STUB FEATURE — delete internal/features/example to start your project.

package di_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/features/example/di"
	"github.com/zercle/zercle-go-template/internal/features/example/usecase"
	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/telemetry"
)

// TestRegister_DepsMissing returns an error when required DI dependencies are
// not registered while the feature is enabled.
func TestRegister_DepsMissing(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Example: config.ExampleConfig{Enabled: true}})

	err := di.Register(injector)
	require.Error(t, err)
}

// TestRegister_DisabledSkipsFeature pins the feature-flag contract: with
// Enabled=false the feature registers nothing, resolves nothing, and mounts no
// routes, so a disabled feature cannot fail to wire or expose endpoints.
func TestRegister_DisabledSkipsFeature(t *testing.T) {
	t.Parallel()

	injector := do.New()
	do.ProvideValue(injector, &config.Config{Example: config.ExampleConfig{Enabled: false}})

	require.NoError(t, di.Register(injector))

	_, err := do.Invoke[usecase.Service](injector)
	require.Error(t, err, "disabled feature must not provide its service")
}

// TestRegister_WithStubs verifies the feature DI registers its providers when
// the required shared dependencies are present.
func TestRegister_WithStubs(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		App: config.AppConfig{
			Environment:     "test",
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
		OTel:   config.OTelConfig{Exporter: "none", ServiceName: "test"},
		Log:    config.LogConfig{Level: "info", Format: "json"},
		Valkey: config.ValkeyConfig{Host: "127.0.0.1", Port: 6379, DB: 0},
		Example: config.ExampleConfig{
			Enabled:         true,
			DefaultPageSize: 20,
			MaxPageSize:     100,
			MaxNameLength:   255,
		},
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)
	require.NoError(t, telemetry.Register(context.Background(), injector))

	err := di.Register(injector)
	require.Error(t, err, "expected error because infra providers are missing")
}
