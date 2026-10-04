package as4

import (
"crypto/sha256"
"encoding/hex"
"errors"
"fmt"
"sync"
"time"
)

var (
ErrReplayDetected = errors.New("rejeu detecte : message_id AS4 deja traite pour ce tenant")
ErrSignatureInvalid = errors.New("signature electronique AS4 invalide")
)

type AS4Message struct {
TenantID       string    `json:"tenant_id"`
MessageID      string    `json:"message_id"`
ConversationID string    `json:"conversation_id"`
Action         string    `json:"action"` // e.g., http://docs.oasis-open.org/ebxml-msg/ebms/v3.0/ns/core/200704/test
Payload        []byte    `json:"payload"`
PayloadDigest  string    `json:"payload_digest"`
Timestamp      time.Time `json:"timestamp"`
}

type AS4Receipt struct {
RefToMessageID string    `json:"ref_to_message_id"`
ReceiptID      string    `json:"receipt_id"`
Status         string    `json:"status"` // RECEIVED, REJECTED
NonRepudiation string    `json:"non_repudiation_token"`
ProcessedAt    time.Time `json:"processed_at"`
}

type ReplayStore struct {
mu       sync.Mutex
seenMsgs map[string]time.Time
}

func NewReplayStore() *ReplayStore {
return &ReplayStore{seenMsgs: make(map[string]time.Time)}
}

func (s *ReplayStore) CheckAndMark(tenantID, messageID string) error {
s.mu.Lock()
defer s.mu.Unlock()

key := fmt.Sprintf("%s:%s", tenantID, messageID)
if _, exists := s.seenMsgs[key]; exists {
return ErrReplayDetected
}

s.seenMsgs[key] = time.Now()
return nil
}

func ProcessInboundAS4(msg AS4Message, replay *ReplayStore) (*AS4Receipt, error) {
if err := replay.CheckAndMark(msg.TenantID, msg.MessageID); err != nil {
return nil, err
}

h := sha256.Sum256(msg.Payload)
calculatedDigest := hex.EncodeToString(h[:])
if msg.PayloadDigest != "" && msg.PayloadDigest != calculatedDigest {
return nil, ErrSignatureInvalid
}

receipt := &AS4Receipt{
RefToMessageID: msg.MessageID,
ReceiptID:      fmt.Sprintf("REC-%d", time.Now().UnixNano()),
Status:         "RECEIVED",
NonRepudiation: calculatedDigest,
ProcessedAt:    time.Now().UTC(),
}
return receipt, nil
}