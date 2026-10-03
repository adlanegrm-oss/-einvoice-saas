package idempotency

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrConflictPayload = errors.New("409 IDEMPOTENCY_CONFLICT: même clé réémise avec un corps de requête différent")
	ErrPayloadMismatch = ErrConflictPayload
	ErrInProgress      = errors.New("opération en cours de traitement pour cette clé d'idempotence")
	ErrNotFound        = errors.New("clé d'idempotence non trouvée")
)

type ExecutionRecord struct {
	Status       string
	ResponseCode int
	ResponseBody string
	InvoiceID    string
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func ComputeHash(payload []byte) string {
	return HashPayload(payload)
}

func HashPayload(payload []byte) string {
	h := sha256.Sum256(payload)
	return hex.EncodeToString(h[:])
}

func (s *Store) LockKey(ctx context.Context, tx *sql.Tx, tenantID string, key string, requestHash string, ttl time.Duration) (*ExecutionRecord, error) {
	var (
		existingHash string
		status       string
		code         sql.NullInt64
		body         sql.NullString
	)

	query := `
SELECT request_hash, status, response_code, response_body
FROM idempotency_keys
WHERE tenant_id = $1 AND idempotency_key = $2
`
	err := tx.QueryRowContext(ctx, query, tenantID, key).Scan(&existingHash, &status, &code, &body)
	if err == nil {
		if existingHash != requestHash {
			return nil, ErrPayloadMismatch
		}
		if status == "PENDING" {
			return nil, ErrInProgress
		}
		return &ExecutionRecord{
			Status:       status,
			ResponseCode: int(code.Int64),
			ResponseBody: body.String,
		}, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	now := time.Now()
	insertQuery := `
INSERT INTO idempotency_keys (tenant_id, idempotency_key, request_hash, status, created_at, expires_at)
VALUES ($1, $2, $3, 'PENDING', $4, $5)
`
	_, err = tx.ExecContext(ctx, insertQuery, tenantID, key, requestHash, now, now.Add(ttl))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Store) LockOrGet(ctx context.Context, tenantID, key, requestHash string, ttl time.Duration) (*ExecutionRecord, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rec, err := s.LockKey(ctx, tx, tenantID, key, requestHash, ttl)
	if err != nil {
		return nil, err
	}
	if rec != nil {
		return rec, nil
	}

	return nil, tx.Commit()
}

// Complete enregistre à la fois le code HTTP et le corps JSON de la réponse.
func (s *Store) Complete(ctx context.Context, tx *sql.Tx, tenantID string, key string, statusCode int, responseBody string) error {
	query := `
UPDATE idempotency_keys
SET status = 'COMPLETED',
    response_code = $1,
    response_body = $2
WHERE tenant_id = $3 AND idempotency_key = $4
`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, statusCode, responseBody, tenantID, key)
	} else {
		_, err = s.db.ExecContext(ctx, query, statusCode, responseBody, tenantID, key)
	}
	return err
}

func (s *Store) CompleteWithBody(ctx context.Context, tx *sql.Tx, tenantID string, key string, statusCode int, responseBody string, invoiceID string) error {
	return s.Complete(ctx, tx, tenantID, key, statusCode, responseBody)
}
