package clearance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type PPFConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	ServiceCode  string // Code service Chorus Pro éventuel
}

type PPFConnector struct {
	cfg       PPFConfig
	mu        sync.RWMutex
	registry  map[string]*ClearanceResponse
}

func NewPPFConnector(cfg PPFConfig) *PPFConnector {
	return &PPFConnector{
		cfg:      cfg,
		registry: make(map[string]*ClearanceResponse),
	}
}

// SubmitInvoice simule le dépôt vers l'API /depot du Portail Public de Facturation
func (p *PPFConnector) SubmitInvoice(ctx context.Context, invoiceID string, payload []byte) (*ClearanceResponse, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("%w: payload de facture vide", ErrSubmissionFailed)
	}

	h := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(h[:8])
	depotID := fmt.Sprintf("PPF-FR-%d-%s", time.Now().UnixNano(), payloadHash)

	resp := &ClearanceResponse{
		ClearanceID: depotID,
		Status:      "CLEARED",
		QRCodeData:  fmt.Sprintf("https://ppf.dgfip.finances.gouv.fr/v1/verif/%s", depotID),
		ClearedAt:   time.Now().UTC(),
	}

	p.mu.Lock()
	p.registry[depotID] = resp
	p.mu.Unlock()

	return resp, nil
}

// CheckStatus consulte l'état du dépôt auprès du PPF
func (p *PPFConnector) CheckStatus(ctx context.Context, clearanceID string) (*ClearanceResponse, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	resp, exists := p.registry[clearanceID]
	if !exists {
		return nil, fmt.Errorf("%w: id=%s", ErrDocumentNotFound, clearanceID)
	}
	return resp, nil
}