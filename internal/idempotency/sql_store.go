package idempotency

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var ErrRequestInProgress = errors.New("opération en cours de traitement pour cette clé d'idempotence")

type Record struct {
	TenantID       string
	IdempotencyKey string
	RequestHash    string
	Status         string
	ResponseCode   int
	ResponseBody   string
}

func ComputeHash(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// LockKey tente de verrouiller transactionnellement une clé pour un tenant donné.
func (s *Store) LockKey(ctx context.Context, tx *sql.Tx, tenantID, key, requestHash string, ttl time.Duration) (*Record, error) {
	if tenantID == "" || key == "" {
		return nil, errors.New("tenant_id et idempotency_key obligatoires")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(ttl)

	var existing Record
	querySelect := `SELECT tenant_id, idempotency_key, request_hash, status, response_code, response_body 
                    FROM idempotency_keys WHERE tenant_id = ? AND idempotency_key = ?`

	err := tx.QueryRowContext(ctx, querySelect, tenantID, key).Scan(
		&existing.TenantID, &existing.IdempotencyKey, &existing.RequestHash,
		&existing.Status, &existing.ResponseCode, &existing.ResponseBody,
	)

	if err == nil {
		if existing.RequestHash != requestHash {
			return nil, ErrPayloadMismatch
		}
		if existing.Status == "STARTED" {
			return nil, ErrRequestInProgress
		}
		return &existing, nil // Déjà complété avec succès
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("lecture clé idempotence: %w", err)
	}

	queryInsert := `INSERT INTO idempotency_keys (tenant_id, idempotency_key, request_hash, status, created_at, expires_at) 
                    VALUES (?, ?, ?, 'STARTED', ?, ?)`
	if _, err := tx.ExecContext(ctx, queryInsert, tenantID, key, requestHash, now, expiresAt); err != nil {
		return nil, fmt.Errorf("création verrou idempotence: %w", err)
	}

	return nil, nil // Nouveau verrou posé
}

// Complete enregistre la réponse finale associée à la clé.
func (s *Store) Complete(ctx context.Context, tx *sql.Tx, tenantID, key string, code int, body string) error {
	query := `UPDATE idempotency_keys SET status = 'COMPLETED', response_code = ?, response_body = ? 
              WHERE tenant_id = ? AND idempotency_key = ?`
	_, err := tx.ExecContext(ctx, query, code, body, tenantID, key)
	return err
}
