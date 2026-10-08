package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// EventType classifie l'ÃƒÂ©vÃƒÂ©nement dans le cycle de vie facture.
type EventType string

const (
	EventReceived           EventType = "INVOICE_RECEIVED"
	EventValidated          EventType = "INVOICE_VALIDATED"
	EventRejected           EventType = "INVOICE_REJECTED"
	EventExported           EventType = "FACTURX_EXPORTED"
	EventTransmitted        EventType = "INVOICE_TRANSMITTED"
	EventTransmissionFailed EventType = "TRANSMISSION_FAILED"
)

// AuditEvent entrÃƒÂ©e de journal append-only.
type AuditEvent struct {
	EventID        string            `json:"event_id"`
	TenantID       string            `json:"tenant_id"`
	InvoiceID      string            `json:"invoice_id"`
	Type           EventType         `json:"type"`
	Severity       string            `json:"severity"` // INFO / WARN / ERROR
	Message        string            `json:"message"`
	PayloadHash    string            `json:"payload_hash,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	PrevHash       string            `json:"prev_hash,omitempty"`
	ChainHash      string            `json:"chain_hash"`
	EventType      string            `json:"event_type,omitempty"`
	Actor          string            `json:"actor,omitempty"`
	TimestampUTC   time.Time         `json:"timestamp_utc,omitempty"`
	DocumentSHA256 string            `json:"document_sha256,omitempty"`
	PayloadSummary string            `json:"payload_summary,omitempty"`
	PreviousHash   string            `json:"previous_hash,omitempty"`
	CurrentHash    string            `json:"current_hash,omitempty"`
}

// Store journal en mÃƒÂ©moire (remplaÃƒÂ§able par Postgres plus tard).
type Store struct {
	mu       sync.RWMutex
	events   []AuditEvent
	lastHash string
}

func NewStore() *Store {
	return &Store{events: make([]AuditEvent, 0, 64), lastHash: "0000000000000000000000000000000000000000000000000000000000000000"}
}

func hashPayload(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func chainHash(prev, eventID, tenant, invoice string, ts time.Time, typ EventType) string {
	raw := prev + "|" + eventID + "|" + tenant + "|" + invoice + "|" + ts.UTC().Format(time.RFC3339Nano) + "|" + string(typ)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// Append ajoute un ÃƒÂ©vÃƒÂ©nement chaÃƒÂ®nÃƒÂ© (intÃƒÂ©gritÃƒÂ© simple).
func (s *Store) Append(tenantID, invoiceID string, typ EventType, severity, message string, payload []byte, meta map[string]string) AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	id := "evt-" + hashPayload(
		[]byte(tenantID + invoiceID + string(typ) + now.Format(time.RFC3339Nano)),
	)[:16]

	prev := s.lastHash

	payloadHash := ""
	if len(payload) > 0 {
		payloadHash = hashPayload(payload)
	}

	ev := AuditEvent{
		EventID:     id,
		TenantID:    tenantID,
		InvoiceID:   invoiceID,
		Type:        typ,
		Severity:    severity,
		Message:     message,
		PayloadHash: payloadHash,
		Metadata:    meta,
		CreatedAt:   now,
		PrevHash:    prev,
		ChainHash:   chainHash(prev, id, tenantID, invoiceID, now, typ),

		// CompatibilitÃƒÂ© ledger
		EventType:      string(typ),
		Actor:          "system",
		TimestampUTC:   now,
		DocumentSHA256: payloadHash,
		PayloadSummary: message,
		PreviousHash:   prev,
	}

	ev.CurrentHash = ev.ChainHash

	s.lastHash = ev.ChainHash
	s.events = append(s.events, ev)

	return ev
}

// ListByInvoice filtre par tenant + facture.
func (s *Store) ListByInvoice(tenantID, invoiceID string) []AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AuditEvent, 0)
	for _, e := range s.events {
		if e.TenantID == tenantID && e.InvoiceID == invoiceID {
			out = append(out, e)
		}
	}
	return out
}

// VerifyChain contrÃƒÂ´le la continuitÃƒÂ© des hash.
func (s *Store) VerifyChain() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prev := "0000000000000000000000000000000000000000000000000000000000000000"
	for _, e := range s.events {
		expected := chainHash(prev, e.EventID, e.TenantID, e.InvoiceID, e.CreatedAt, e.Type)
		if e.ChainHash != expected || e.PrevHash != prev {
			return errChainBroken
		}
		prev = e.ChainHash
	}
	return nil
}

var errChainBroken = &chainError{msg: "evidence: chaÃƒÂ®ne d'audit rompue"}

type chainError struct{ msg string }

func (e *chainError) Error() string { return e.msg }

// ExportJSON dump pour export lÃƒÂ©gal / support.
func (s *Store) ExportJSON(tenantID, invoiceID string) ([]byte, error) {
	return json.MarshalIndent(s.ListByInvoice(tenantID, invoiceID), "", "  ")
}
