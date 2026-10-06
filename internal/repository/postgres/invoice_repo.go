package postgres

import (
	"context"
	"database/sql"
	"errors"

	"einvoice-saas/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

type InvoiceRepo struct {
	db *sql.DB
}

func NewInvoiceRepo(db *sql.DB) *InvoiceRepo {
	return &InvoiceRepo{db: db}
}

func mapInvoiceCreateError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	if pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "uq_tenant_invoice_business":
			return repository.ErrDuplicateBusiness
		case "uq_tenant_document_sha256":
			return repository.ErrDuplicatePayload
		}
	}
	return err
}

func (r *InvoiceRepo) Create(ctx context.Context, inv *repository.InvoiceRecord) error {
	query := `
INSERT INTO invoices (
tenant_id, id, invoice_number, seller_identifier, buyer_identifier,
issue_date, currency, total_tax_inclusive, syntax, profile,
status, document_sha256, document_storage_key, created_at, updated_at
) VALUES (
$1, $2, $3, $4, $5,
$6, $7, $8, $9, $10,
$11, $12, $13, NOW(), NOW()
)
`
	_, err := r.db.ExecContext(ctx, query,
		inv.TenantID,
		inv.ID,
		inv.InvoiceNumber,
		inv.SellerIdentifier,
		inv.BuyerIdentifier,
		inv.IssueDate,
		inv.Currency,
		inv.TotalTaxInclusive,
		inv.Syntax,
		inv.Profile,
		string(inv.Status),
		inv.DocumentSHA256,
		inv.DocumentStorageKey,
	)
	if err != nil {
		return mapInvoiceCreateError(err)
	}
	return nil
}

func (r *InvoiceRepo) GetByID(ctx context.Context, tenantID, id string) (*repository.InvoiceRecord, error) {
	query := `
SELECT 
tenant_id, id, invoice_number, seller_identifier, buyer_identifier,
issue_date, currency, total_tax_inclusive, syntax, profile,
status, document_sha256, COALESCE(document_storage_key, ''), created_at, updated_at
FROM invoices
WHERE tenant_id = $1 AND id = $2
`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)

	var inv repository.InvoiceRecord
	var status string
	err := row.Scan(
		&inv.TenantID,
		&inv.ID,
		&inv.InvoiceNumber,
		&inv.SellerIdentifier,
		&inv.BuyerIdentifier,
		&inv.IssueDate,
		&inv.Currency,
		&inv.TotalTaxInclusive,
		&inv.Syntax,
		&inv.Profile,
		&status,
		&inv.DocumentSHA256,
		&inv.DocumentStorageKey,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	inv.Status = repository.InvoiceStatus(status)
	return &inv, nil
}

func (r *InvoiceRepo) GetBySHA256(ctx context.Context, tenantID, sha256Hash string) (*repository.InvoiceRecord, error) {
	query := `
SELECT 
tenant_id, id, invoice_number, seller_identifier, buyer_identifier,
issue_date, currency, total_tax_inclusive, syntax, profile,
status, document_sha256, COALESCE(document_storage_key, ''), created_at, updated_at
FROM invoices
WHERE tenant_id = $1 AND document_sha256 = $2
`
	row := r.db.QueryRowContext(ctx, query, tenantID, sha256Hash)

	var inv repository.InvoiceRecord
	var status string
	err := row.Scan(
		&inv.TenantID,
		&inv.ID,
		&inv.InvoiceNumber,
		&inv.SellerIdentifier,
		&inv.BuyerIdentifier,
		&inv.IssueDate,
		&inv.Currency,
		&inv.TotalTaxInclusive,
		&inv.Syntax,
		&inv.Profile,
		&status,
		&inv.DocumentSHA256,
		&inv.DocumentStorageKey,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	inv.Status = repository.InvoiceStatus(status)
	return &inv, nil
}

func (r *InvoiceRepo) UpdateStatus(ctx context.Context, tenantID, id string, targetStatus repository.InvoiceStatus) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var currentStatusStr string
	err = tx.QueryRowContext(ctx, `
SELECT status FROM invoices WHERE tenant_id = $1 AND id = $2 FOR UPDATE
`, tenantID, id).Scan(&currentStatusStr)

	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrNotFound
	}
	if err != nil {
		return err
	}

	currentStatus := repository.InvoiceStatus(currentStatusStr)
	if !repository.ValidateTransition(currentStatus, targetStatus) {
		return repository.ErrInvalidTransition
	}

	_, err = tx.ExecContext(ctx, `
UPDATE invoices
SET status = $1, updated_at = NOW()
WHERE tenant_id = $2 AND id = $3
`, string(targetStatus), tenantID, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
