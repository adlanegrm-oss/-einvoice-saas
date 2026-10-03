package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Métriques HTTP globales
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "einvoice_http_requests_total",
			Help: "Nombre total de requêtes HTTP reçues",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "einvoice_http_request_duration_seconds",
			Help:    "Latence des requêtes HTTP en secondes",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// Métriques Métier & Facturation
	InvoicesProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "einvoice_invoices_processed_total",
			Help: "Nombre total de factures traitées par statut final",
		},
		[]string{"format", "status"},
	)

	ValidationErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "einvoice_validation_errors_total",
			Help: "Nombre d'erreurs de validation sémantique/syntaxique par règle",
		},
		[]string{"rule_id", "severity"},
	)

	// Métriques Transport AS4 & Outbox
	AS4MessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "einvoice_as4_messages_total",
			Help: "Nombre de messages AS4 émis par issue",
		},
		[]string{"receiver_id", "status"}, // status: "success", "retry", "dlq"
	)

	AS4OutboxQueueLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "einvoice_as4_outbox_queue_length",
			Help: "Nombre actuel de messages en attente dans la file Outbox",
		},
	)

	AS4RetriesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "einvoice_as4_retries_total",
			Help: "Nombre total de tentatives de rejeu AS4 exécutées",
		},
	)

	AS4DLQTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "einvoice_as4_dlq_total",
			Help: "Nombre total de messages transférés vers la Dead Letter Queue",
		},
	)
)

// HandlerMetrics expose le endpoint HTTP /metrics
func HandlerMetrics() http.Handler {
	return promhttp.Handler()
}

// ObserveHTTPRequest enregistre la métrique d'une requête HTTP
func ObserveHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}