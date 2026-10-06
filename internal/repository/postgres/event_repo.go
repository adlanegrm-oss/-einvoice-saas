package postgres

import (
	"context"
	"database/sql"
	"errors"

	"einvoice-saas/internal/evidence"
	"einvoice-saas/internal/repository"
)

const GenesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

type EventRepo struct {
	db *sql.DB
}

func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Append(ctx context.Context, ev *repository.InvoiceEventRecord) error {
	// 1. Contrôle cryptographique interne (CalculateChainHash)
	auditEv := evidence.AuditEvent{
		TenantID:       ev.TenantID,
		InvoiceID:      ev.InvoiceID,
		EventType:      ev.EventType,
		Actor:          ev.Actor,
		TimestampUTC:   ev.TimestampUTC,
		DocumentSHA256: ev.DocumentSHA256,
		PayloadSummary: ev.PayloadSummary,
		PreviousHash:   ev.PreviousHash,
	}
	expectedHash := evidence.CalculateChainHash(&auditEv)
	if ev.CurrentHash != expectedHash {
		return repository.ErrInvalidEventHash
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 2. Verrou pessimiste sur la facture parente pour sérialiser la chaîne (évite le conflit sur sequence 1)
	var parentID string
	err = tx.QueryRowContext(ctx, `
SELECT id FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE
`, ev.TenantID, ev.InvoiceID).Scan(&parentID)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrNotFound
	}
	if err != nil {
		return err
	}

	// 3. Récupération du dernier état de la chaîne
	queryLatest := `
SELECT sequence, current_hash
FROM invoice_events
WHERE tenant_id = $1 AND invoice_id = $2
ORDER BY sequence DESC
LIMIT 1
`
	var lastSeq int
	var lastHash string
	err = tx.QueryRowContext(ctx, queryLatest, ev.TenantID, ev.InvoiceID).Scan(&lastSeq, &lastHash)

	if errors.Is(err, sql.ErrNoRows) {
		if ev.Sequence != 1 || ev.PreviousHash != GenesisHash {
			return repository.ErrBrokenHashChain
		}
	} else if err != nil {
		return err
	} else {
		if ev.Sequence != lastSeq+1 || ev.PreviousHash != lastHash {
			return repository.ErrBrokenHashChain
		}
	}

	// 4. Insertion immuable
	queryInsert := `
INSERT INTO invoice_events (
tenant_id, invoice_id, sequence, event_id,
event_type, actor, document_sha256, payload_summary,
previous_hash, current_hash, timestamp_utc, created_at
) VALUES (
$1, $2, $3, $4,
$5, $6, $7, $8,
$9, $10, $11, NOW()
)
`
	_, err = tx.ExecContext(ctx, queryInsert,
		ev.TenantID,
		ev.InvoiceID,
		ev.Sequence,
		ev.EventID,
		ev.EventType,
		ev.Actor,
		ev.DocumentSHA256,
		ev.PayloadSummary,
		ev.PreviousHash,
		ev.CurrentHash,
		ev.TimestampUTC,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *EventRepo) GetHistory(ctx context.Context, tenantID, invoiceID string) ([]repository.InvoiceEventRecord, error) {
	query := `
SELECT 
tenant_id, invoice_id, sequence, event_id,
event_type, actor, document_sha256, COALESCE(payload_summary, ''),
previous_hash, current_hash, timestamp_utc, created_at
FROM invoice_events
WHERE tenant_id = $1 AND invoice_id = $2
ORDER BY sequence ASC
`
	rows, err := r.db.QueryContext(ctx, query, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []repository.InvoiceEventRecord
	for rows.Next() {
		var ev repository.InvoiceEventRecord
		if err := rows.Scan(
			&ev.TenantID,
			&ev.InvoiceID,
			&ev.Sequence,
			&ev.EventID,
			&ev.EventType,
			&ev.Actor,
			&ev.DocumentSHA256,
			&ev.PayloadSummary,
			&ev.PreviousHash,
			&ev.CurrentHash,
			&ev.TimestampUTC,
			&ev.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (r *EventRepo) GetLatestSequence(ctx context.Context, tenantID, invoiceID string) (int, string, error) {
	query := `
SELECT sequence, current_hash
FROM invoice_events
WHERE tenant_id = $1 AND invoice_id = $2
ORDER BY sequence DESC
LIMIT 1
`
	row := r.db.QueryRowContext(ctx, query, tenantID, invoiceID)

	var seq int
	var hash string
	err := row.Scan(&seq, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return seq, hash, nil
}
