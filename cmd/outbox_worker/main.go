package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/audit"
)

func ProcessOutboxBatch(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT id, tenant_id, invoice_id, destination_id, retry_count
		FROM transactional_outbox
		WHERE status IN ('PENDING', 'FAILED') AND next_retry_at <= NOW()
		ORDER BY id ASC
		LIMIT 10
		FOR UPDATE SKIP LOCKED;
	`
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	type OutboxTask struct {
		ID         int64
		TenantID   string
		InvoiceID  string
		DestID     string
		RetryCount int
	}

	var tasks []OutboxTask
	for rows.Next() {
		var t OutboxTask
		if err := rows.Scan(&t.ID, &t.TenantID, &t.InvoiceID, &t.DestID, &t.RetryCount); err != nil {
			return err
		}
		tasks = append(tasks, t)
	}

	for _, task := range tasks {
		dispatchErr := sendViaAS4(task.DestID, task.InvoiceID)

		if dispatchErr == nil {
			_, _ = tx.ExecContext(ctx, `
				UPDATE transactional_outbox 
				SET status = 'DELIVERED', updated_at = NOW() 
				WHERE id = $1;
			`, task.ID)

			_, _ = tx.ExecContext(ctx, `
				UPDATE invoices 
				SET status = 'TRANSMITTED', updated_at = NOW() 
				WHERE tenant_id = $1 AND id = $2;
			`, task.TenantID, task.InvoiceID)

			var prevHash string
			var prevSeq int64
			_ = tx.QueryRowContext(ctx, `
				SELECT event_hash, sequence_num FROM audit_events 
				WHERE tenant_id = $1 AND invoice_id = $2 
				ORDER BY sequence_num DESC LIMIT 1;
			`, task.TenantID, task.InvoiceID).Scan(&prevHash, &prevSeq)

			prevEvt := &audit.AuditEvent{EventHash: prevHash, SequenceNum: prevSeq}
			nextEvt := audit.CreateNextEvent(task.TenantID, task.InvoiceID, "INVOICE_TRANSMITTED", prevHash, prevEvt)

			_, _ = tx.ExecContext(ctx, `
				INSERT INTO audit_events (
					id, tenant_id, invoice_id, sequence_num, event_type, 
					payload_hash, prev_event_hash, event_hash, created_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
			`, nextEvt.ID, nextEvt.TenantID, nextEvt.InvoiceID, nextEvt.SequenceNum,
				nextEvt.EventType, nextEvt.PayloadHash, nextEvt.PrevEventHash, nextEvt.EventHash, nextEvt.CreatedAt)

		} else {
			nextRetry := time.Now().Add(time.Duration(1<<task.RetryCount) * 30 * time.Second)
			status := "FAILED"
			if task.RetryCount+1 >= 5 {
				status = "DEAD_LETTER"
			}

			_, _ = tx.ExecContext(ctx, `
				UPDATE transactional_outbox
				SET status = $1, retry_count = retry_count + 1, last_error = $2, next_retry_at = $3, updated_at = NOW()
				WHERE id = $4;
			`, status, dispatchErr.Error(), nextRetry, task.ID)
		}
	}

	return tx.Commit()
}

func sendViaAS4(destination, invoiceID string) error {
	// Adapter AS4 / PDP
	return nil
}

func main() {
	log.Println("Outbox worker ready.")
}
