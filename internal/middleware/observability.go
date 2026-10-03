package middleware

import (
	"net/http"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ObservabilityMiddleware injecte le traçage OTel, extrait le trace_id et observe les métriques Prometheus
func ObservabilityMiddleware(next http.Handler) http.Handler {
	propagator := otel.GetTextMapPropagator()
	tracer := observability.Tracer()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extraction du contexte distant si transmis par headers W3C TraceContext
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.target", r.URL.Path),
			),
		)
		defer span.End()

		traceID := observability.GetTraceID(ctx)
		w.Header().Set("X-Trace-ID", traceID)

		wrappedWriter := &statusResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

		duration := time.Since(start)
		span.SetAttributes(attribute.Int("http.status_code", wrappedWriter.statusCode))

		// Enregistrement des métriques Prometheus (en ignorant les endpoints internes de scrape)
		if r.URL.Path != "/metrics" && r.URL.Path != "/healthz" {
			observability.ObserveHTTPRequest(r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
		}
	})
}
