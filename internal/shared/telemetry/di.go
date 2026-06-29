// DI registration for telemetry providers and the health registry.
package telemetry

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/zercle/zercle-go-template/internal/config"
)

// Register wires logger, tracer provider, meter provider, and health registry
// into the DI container. The per-provider shutdown callbacks are intentionally
// discarded; the Application resolves the providers directly and calls
// provider.Shutdown itself so lifecycle ordering is explicit.
func Register(ctx context.Context, c do.Injector) error {
	do.Provide(c, func(i do.Injector) (*zerolog.Logger, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return NewLogger(cfg)
	})

	do.Provide(c, func(i do.Injector) (*trace.TracerProvider, error) {
		cfg := do.MustInvoke[*config.Config](i)
		provider, _, err := NewTracerProvider(ctx, cfg)
		return provider, err
	})

	do.Provide(c, func(i do.Injector) (*metric.MeterProvider, error) {
		cfg := do.MustInvoke[*config.Config](i)
		provider, _, err := NewMeterProvider(cfg)
		return provider, err
	})

	do.Provide(c, func(_ do.Injector) (*Registry, error) {
		return NewRegistry(), nil
	})

	return nil
}
