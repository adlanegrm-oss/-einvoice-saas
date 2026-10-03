package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type AuditEvent struct {
	ID                string
	InvoiceID         string
	SequenceNumber    int
	FromState         string
	ToState           string
	Actor             string
	Reason            string
	Timestamp         time.Time
	PayloadHash       string
	PreviousEventHash string
	EventHash         string
	TenantID          string
}

// ComputeEventHash calcule le hash cryptographique scellant l'événement
func ComputeEventHash(prevHash, invoiceID string, seq int, fromState, toState, actor, payloadHash string, ts time.Time) string {
	raw := fmt.Sprintf("%s|%s|%d|%s|%s|%s|%s|%s",
		prevHash,
		invoiceID,
		seq,
		fromState,
		toState,
		actor,
		payloadHash,
		ts.Format(time.RFC3339Nano),
	)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// AppendAuditEvent insère un événement en garantissant le chaînage strict
func AppendAuditEvent(ctx context.Context, tx *sql.Tx, invID, fromState, toState, actor, reason, payloadHash, tenantID string) (*AuditEvent, error) {
	// 1. Verrou de lecture sur le dernier événement de la facture
	var lastSeq int
	var lastHash string

	queryLast := `
		SELECT sequence_number, event_hash 
		FROM audit_events 
		WHERE invoice_id = ? 
		ORDER BY sequence_number DESC 
		LIMIT 1
	`
	err := tx.QueryRowContext(ctx, queryLast, invID).Scan(&lastSeq, &lastHash)
	if err == sql.ErrNoRows {
		lastSeq = 0
		lastHash = "GENESIS_BLOCK_00000000000000000000000000000000000000000000000000000000"
	} else if err != nil {
		return nil, fmt.Errorf("lecture dernier hash audit : %w", err)
	}

	newSeq := lastSeq + 1
	now := time.Now().UTC()
	eventID := fmt.Sprintf("evt_%s_%d", invID, newSeq)
	currentHash := ComputeEventHash(lastHash, invID, newSeq, fromState, toState, actor, payloadHash, now)

	insertQuery := `
		INSERT INTO audit_events (
			id, invoice_id, sequence_number, from_state, to_state, actor, 
			reason, timestamp, payload_hash, previous_event_hash, event_hash, tenant_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		eventID, invID, newSeq, fromState, toState, actor,
		reason, now, payloadHash, lastHash, currentHash, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("insertion événement audit chaîné : %w", err)
	}

	return &AuditEvent{
		ID:                eventID,
		InvoiceID:         invID,
		SequenceNumber:    newSeq,
		FromState:         fromState,
		ToState:           toState,
		Actor:             actor,
		Reason:            reason,
		Timestamp:         now,
		PayloadHash:       payloadHash,
		PreviousEventHash: lastHash,
		EventHash:         currentHash,
		TenantID:          tenantID,
	}, nil
}

// VerifyAuditTrail vérifie l'intégrité de la chaîne pour une facture donnée
func VerifyAuditTrail(ctx context.Context, db *sql.DB, invID string) (bool, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT sequence_number, from_state, to_state, actor, timestamp, payload_hash, previous_event_hash, event_hash 
		FROM audit_events 
		WHERE invoice_id = ? 
		ORDER BY sequence_number ASC
	`, invID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	expectedPrev := "GENESIS_BLOCK_00000000000000000000000000000000000000000000000000000000"
	expectedSeq := 1

	for rows.Next() {
		var seq int
		var fromState, toState, actor, payloadHash, prevHash, hash string
		var ts time.Time

		if err := rows.Scan(&seq, &fromState, &toState, &actor, &ts, &payloadHash, &prevHash, &hash); err != nil {
			return false, err
		}

		if seq != expectedSeq || prevHash != expectedPrev {
			return false, fmt.Errorf("rupture de séquence à l'étape %d: hash précédent incohérent", seq)
		}

		recomputed := ComputeEventHash(prevHash, invID, seq, fromState, toState, actor, payloadHash, ts)
		if recomputed != hash {
			return false, fmt.Errorf("altération détectée à l'étape %d: signature invalide", seq)
		}

		expectedPrev = hash
		expectedSeq++
	}

	return true, nil
}