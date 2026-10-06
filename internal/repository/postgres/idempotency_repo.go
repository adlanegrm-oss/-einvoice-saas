package postgres

import (
	"context"
	"database/sql"
	"errors"

	"einvoice-saas/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

type IdempotencyRepo struct {
	db *sql.DB
}

func NewIdempotencyRepo(db *sql.DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

func mapIdempotencySaveError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	if pgErr.Code == "23505" && pgErr.ConstraintName == "idempotency_keys_pkey" {
		return repository.ErrIdempotencyKeyExists
	}

	return err
}

func (r *IdempotencyRepo) Save(ctx context.Context, record *repository.IdempotencyRecord) error {
	query := `
INSERT INTO idempotency_keys (
tenant_id, key, request_hash, invoice_id,
response_status, response_body, created_at, expires_at
) VALUES (
$1, $2, $3, $4,
$5, $6, NOW(), $7
)
`
	_, err := r.db.ExecContext(ctx, query,
		record.TenantID,
		record.Key,
		record.RequestHash,
		record.InvoiceID,
		record.ResponseStatus,
		record.ResponseBody,
		record.ExpiresAt,
	)
	if err != nil {
		return mapIdempotencySaveError(err)
	}
	return nil
}

func (r *IdempotencyRepo) Get(ctx context.Context, tenantID, key string) (*repository.IdempotencyRecord, error) {
	query := `
SELECT 
tenant_id, key, request_hash, COALESCE(invoice_id, ''),
response_status, response_body, created_at, expires_at
FROM idempotency_keys
WHERE tenant_id = $1 AND key = $2
`
	row := r.db.QueryRowContext(ctx, query, tenantID, key)

	var rec repository.IdempotencyRecord
	err := row.Scan(
		&rec.TenantID,
		&rec.Key,
		&rec.RequestHash,
		&rec.InvoiceID,
		&rec.ResponseStatus,
		&rec.ResponseBody,
		&rec.CreatedAt,
		&rec.ExpiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
