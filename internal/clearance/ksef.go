package clearance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type KSeFConfig struct {
	EnvironmentURL string
	NIP            string // Identifiant fiscal polonais
	AuthToken      string
}

type KSeFConnector struct {
	cfg      KSeFConfig
	mu       sync.RWMutex
	registry map[string]*ClearanceResponse
}

func NewKSeFConnector(cfg KSeFConfig) *KSeFConnector {
	return &KSeFConnector{
		cfg:      cfg,
		registry: make(map[string]*ClearanceResponse),
	}
}

// SubmitInvoice simule la session KSeF, le chiffrement et la génération du KSeF Reference Number
func (k *KSeFConnector) SubmitInvoice(ctx context.Context, invoiceID string, payload []byte) (*ClearanceResponse, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("%w: flux FA_VAT KSeF vide", ErrSubmissionFailed)
	}

	datePrefix := time.Now().UTC().Format("20060102")
	hash := sha256.Sum256(payload)
	shortHash := hex.EncodeToString(hash[:4])

	// Format réglementaire : NIP-YYYYMMDD-HEXID
	ksefRef := fmt.Sprintf("%s-%s-%s", k.cfg.NIP, datePrefix, shortHash)

	// Simulation du document UPO (Urzędowe Poświadczenie Odbioru)
	upoXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Potwierdzenie xmlns="http://ksef.mf.gov.pl/schema/upo/v2-5">
    <NumerKSeF>%s</NumerKSeF>
    <KodPojazdu>SUCCESS</KodPojazdu>
    <Timestamp>%s</Timestamp>
</Potwierdzenie>`, ksefRef, time.Now().UTC().Format(time.RFC3339))

	resp := &ClearanceResponse{
		ClearanceID: ksefRef,
		Status:      "CLEARED",
		UPODocument: []byte(upoXML),
		QRCodeData:  fmt.Sprintf("https://ksef.mf.gov.pl/web/verify/%s", ksefRef),
		ClearedAt:   time.Now().UTC(),
	}

	k.mu.Lock()
	k.registry[ksefRef] = resp
	k.mu.Unlock()

	return resp, nil
}

func (k *KSeFConnector) CheckStatus(ctx context.Context, clearanceID string) (*ClearanceResponse, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	resp, exists := k.registry[clearanceID]
	if !exists {
		return nil, fmt.Errorf("%w: ref KSeF %s", ErrDocumentNotFound, clearanceID)
	}
	return resp, nil
}
