package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrPayloadMismatch     = errors.New("idempotency: charge utile differente pour la meme cle")
	ErrConcurrentExecution = errors.New("idempotency: requete en cours de traitement par une autre instance")
)

type CachedResponse struct {
	Code int
	Body []byte
}

type Manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS idempotency_keys (
		organization_id TEXT NOT NULL,
		idempotency_key TEXT NOT NULL,
		request_hash TEXT NOT NULL,
		response_code INTEGER,
		response_body TEXT,
		status TEXT NOT NULL CHECK(status IN ('PROCESSING', 'RESOLVED')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		locked_until DATETIME NOT NULL,
		PRIMARY KEY (organization_id, idempotency_key)
	);`
	_, err := m.db.ExecContext(ctx, query)
	return err
}

func (m *Manager) AcquireLock(ctx context.Context, orgID, key, reqHash string, ttl time.Duration) (bool, *CachedResponse, error) {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return false, nil, err
	}
	defer tx.Rollback()

	var (
		existingHash string
		status       string
		code         sql.NullInt64
		body         sql.NullString
		lockedUntil  time.Time
	)

	query := `SELECT request_hash, status, response_code, response_body, locked_until 
	          FROM idempotency_keys WHERE organization_id = ? AND idempotency_key = ?`

	err = tx.QueryRowContext(ctx, query, orgID, key).Scan(&existingHash, &status, &code, &body, &lockedUntil)

	if errors.Is(err, sql.ErrNoRows) {
		insertQuery := `INSERT INTO idempotency_keys (organization_id, idempotency_key, request_hash, status, locked_until)
		                VALUES (?, ?, ?, 'PROCESSING', ?)`
		if _, err := tx.ExecContext(ctx, insertQuery, orgID, key, reqHash, time.Now().Add(ttl)); err != nil {
			return false, nil, err
		}
		return true, nil, tx.Commit()
	}

	if err != nil {
		return false, nil, err
	}

	if existingHash != reqHash {
		return false, nil, ErrPayloadMismatch
	}

	if status == "RESOLVED" {
		return false, &CachedResponse{Code: int(code.Int64), Body: []byte(body.String)}, nil
	}

	if time.Now().After(lockedUntil) {
		updateQuery := `UPDATE idempotency_keys SET locked_until = ? WHERE organization_id = ? AND idempotency_key = ?`
		if _, err := tx.ExecContext(ctx, updateQuery, time.Now().Add(ttl), orgID, key); err != nil {
			return false, nil, err
		}
		return true, nil, tx.Commit()
	}

	return false, nil, ErrConcurrentExecution
}

func (m *Manager) Resolve(ctx context.Context, orgID, key string, code int, body []byte) error {
	query := `UPDATE idempotency_keys 
	          SET status = 'RESOLVED', response_code = ?, response_body = ? 
	          WHERE organization_id = ? AND idempotency_key = ?`
	_, err := m.db.ExecContext(ctx, query, code, string(body), orgID, key)
	return err
}
