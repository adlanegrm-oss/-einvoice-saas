package evidence

import (
"crypto/sha256"
"encoding/hex"
"fmt"
"time"
)

type EventLog struct {
Index       int64     `json:"index"`
EventType   string    `json:"event_type"`
PayloadHash string    `json:"payload_hash"`
PrevHash    string    `json:"prev_hash"`
EventHash   string    `json:"event_hash"`
Timestamp   time.Time `json:"timestamp"`
}

type EvidenceDossier struct {
InvoiceID      string     `json:"invoice_id"`
TenantID       string     `json:"tenant_id"`
DocumentHash   string     `json:"document_fingerprint_sha256"`
RulesetVersion string     `json:"ruleset_version"`
FinalState     string     `json:"final_state"`
ChainValid     bool       `json:"chain_valid"`
Events         []EventLog `json:"audit_trail"`
ExportedAt     time.Time  `json:"exported_at"`
}

func HashStep(prevHash, payloadHash, eventType string, t time.Time) string {
raw := fmt.Sprintf("%s|%s|%s|%s", prevHash, payloadHash, eventType, t.UTC().Format(time.RFC3339Nano))
h := sha256.Sum256([]byte(raw))
return hex.EncodeToString(h[:])
}

// VerifyHashChain audite la chaîne séquentielle récursive : H_n = SHA256(H_{n-1} + payload + type + timestamp)
func VerifyHashChain(events []EventLog) (bool, int) {
currentPrev := "0000000000000000000000000000000000000000000000000000000000000000"
for idx, e := range events {
if e.PrevHash != currentPrev {
return false, idx
}
expected := HashStep(e.PrevHash, e.PayloadHash, e.EventType, e.Timestamp)
if e.EventHash != expected {
return false, idx
}
currentPrev = e.EventHash
}
return true, -1
}
