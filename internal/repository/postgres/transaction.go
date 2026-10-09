package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type transactionContextKey struct{}

type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func executorFor(ctx context.Context, db *sql.DB) sqlExecutor {
	if tx, ok := ctx.Value(transactionContextKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}

type TransactionRunner struct {
	db *sql.DB
}

func NewTransactionRunner(db *sql.DB) *TransactionRunner {
	return &TransactionRunner{db: db}
}

func (r *TransactionRunner) WithinTransaction(
	ctx context.Context,
	lockKey string,
	fn func(context.Context) error,
) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("transaction_runner_database_is_nil")
	}
	if fn == nil {
		return fmt.Errorf("transaction_callback_is_nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin_transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if lockKey != "" {
		var lockResult int
		err = tx.QueryRowContext(
			ctx,
			"SELECT 1 FROM (SELECT pg_advisory_xact_lock(hashtext($1))) AS acquired",
			lockKey,
		).Scan(&lockResult)
		if err != nil {
			return fmt.Errorf("lock_idempotency_key: %w", err)
		}
	}

	txCtx := context.WithValue(ctx, transactionContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit_transaction: %w", err)
	}
	return nil
}
