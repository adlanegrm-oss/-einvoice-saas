package worker

import (
"context"
"database/sql"
"log"
"time"
)

type OutboxWorker struct {
db       *sql.DB
stopChan chan struct{}
}

func NewOutboxWorker(db *sql.DB) *OutboxWorker {
return &OutboxWorker{
db:       db,
stopChan: make(chan struct{}),
}
}

func (w *OutboxWorker) Start(concurrency int) {
log.Printf("[OUTBOX] Démarrage du worker outbox avec %d processeurs parallèles", concurrency)
for i := 0; i < concurrency; i++ {
go w.processLoop(i)
}
}

func (w *OutboxWorker) Stop() {
close(w.stopChan)
}

func (w *OutboxWorker) processLoop(workerID int) {
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for {
select {
case <-w.stopChan:
return
case <-ticker.C:
w.claimAndExecute(workerID)
}
}
}

func (w *OutboxWorker) claimAndExecute(workerID int) {
ctx := context.Background()
tx, err := w.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
if err != nil {
return
}
defer tx.Rollback()

query := `
SELECT id, tenant_id, aggregate_id, event_type, payload, retry_count, max_retries
FROM outbox_events
WHERE status IN ('PENDING', 'FAILED') AND next_attempt_at <= NOW()
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

var (
id          string
tenantID    string
aggregateID string
eventType   string
payloadRaw  []byte
retryCount  int
maxRetries  int
)

err = tx.QueryRowContext(ctx, query).Scan(&id, &tenantID, &aggregateID, &eventType, &payloadRaw, &retryCount, &maxRetries)
if err != nil {
return
}

_, _ = tx.ExecContext(ctx, "UPDATE outbox_events SET status = 'PROCESSING' WHERE id = $1", id)
if err := tx.Commit(); err != nil {
return
}

dispatchErr := w.dispatch(tenantID, aggregateID, eventType, payloadRaw)

if dispatchErr == nil {
_, _ = w.db.Exec("UPDATE outbox_events SET status = 'COMPLETED', updated_at = NOW() WHERE id = $1", id)
_, _ = w.db.Exec("UPDATE invoices SET transmission_status = 'DELIVERED' WHERE id = $1", aggregateID)
} else {
retryCount++
if retryCount >= maxRetries {
_, _ = w.db.Exec("UPDATE outbox_events SET status = 'DLQ', last_error = $1, updated_at = NOW() WHERE id = $2", dispatchErr.Error(), id)
_, _ = w.db.Exec("UPDATE invoices SET transmission_status = 'FAILED' WHERE id = $1", aggregateID)
} else {
backoffSec := (1 << retryCount) * 2
nextAttempt := time.Now().Add(time.Duration(backoffSec) * time.Second)
_, _ = w.db.Exec("UPDATE outbox_events SET status = 'FAILED', retry_count = $1, next_attempt_at = $2, last_error = $3, updated_at = NOW() WHERE id = $4",
retryCount, nextAttempt, dispatchErr.Error(), id)
}
}
}

func (w *OutboxWorker) dispatch(tenantID, invoiceID, eventType string, payload []byte) error {
return nil
}
