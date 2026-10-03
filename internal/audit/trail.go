package audit

import (
"crypto/sha256"
"encoding/hex"
"fmt"
"time"
)

type Event struct {
TenantID      string    `json:"tenant_id"`
InvoiceID     string    `json:"invoice_id"`
SequenceID    int64     `json:"sequence_id"`
EventType     string    `json:"event_type"`
PayloadHash   string    `json:"payload_hash"`
PrevEventHash string    `json:"prev_event_hash"`
EventHash     string    `json:"event_hash"`
RecordedAt    time.Time `json:"recorded_at"`
}

func ComputePayloadHash(payload []byte) string {
h := sha256.Sum256(payload)
return hex.EncodeToString(h[:])
}

// ComputeRawEventHash évite le conflit avec ComputeEventHash(StatusEvent) existant dans audit_chain.go
func ComputeRawEventHash(prevHash string, payloadHash string, eventType string, recordedAt time.Time) string {
raw := fmt.Sprintf("%s|%s|%s|%s", prevHash, payloadHash, eventType, recordedAt.UTC().Format(time.RFC3339Nano))
h := sha256.Sum256([]byte(raw))
return hex.EncodeToString(h[:])
}

func VerifyEventChain(events []Event) (bool, int) {
var currentPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"

for i, evt := range events {
if evt.PrevEventHash != currentPrevHash {
return false, i
}
expectedHash := ComputeRawEventHash(evt.PrevEventHash, evt.PayloadHash, evt.EventType, evt.RecordedAt)
if evt.EventHash != expectedHash {
return false, i
}
currentPrevHash = evt.EventHash
}
return true, -1
}
