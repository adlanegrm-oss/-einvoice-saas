package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type CompleteEvidencePackage struct {
	ExportedAt     time.Time    `json:"exported_at_utc"`
	TenantID       string       `json:"tenant_id"`
	InvoiceID      string       `json:"invoice_id"`
	DocumentDigest string       `json:"document_digest_sha256"`
	EventsChain    []AuditEvent `json:"events_chain"`
	IsChainValid   bool         `json:"is_chain_valid"`
}

func CalculateChainHash(ev *AuditEvent) string {
	if ev == nil {
		return ""
	}

	if ev.EventType != "" {
		data := struct {
			EventID        string    `json:"event_id"`
			TenantID       string    `json:"tenant_id"`
			InvoiceID      string    `json:"invoice_id"`
			EventType      string    `json:"event_type"`
			Actor          string    `json:"actor"`
			TimestampUTC   time.Time `json:"timestamp_utc"`
			DocumentSHA256 string    `json:"document_sha256"`
			PayloadSummary string    `json:"payload_summary"`
			PreviousHash   string    `json:"previous_hash"`
		}{
			ev.EventID, ev.TenantID, ev.InvoiceID, ev.EventType,
			ev.Actor, ev.TimestampUTC, ev.DocumentSHA256,
			ev.PayloadSummary, ev.PreviousHash,
		}
		b, err := json.Marshal(data)
		if err != nil {
			return ""
		}
		sum := sha256.Sum256(b)
		return hex.EncodeToString(sum[:])
	}

	return chainHash(ev.PrevHash, ev.EventID, ev.TenantID,
		ev.InvoiceID, ev.CreatedAt, ev.Type)
}

func VerifyChain(events []AuditEvent) (bool, error) {
	const genesis = "0000000000000000000000000000000000000000000000000000000000000000"
	prev := genesis

	for _, ev := range events {
		previous, current := ev.PrevHash, ev.ChainHash
		if ev.EventType != "" {
			previous, current = ev.PreviousHash, ev.CurrentHash
		}

		if previous != prev {
			return false, fmt.Errorf("chain broken at event %s", ev.EventID)
		}
		if current == "" || CalculateChainHash(&ev) != current {
			return false, fmt.Errorf("integrity violation at event %s", ev.EventID)
		}
		prev = current
	}
	return true, nil
}

func (p *CompleteEvidencePackage) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}
