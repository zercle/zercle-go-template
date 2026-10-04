//go:build unit

package server

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
	"github.com/zercle/zercle-go-template/internal/infrastructure/telemetry"
)

// TestApplication_ShutdownLeavesProvidersToInjector pins that Application.shutdown
// does not shut down the OTel providers itself: samber/do owns them and calls
// their Shutdown when the injector is shut down. Shutting them down twice makes
// the Prometheus reader fail with "reader is shutdown", which surfaced as a
// spurious "injector shutdown error" on every clean exit.
func TestApplication_ShutdownLeavesProvidersToInjector(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		App:  config.AppConfig{Environment: "test", ShutdownTimeout: 2 * time.Second},
		OTel: config.OTelConfig{Exporter: "none", ServiceName: "test", Sampling: 1.0},
		Log:  config.LogConfig{Level: "info", Format: "json"},
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)
	require.NoError(t, telemetry.Register(context.Background(), injector))

	logger, err := do.Invoke[*zerolog.Logger](injector)
	require.NoError(t, err)

	app := NewApplication(injector, cfg, logger)

	ctx := context.Background()
	app.shutdown(ctx)

	report := injector.Shutdown()
	require.True(t, report == nil || report.Succeed,
		"injector shutdown must be clean; a failure here means a provider was shut down twice: %v", report)
}
