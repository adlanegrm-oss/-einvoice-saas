package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const ServiceName = "einvoice-saas"

var tracer trace.Tracer

func init() {
	tracer = otel.Tracer(ServiceName)
}

// Tracer retourne l'instance active
func Tracer() trace.Tracer {
	return tracer
}

// InitTracer initialise le TracerProvider global OpenTelemetry
func InitTracer(ctx context.Context, env string) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(ServiceName),
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("création de ressource OpenTelemetry: %w", err)
	}

	// Utilise un BatchSpanProcessor (ou SimpleSpanProcessor sans exporter distant en mode local)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// GetTraceID extrait la chaîne hexadécimale du trace_id courant pour l'injection dans les logs
func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	// Fallback pour corrélation si pas encore de span active
	return fmt.Sprintf("tr-%d", time.Now().UnixNano())
}
