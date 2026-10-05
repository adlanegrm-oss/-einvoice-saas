package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	InvoicesProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "einvoice_processed_total",
			Help: "Nombre total de factures traitees par tenant et statut",
		},
		[]string{"tenant_id", "format", "status"},
	)

	TransmissionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "einvoice_transmission_duration_seconds",
			Help:    "Duree d'envoi vers PDP/PPF/AS4 en secondes",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tenant_id", "protocol"},
	)

	SchematronValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "einvoice_schematron_errors_total",
			Help: "Nombre total de violations normatives par regle",
		},
		[]string{"rule_id", "severity"},
	)
)
