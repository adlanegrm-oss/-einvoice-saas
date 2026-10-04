package outbox

import (
	"context"
	"database/sql"
	"math"
	"math/rand"
	"time"
)

type OutboxWorker struct {
	db *sql.DB
}

func NewOutboxWorker(db *sql.DB) *OutboxWorker {
	return &OutboxWorker{db: db}
}

func (w *OutboxWorker) ProcessPendingEvents(ctx context.Context, sendFunc func(ctx context.Context, tenantID, aggregateID, payload string) error) error {
	tx, err := w.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
SELECT id, tenant_id, aggregate_id, event_type, payload, attempts, max_attempts
FROM outbox_events
WHERE status IN ('PENDING', 'FAILED') AND next_attempt_at <= CURRENT_TIMESTAMP
ORDER BY next_attempt_at ASC
LIMIT 10
`
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, tenantID, aggregateID, eventType, payload string
		var attempts, maxAttempts int

		if err := rows.Scan(&id, &tenantID, &aggregateID, &eventType, &payload, &attempts, &maxAttempts); err != nil {
			continue
		}

		errSend := sendFunc(ctx, tenantID, aggregateID, payload)
		if errSend != nil {
			attempts++
			if attempts >= maxAttempts {
				tx.ExecContext(ctx, `
UPDATE outbox_events 
SET status = 'DEAD_LETTER', attempts = ?, processed_at = CURRENT_TIMESTAMP 
WHERE id = ?`, attempts, id)
			} else {
				backoff := time.Duration(math.Pow(2, float64(attempts))) * 5 * time.Second
				jitter := time.Duration(rand.Int63n(int64(2 * time.Second)))
				nextRun := time.Now().Add(backoff + jitter)

				tx.ExecContext(ctx, `
UPDATE outbox_events 
SET status = 'FAILED', attempts = ?, next_attempt_at = ? 
WHERE id = ?`, attempts, nextRun, id)
			}
		} else {
			tx.ExecContext(ctx, `
UPDATE outbox_events 
SET status = 'SENT', processed_at = CURRENT_TIMESTAMP 
WHERE id = ?`, id)
		}
	}

	return tx.Commit()
}
