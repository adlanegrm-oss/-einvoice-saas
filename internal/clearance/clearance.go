// Package clearance prépare la transmission des factures à une plateforme
// agréée / au portail public. Aucune intégration réelle n'existe encore.
package clearance

import (
	"context"
	"errors"
	"time"
)

// ErrNotConfigured : aucun connecteur réel n'est branché.
var ErrNotConfigured = errors.New("clearance : aucun connecteur PDP/PPF configuré")

type ClearanceResponse struct {
	ClearanceID string    `json:"clearance_id"`
	Status      string    `json:"status"` // CLEARED ou REJECTED
	QRCodeData  string    `json:"qr_code_data,omitempty"`
	ClearedAt   time.Time `json:"cleared_at"`
	Errors      []string  `json:"errors,omitempty"`
}

// Submitter est le contrat d'un connecteur de clearance.
type Submitter interface {
	SubmitInvoice(ctx context.Context, invoiceID string, payload []byte) (*ClearanceResponse, error)
}

// Service est le connecteur par défaut : il refuse toute soumission tant
// qu'aucune intégration réelle n'est configurée. Il ne simule JAMAIS un succès.
type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) SubmitInvoice(ctx context.Context, invoiceID string, payload []byte) (*ClearanceResponse, error) {
	return nil, ErrNotConfigured
}

// MockService renvoie un succès factice. À n'utiliser que dans les tests / en DEV.
type MockService struct{}

func (MockService) SubmitInvoice(ctx context.Context, invoiceID string, payload []byte) (*ClearanceResponse, error) {
	return &ClearanceResponse{
		ClearanceID: "MOCK-" + invoiceID,
		Status:      "CLEARED",
		ClearedAt:   time.Now(),
	}, nil
}
