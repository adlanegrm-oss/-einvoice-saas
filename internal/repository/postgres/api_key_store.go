package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"einvoice-saas/internal/middleware"
)

var (
	ErrCredentialRevoked  = errors.New("api key credential has been revoked")
	ErrCredentialInactive = errors.New("api key credential is inactive")
	ErrTenantInactive     = errors.New("tenant account is inactive or disabled")
	ErrCredentialNotFound = errors.New("api key credential not found")
)

type PostgresAPIKeyStore struct {
	db *sql.DB
}

func NewPostgresAPIKeyStore(db *sql.DB) *PostgresAPIKeyStore {
	if db == nil {
		panic("PostgresAPIKeyStore requiert une connexion *sql.DB non nulle")
	}
	return &PostgresAPIKeyStore{db: db}
}

func (s *PostgresAPIKeyStore) FindTenantByKeyHash(ctx context.Context, hash string) (*middleware.TenantRecord, error) {
	if hash == "" {
		return nil, ErrCredentialNotFound
	}

	query := `
SELECT 
c.key_id,
c.tenant_id,
c.status,
c.revoked_at,
t.id
FROM api_credentials c
INNER JOIN tenants t ON t.id = c.tenant_id
WHERE c.secret_hash = $1
LIMIT 1;
`

	var (
		keyID      string
		tenantID   string
		credStatus string
		revokedAt  sql.NullTime
		matchedTID string
	)

	err := s.db.QueryRowContext(ctx, query, hash).Scan(
		&keyID,
		&tenantID,
		&credStatus,
		&revokedAt,
		&matchedTID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredentialNotFound
		}
		return nil, fmt.Errorf("requete api_credentials: %w", err)
	}

	if revokedAt.Valid {
		return nil, ErrCredentialRevoked
	}

	if credStatus != "ACTIVE" {
		return nil, ErrCredentialInactive
	}

	go func(kid string) {
		updateCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = s.db.ExecContext(updateCtx, `UPDATE api_credentials SET last_used_at = NOW() WHERE key_id = $1`, kid)
	}(keyID)

	return &middleware.TenantRecord{
		ID:      matchedTID,
		KeyID:   keyID,
		Active:  true,
		KeyHash: hash,
	}, nil
}
