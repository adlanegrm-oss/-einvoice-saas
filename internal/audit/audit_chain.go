package audit

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type StatusEvent struct {
	ID             string
	OrganizationID string
	InvoiceID      string
	FromState      string
	ToState        string
	Actor          string
	Reason         string
	PayloadHash    string
	PrevEventHash  string
	EventHash      string
	CreatedAt      time.Time
}

func InitAuditSchema(ctx context.Context, db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS invoice_status_events (
		id TEXT PRIMARY KEY,
		organization_id TEXT NOT NULL,
		invoice_id TEXT NOT NULL,
		from_state TEXT NOT NULL,
		to_state TEXT NOT NULL,
		actor TEXT NOT NULL,
		reason TEXT,
		payload_hash TEXT NOT NULL,
		prev_event_hash TEXT NOT NULL,
		event_hash TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_audit_chain_order ON invoice_status_events(invoice_id, created_at ASC);
	`
	_, err := db.ExecContext(ctx, query)
	return err
}

func ComputeEventHash(event StatusEvent) string {
	h := sha256.New()
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s",
		event.ID,
		event.OrganizationID,
		event.InvoiceID,
		event.FromState,
		event.ToState,
		event.Actor,
		event.PayloadHash,
		event.PrevEventHash,
		event.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func AppendEvent(ctx context.Context, tx *sql.Tx, ev *StatusEvent) error {
	var lastHash sql.NullString
	queryLast := `SELECT event_hash FROM invoice_status_events 
	              WHERE invoice_id = ? ORDER BY created_at DESC LIMIT 1`
	err := tx.QueryRowContext(ctx, queryLast, ev.InvoiceID).Scan(&lastHash)
	
	if errors.Is(err, sql.ErrNoRows) || !lastHash.Valid {
		ev.PrevEventHash = "GENESIS"
	} else if err != nil {
		return err
	} else {
		ev.PrevEventHash = lastHash.String
	}

	ev.CreatedAt = time.Now().UTC()
	ev.EventHash = ComputeEventHash(*ev)

	insertQuery := `
		INSERT INTO invoice_status_events (
			id, organization_id, invoice_id, from_state, to_state, 
			actor, reason, payload_hash, prev_event_hash, event_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		ev.ID, ev.OrganizationID, ev.InvoiceID, ev.FromState, ev.ToState,
		ev.Actor, ev.Reason, ev.PayloadHash, ev.PrevEventHash, ev.EventHash, ev.CreatedAt,
	)
	return err
}

func VerifyAuditTrail(ctx context.Context, db *sql.DB, invoiceID string) (bool, error) {
	query := `SELECT id, organization_id, invoice_id, from_state, to_state, actor, 
	                 COALESCE(reason,''), payload_hash, prev_event_hash, event_hash, created_at 
	          FROM invoice_status_events 
	          WHERE invoice_id = ? 
	          ORDER BY created_at ASC`

	rows, err := db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	expectedPrevHash := "GENESIS"

	for rows.Next() {
		var ev StatusEvent
		if err := rows.Scan(&ev.ID, &ev.OrganizationID, &ev.InvoiceID, &ev.FromState, 
			&ev.ToState, &ev.Actor, &ev.Reason, &ev.PayloadHash, 
			&ev.PrevEventHash, &ev.EventHash, &ev.CreatedAt); err != nil {
			return false, err
		}

		if ev.PrevEventHash != expectedPrevHash {
			return false, fmt.Errorf("rupture de chaine sur l'evenement %s: attendu=%s, trouve=%s", 
				ev.ID, expectedPrevHash, ev.PrevEventHash)
		}

		computed := ComputeEventHash(ev)
		if computed != ev.EventHash {
			return false, fmt.Errorf("alteration cryptographique sur l'evenement %s: calcule=%s, stocke=%s", 
				ev.ID, computed, ev.EventHash)
		}

		expectedPrevHash = ev.EventHash
	}

	return true, nil
}
