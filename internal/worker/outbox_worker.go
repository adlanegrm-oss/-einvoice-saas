package worker

import (
"context"
"database/sql"
"fmt"
"time"

"github.com/adlanegrm-oss/einvoice-saas/internal/connector"
)

type OutboxWorker struct {
db        *sql.DB
connector connector.NetworkConnector
stopChan  chan struct{}
}

func NewOutboxWorker(db *sql.DB, conn connector.NetworkConnector) *OutboxWorker {
return &OutboxWorker{
db:        db,
connector: conn,
stopChan:  make(chan struct{}),
}
}

func (w *OutboxWorker) Stop() {
close(w.stopChan)
}

func (w *OutboxWorker) ProcessNextBatch(ctx context.Context) (int, error) {
tx, err := w.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
if err != nil {
return 0, err
}
defer tx.Rollback()

query := `
SELECT id, tenant_id, aggregate_id, payload_json, retry_count, max_retries 
FROM outbox_events 
WHERE status IN ('PENDING', 'FAILED') AND next_attempt_at <= ?
ORDER BY created_at ASC LIMIT 1
`
var (
id, tenantID, invoiceID, payload string
retryCount, maxRetries           int
)
err = tx.QueryRowContext(ctx, query, time.Now().UTC()).Scan(&id, &tenantID, &invoiceID, &payload, &retryCount, &maxRetries)
if err == sql.ErrNoRows {
return 0, nil
}
if err != nil {
return 0, err
}

now := time.Now().UTC()
_, _ = tx.ExecContext(ctx, "UPDATE outbox_events SET status = 'PROCESSING', updated_at = ? WHERE id = ?", now, id)
if err := tx.Commit(); err != nil {
return 0, err
}

var connErr error
var receipt *connector.TransmissionReceipt
if w.connector != nil {
receipt, connErr = w.connector.Submit(ctx, tenantID, invoiceID, []byte(payload))
} else {
receipt = &connector.TransmissionReceipt{
MessageID:   fmt.Sprintf("mock-%d", time.Now().UnixNano()),
Status:      "ACCEPTED",
ReceiptHash: "mock-hash",
}
}

execNow := time.Now().UTC()
if connErr == nil && receipt != nil && receipt.Status == "ACCEPTED" {
_, _ = w.db.Exec("UPDATE outbox_events SET status = 'SUCCESS', updated_at = ? WHERE id = ?", execNow, id)
_, _ = w.db.Exec("UPDATE invoices SET transmission_status = 'ACCEPTED', updated_at = ? WHERE id = ?", execNow, invoiceID)

auditID := fmt.Sprintf("evt-sub-%d", time.Now().UnixNano())
_, _ = w.db.Exec(`
INSERT INTO audit_events (id, tenant_id, invoice_id, sequence_id, event_type, payload_hash, prev_hash, event_hash, recorded_at)
VALUES (?, ?, ?, 2, 'NETWORK_SUBMISSION_ACCEPTED', ?, '0000000000000000000000000000000000000000000000000000000000000000', ?, ?)
`, auditID, tenantID, invoiceID, receipt.ReceiptHash, receipt.ReceiptHash, execNow)
return 1, nil
}

retryCount++
if retryCount >= maxRetries {
errMsg := "dispatch error"
if connErr != nil {
errMsg = connErr.Error()
}
_, _ = w.db.Exec("UPDATE outbox_events SET status = 'DLQ', last_error = ?, updated_at = ? WHERE id = ?", errMsg, execNow, id)
_, _ = w.db.Exec("UPDATE invoices SET transmission_status = 'FAILED', updated_at = ? WHERE id = ?", execNow, invoiceID)
} else {
backoff := time.Duration(1<<retryCount) * time.Second
next := execNow.Add(backoff)
errMsg := "retry scheduled"
if connErr != nil {
errMsg = connErr.Error()
}
_, _ = w.db.Exec("UPDATE outbox_events SET status = 'FAILED', retry_count = ?, next_attempt_at = ?, last_error = ?, updated_at = ? WHERE id = ?",
retryCount, next, errMsg, execNow, id)
}
return 1, nil
}
