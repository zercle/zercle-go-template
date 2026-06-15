// OpenTelemetry tracer provider construction.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/zercle/zercle-go-template/internal/config"
)

// NewTracer builds a trace.TracerProvider from configuration. When
// cfg.OTel.Exporter is "none" it returns a no-op provider and a nil shutdown
// function so callers can safely skip shutdown. For "otlp" it configures an
// OTLP HTTP exporter pointing at cfg.OTel.Endpoint, a resource carrying the
// service name, and a TraceIDRatioBased sampler.
func NewTracer(ctx context.Context, cfg *config.Config) (*trace.TracerProvider, func(context.Context) error, error) {
	if cfg.OTel.Exporter == "none" {
		return trace.NewTracerProvider(), nil, nil
	}

	if cfg.OTel.Endpoint == "" {
		return nil, nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_EXPORTER=%s", cfg.OTel.Exporter)
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.OTel.Endpoint))
	if err != nil {
		return nil, nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", cfg.OTel.ServiceName),
	))
	if err != nil {
		return nil, nil, fmt.Errorf("create OTel resource: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(cfg.OTel.Sampling)),
	)

	return provider, provider.Shutdown, nil
}
