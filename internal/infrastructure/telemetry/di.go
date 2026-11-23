// DI registration for telemetry providers and the health registry.
package telemetry

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

// Register wires logger, tracer provider, Prometheus registry, propagator,
// and the health registry into the DI container, and installs the tracer
// provider and propagator as OTel globals so third-party instrumentation
// (otelhttp) that reads the globals links into the same trace. Owned
// boundaries receive the providers explicitly instead of resolving them
// globally.
//
// The per-provider shutdown callbacks are intentionally discarded; the
// Application resolves the providers directly and calls provider.Shutdown
// itself so lifecycle ordering is explicit.
func Register(ctx context.Context, c do.Injector) error {
	do.Provide(c, func(i do.Injector) (*zerolog.Logger, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return NewLogger(cfg)
	})

	do.Provide(c, func(i do.Injector) (propagation.TextMapPropagator, error) {
		propagator := NewPropagator()
		otel.SetTextMapPropagator(propagator)
		return propagator, nil
	})

	do.Provide(c, func(i do.Injector) (*trace.TracerProvider, error) {
		cfg := do.MustInvoke[*config.Config](i)
		provider, _, err := NewTracerProvider(ctx, cfg)
		if err != nil {
			return nil, err
		}
		otel.SetTracerProvider(provider)
		return provider, nil
	})

	do.Provide(c, func(_ do.Injector) (*prometheus.Registry, error) {
		return NewPrometheusRegistry(), nil
	})

	do.Provide(c, func(_ do.Injector) (*Registry, error) {
		return NewRegistry(), nil
	})

	return nil
}
