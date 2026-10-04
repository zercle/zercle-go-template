// OpenTelemetry context propagation construction.
package telemetry

import (
	"go.opentelemetry.io/otel/propagation"
)

// NewPropagator returns the W3C Trace Context and Baggage propagator used to
// extract inbound parent contexts and inject outbound ones. HTTP servers and
// boundaries receive it explicitly so distributed traces link across services.
func NewPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
