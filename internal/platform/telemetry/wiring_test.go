//go:build unit

package telemetry_test

import (
	"context"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/zercle/zercle-go-template/internal/platform/config"
	"github.com/zercle/zercle-go-template/internal/platform/telemetry"
)

// TestRegister_InstallsOTelGlobals verifies that resolving the tracer provider
// and propagator from the injector installs them as OTel globals, so
// third-party instrumentation that reads the globals (otelhttp)
// participates in the same trace instead of recording into noop providers.
func TestRegister_InstallsOTelGlobals(t *testing.T) {
	cfg := &config.Config{
		OTel: config.OTelConfig{Exporter: "none", ServiceName: "globals-test", Sampling: 1.0},
		Log:  config.LogConfig{Level: "info", Format: "json"},
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)
	require.NoError(t, telemetry.Register(context.Background(), injector))

	tp, err := do.Invoke[*trace.TracerProvider](injector)
	require.NoError(t, err)
	propagator, err := do.Invoke[propagation.TextMapPropagator](injector)
	require.NoError(t, err)
	mp, err := do.Invoke[*metric.MeterProvider](injector)
	require.NoError(t, err)

	require.Same(t, tp, otel.GetTracerProvider(), "tracer provider must be installed globally")
	require.Equal(t, propagator, otel.GetTextMapPropagator(), "propagator must be installed globally")
	require.Same(t, mp, otel.GetMeterProvider(), "meter provider must be installed globally")
}

// TestNewPropagator_ExtractsW3CTraceContext verifies the propagator extracts a
// W3C traceparent header, which is what links an inbound request to its
// upstream trace.
func TestNewPropagator_ExtractsW3CTraceContext(t *testing.T) {
	propagator := telemetry.NewPropagator()

	header := propagation.HeaderCarrier{}
	header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	ctx := propagator.Extract(context.Background(), header)
	sc := oteltrace.SpanContextFromContext(ctx)

	require.True(t, sc.IsValid(), "expected a valid extracted span context")
	require.True(t, sc.IsRemote(), "extracted context must be marked remote")
	require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", sc.TraceID().String())

	out := propagation.HeaderCarrier{}
	propagator.Inject(ctx, out)
	require.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", out.Get("traceparent"))
}
