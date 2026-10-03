package evidence

import (
"crypto/sha256"
"encoding/hex"
"fmt"
"time"
)

type AuditEvent struct {
ID          string    `json:"id"`
SequenceID  int64     `json:"sequence_id"`
EventType   string    `json:"event_type"`
PayloadHash string    `json:"payload_hash"`
PrevHash    string    `json:"prev_hash"`
EventHash   string    `json:"event_hash"`
RecordedAt  time.Time `json:"recorded_at"`
}

type EvidenceReport struct {
InvoiceID          string       `json:"invoice_id"`
TenantID           string       `json:"tenant_id"`
InvoiceNumber      string       `json:"invoice_number"`
ComplianceStatus   string       `json:"compliance_status"`
TransmissionStatus string       `json:"transmission_status"`
EventsChain        []AuditEvent `json:"events_chain"`
ChainValid         bool         `json:"chain_valid"`
GeneratedAt        time.Time    `json:"generated_at"`
}

func ComputeEventHash(prevHash, payloadHash, eventType string, recordedAt time.Time) string {
raw := fmt.Sprintf("%s|%s|%s|%s", prevHash, payloadHash, eventType, recordedAt.UTC().Format(time.RFC3339Nano))
h := sha256.Sum256([]byte(raw))
return hex.EncodeToString(h[:])
}

func VerifyChain(events []AuditEvent) bool {
prev := "0000000000000000000000000000000000000000000000000000000000000000"
for _, e := range events {
if e.PrevHash != prev {
return false
}
expected := ComputeEventHash(e.PrevHash, e.PayloadHash, e.EventType, e.RecordedAt)
if e.EventHash != expected {
return false
}
prev = e.EventHash
}
return true
}
