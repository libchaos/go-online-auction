// Package telemetry provides a minimal OpenTelemetry tracing baseline.
//
// It installs a global tracer provider that exports spans to stdout via the
// stdouttrace exporter. This deliberately avoids OTLP/collector dependencies so
// the application runs (and stays safe) with no external tracing infrastructure
// present. The global W3C TraceContext propagator is also registered so trace
// context can be injected into outbound messages (e.g. NATS headers).
package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracerProvider configures the global tracer provider with a stdout
// exporter and registers the W3C TraceContext (and Baggage) propagators. It is
// safe to call in any environment: the stdout exporter needs no collector. It
// returns a shutdown function that flushes and stops the provider. On failure
// it logs and returns a no-op shutdown so the caller can proceed.
func InitTracerProvider(ctx context.Context, serviceName string) func(context.Context) error {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		//nolint:sloglint // this is a module
		slog.Error("telemetry: failed to create stdout trace exporter", "error", err)
		return func(_ context.Context) error { return nil }
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", serviceName)),
	)
	if err != nil {
		res = resource.Default()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown
}
