// backend/internal/observability/telemetry.go
package observability

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryService provides observability capabilities
type TelemetryService struct {
	tracer trace.Tracer
	exporter *sdktrace.SpanProcessor
	tp       *sdktrace.TracerProvider
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(serviceName, serviceVersion string) (*TelemetryService, error) {
	service := &TelemetryService{}

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter
	traceExporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.TraceContext{})

	service.tracer = tp.Tracer(serviceName)
	service.exporter = &bsp
	service.tp = tp

	log.Println("Telemetry service initialized successfully")
	return service, nil
}

// StartSpan starts a new span with the given name and options
func (ts *TelemetryService) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ts.tracer.Start(ctx, spanName, opts...)
}

// SetAttribute sets an attribute on the current span
func (ts *TelemetryService) SetAttribute(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		switch v := value.(type) {
		case string:
			span.SetAttributes(semconv.String(key, v))
		case int:
			span.SetAttributes(semconv.Int(key, v))
		case int64:
			span.SetAttributes(semconv.Int64(key, v))
		case float64:
			span.SetAttributes(semconv.Float64(key, v))
		case bool:
			span.SetAttributes(semconv.Bool(key, v))
		}
	}
}

// AddEvent adds an event to the current span
func (ts *TelemetryService) AddEvent(ctx context.Context, name string, attrs ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(name, attrs...)
	}
}

// RecordError records an error on the current span
func (ts *TelemetryService) RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err)
	}
}

// Shutdown shuts down the tracer provider
func (ts *TelemetryService) Shutdown(ctx context.Context) error {
	if ts.tp != nil {
		return ts.tp.Shutdown(ctx)
	}
	return nil
}

// WithContext adds telemetry context to the given context
func (ts *TelemetryService) WithContext(ctx context.Context) context.Context {
	return ctx
}

// GetTracer returns the tracer
func (ts *TelemetryService) GetTracer() trace.Tracer {
	return ts.tracer
}