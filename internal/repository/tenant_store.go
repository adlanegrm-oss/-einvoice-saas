package repository

import (
	"context"
	"database/sql"
	"errors"
)

type TenantContext struct {
	OrganizationID string
	UserID         string
	Role           string
}

type HardenedStore struct {
	db *sql.DB
}

func NewHardenedStore(db *sql.DB) *HardenedStore {
	return &HardenedStore{db: db}
}

func (s *HardenedStore) BootstrapAdmin(ctx context.Context, id, email, passwordHash string) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, created_at)
		VALUES (?, ?, ?, 'admin', CURRENT_TIMESTAMP)
		ON CONFLICT(email) DO NOTHING;
	`
	_, err := s.db.ExecContext(ctx, query, id, email, passwordHash)
	return err
}

func (s *HardenedStore) GetInvoice(ctx context.Context, t TenantContext, invoiceID string) (*sql.Row, error) {
	if t.OrganizationID == "" {
		return nil, errors.New("securite : organization_id absent du contexte d'execution")
	}

	query := `
		SELECT id, organization_id, amount_cents, status, canonical_xml_hash 
		FROM invoices 
		WHERE id = ? AND organization_id = ?
	`
	return s.db.QueryRowContext(ctx, query, invoiceID, t.OrganizationID), nil
}
