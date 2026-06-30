// OpenTelemetry tracer provider construction.
package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/zercle/zercle-go-template/internal/config"
)

// NewTracerProvider builds a trace.TracerProvider from configuration. When
// cfg.OTel.Exporter is "none" it returns a no-op provider and a nil shutdown
// function so callers can safely skip shutdown. For "otlp" it configures an
// OTLP HTTP exporter pointing at cfg.OTel.Endpoint (treated as a base URL; the
// /v1/traces path is appended if missing), a resource carrying the service
// name, and a TraceIDRatioBased sampler.
func NewTracerProvider(ctx context.Context, cfg *config.Config) (*trace.TracerProvider, func(context.Context) error, error) {
	if cfg.OTel.Exporter == "none" {
		return trace.NewTracerProvider(), nil, nil
	}

	if cfg.OTel.Endpoint == "" {
		return nil, nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_EXPORTER=%s", cfg.OTel.Exporter)
	}

	endpoint := otlpTracesEndpoint(cfg.OTel.Endpoint)

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
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

// otlpTracesEndpoint normalizes a raw OTLP endpoint so it always points at the
// OTLP HTTP traces path. If the endpoint already ends with "/v1/traces" it is
// returned unchanged; otherwise any trailing slash is stripped and "/v1/traces"
// is appended.
func otlpTracesEndpoint(raw string) string {
	if strings.HasSuffix(raw, "/v1/traces") {
		return raw
	}
	return strings.TrimSuffix(raw, "/") + "/v1/traces"
}
