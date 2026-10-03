package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type AuditEvent struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	InvoiceID     string    `json:"invoice_id"`
	SequenceNum   int64     `json:"sequence_num"`
	EventType     string    `json:"event_type"`
	PayloadHash   string    `json:"payload_hash"`
	PrevEventHash string    `json:"prev_event_hash"`
	EventHash     string    `json:"event_hash"`
	CreatedAt     time.Time `json:"created_at"`
}

// ComputeLinkedLogHash calcule H_n = SHA256(seq || tenant || invoice || type || payloadHash || prevHash || time)
func ComputeLinkedLogHash(e AuditEvent) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s",
		e.SequenceNum,
		e.TenantID,
		e.InvoiceID,
		e.EventType,
		e.PayloadHash,
		e.PrevEventHash,
		e.CreatedAt.UTC().Format(time.RFC3339Nano),
	)))
	return hex.EncodeToString(hasher.Sum(nil))
}

// CreateNextEvent prépare le prochain événement scellé
func CreateNextEvent(tenantID, invoiceID, eventType, payloadHash string, prevEvent *AuditEvent) AuditEvent {
	now := time.Now().UTC()
	var seq int64 = 1
	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"

	if prevEvent != nil {
		seq = prevEvent.SequenceNum + 1
		prevHash = prevEvent.EventHash
	}

	event := AuditEvent{
		ID:            fmt.Sprintf("evt_%d_%s", seq, invoiceID),
		TenantID:      tenantID,
		InvoiceID:     invoiceID,
		SequenceNum:   seq,
		EventType:     eventType,
		PayloadHash:   payloadHash,
		PrevEventHash: prevHash,
		CreatedAt:     now,
	}
	event.EventHash = ComputeLinkedLogHash(event)
	return event
}

// VerifyAuditChain inspecte une séquence ordonnée d'événements et identifie toute rupture
func VerifyAuditChain(events []AuditEvent) (bool, error) {
	if len(events) == 0 {
		return true, nil
	}

	for i := 0; i < len(events); i++ {
		curr := events[i]

		computed := ComputeLinkedLogHash(curr)
		if curr.EventHash != computed {
			return false, fmt.Errorf("altération payload sur evt %s (seq %d): hash stocké %s, recalculé %s",
				curr.ID, curr.SequenceNum, curr.EventHash, computed)
		}

		if i == 0 {
			if curr.SequenceNum != 1 {
				return false, fmt.Errorf("l'événement initial doit débuter à seq 1, reçu %d", curr.SequenceNum)
			}
		} else {
			prev := events[i-1]
			if curr.SequenceNum != prev.SequenceNum+1 {
				return false, fmt.Errorf("rupture de séquence entre evt %s (%d) et %s (%d)",
					prev.ID, prev.SequenceNum, curr.ID, curr.SequenceNum)
			}
			if curr.PrevEventHash != prev.EventHash {
				return false, fmt.Errorf("rupture de chaîne cryptographique à l'evt %s (seq %d)",
					curr.ID, curr.SequenceNum)
			}
		}
	}

	return true, nil
}
