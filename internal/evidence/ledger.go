package evidence

import (
"crypto/sha256"
"encoding/hex"
"encoding/json"
"fmt"
"time"
)

type AuditEvent struct {
EventID        string    `json:"event_id"`
TenantID       string    `json:"tenant_id"`
InvoiceID      string    `json:"invoice_id"`
EventType      string    `json:"event_type"`
Actor          string    `json:"actor"`
TimestampUTC   time.Time `json:"timestamp_utc"`
DocumentSHA256 string    `json:"document_sha256"`
PayloadSummary string    `json:"payload_summary"`
PreviousHash   string    `json:"previous_hash"`
CurrentHash    string    `json:"current_hash"`
}

func CalculateChainHash(ev *AuditEvent) string {
h := sha256.New()
h.Write([]byte(ev.TenantID))
h.Write([]byte(ev.InvoiceID))
h.Write([]byte(ev.EventType))
h.Write([]byte(ev.Actor))
h.Write([]byte(ev.TimestampUTC.Format(time.RFC3339Nano)))
h.Write([]byte(ev.DocumentSHA256))
h.Write([]byte(ev.PayloadSummary))
h.Write([]byte(ev.PreviousHash))
return hex.EncodeToString(h.Sum(nil))
}

type CompleteEvidencePackage struct {
ExportedAt     time.Time    `json:"exported_at_utc"`
TenantID       string       `json:"tenant_id"`
InvoiceID      string       `json:"invoice_id"`
DocumentDigest string       `json:"document_digest_sha256"`
EventsChain    []AuditEvent `json:"events_chain"`
IsChainValid   bool         `json:"is_chain_valid"`
}

func VerifyChain(events []AuditEvent) (bool, error) {
if len(events) == 0 {
return true, nil
}

var expectedPrev string = "0000000000000000000000000000000000000000000000000000000000000000"

for idx, ev := range events {
if idx == 0 && ev.PreviousHash != expectedPrev {
return false, fmt.Errorf("genesis event has corrupted previous hash: %s", ev.PreviousHash)
}
if idx > 0 && ev.PreviousHash != events[idx-1].CurrentHash {
return false, fmt.Errorf("chain broken at event %s: expected prev %s, got %s", ev.EventID, events[idx-1].CurrentHash, ev.PreviousHash)
}

computed := CalculateChainHash(&ev)
if computed != ev.CurrentHash {
return false, fmt.Errorf("integrity violation at event %s: recalculation mismatch", ev.EventID)
}
expectedPrev = ev.CurrentHash
}

return true, nil
}

func (p *CompleteEvidencePackage) ToJSON() ([]byte, error) {
return json.MarshalIndent(p, "", "  ")
}
