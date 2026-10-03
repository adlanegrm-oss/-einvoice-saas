package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
	"github.com/adlanegrm-oss/einvoice-saas/internal/observability"
)

func TestMetricsAndTraceMiddleware(t *testing.T) {
	ctx := context.Background()
	_, err := observability.InitTracer(ctx, "test")
	if err != nil {
		t.Fatalf("échec InitTracer: %v", err)
	}

	handler := middleware.ObservabilityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := observability.GetTraceID(r.Context())
		if traceID == "" {
			t.Errorf("traceID manquant dans le contexte")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("attendu statut 200, obtenu %d", rec.Code)
	}

	traceHeader := rec.Header().Get("X-Trace-ID")
	if traceHeader == "" {
		t.Errorf("header X-Trace-ID manquant dans la réponse")
	}

	// Vérifier l'exposition /metrics
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	observability.HandlerMetrics().ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Errorf("attendu statut 200 sur /metrics, obtenu %d", metricsRec.Code)
	}

	body := metricsRec.Body.String()
	if !contains(body, "einvoice_http_requests_total") {
		t.Errorf("métrique einvoice_http_requests_total absente du scrape /metrics")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}