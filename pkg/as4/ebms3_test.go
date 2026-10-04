package as4

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestAS4ReplayAndReceipt(t *testing.T) {
	store := NewReplayStore()
	payload := []byte("<Invoice>TEST</Invoice>")
	h := sha256.Sum256(payload)
	digest := hex.EncodeToString(h[:])

	msg := AS4Message{
		TenantID:      "tenant-alpha",
		MessageID:     "msg-001",
		SenderID:      "sender-fr-01",
		Payload:       payload,
		PayloadDigest: digest,
		Timestamp:     time.Now().UTC(),
	}

	receipt, err := ProcessInboundAS4(msg, store)
	if err != nil || receipt.Status != "RECEIVED" {
		t.Fatalf("reception AS4 valide en echec: %v", err)
	}

	_, errReplay := ProcessInboundAS4(msg, store)
	if errReplay != ErrReplayDetected {
		t.Fatalf("rejeu non bloque: %v", errReplay)
	}
}
