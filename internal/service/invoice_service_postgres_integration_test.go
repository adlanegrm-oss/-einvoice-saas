package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"einvoice-saas/internal/model"
	"einvoice-saas/internal/repository"
	pgrepo "einvoice-saas/internal/repository/postgres"
	"einvoice-saas/internal/validator"

	"github.com/shopspring/decimal"
)

var errInjectedAfterIdempotencySave = errors.New("injected failure after idempotency save")

type failAfterSaveIdempotencyRepo struct {
	repository.IdempotencyRepository
	injectedErr error
}

func (r *failAfterSaveIdempotencyRepo) Save(
	ctx context.Context,
	record *repository.IdempotencyRecord,
) error {
	if err := r.IdempotencyRepository.Save(ctx, record); err != nil {
		return err
	}
	return r.injectedErr
}

func TestPostgresTransactionRollsBackInvoiceAuditAndIdempotency(t *testing.T) {
	dsn := os.Getenv("EINVOICE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set EINVOICE_TEST_DATABASE_URL to run the PostgreSQL integration test")
	}

	// Protection : ce test ne peut cibler que la base de test dédiée.
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("invalid test database URL: %v", err)
	}
	if parsed.Host != "127.0.0.1:5433" ||
		parsed.Path != "/einvoice_test_tx" {
		t.Fatalf(
			"refusing to run against unexpected database; want host 127.0.0.1:5433 and database /einvoice_test_tx",
		)
	}

	db, err := pgrepo.OpenDB(dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	tenantID := "it_tx_" + suffix
	invoiceID := "IT-" + suffix
	idemKey := "it-key-" + suffix

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO tenants (id, name) VALUES ($1, $2)`,
		tenantID,
		"PostgreSQL rollback integration test",
	)
	if err != nil {
		t.Fatalf("create isolated test tenant: %v", err)
	}

	invoiceRepo := pgrepo.NewInvoiceRepo(db)
	eventRepo := pgrepo.NewEventRepo(db)
	idemRepo := pgrepo.NewIdempotencyRepo(db)
	failingIdemRepo := &failAfterSaveIdempotencyRepo{
		IdempotencyRepository: idemRepo,
		injectedErr:           errInjectedAfterIdempotencySave,
	}
	runner := pgrepo.NewTransactionRunner(db)

	validate := func(
		_ []byte,
		_ validator.ValidationProfile,
	) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		return "UBL-2.1", &model.CanonicalInvoice{
			ID:               invoiceID,
			IssueDate:        "2026-10-08",
			Seller:           model.Party{LegalID: "SELLER-" + suffix},
			Buyer:            model.Party{LegalID: "BUYER-" + suffix},
			DocumentCurrency: "EUR",
			Totals: model.Totals{
				TaxInclusiveAmount: decimal.NewFromInt(100),
			},
		}, true, nil, nil
	}

	svc := NewInvoiceServiceWithTransactionRunner(
		invoiceRepo,
		eventRepo,
		failingIdemRepo,
		validate,
		runner,
	)

	result, ingestErr := svc.IngestInvoice(ctx, IngestionCommand{
		TenantID:       tenantID,
		IdempotencyKey: idemKey,
		Profile:        validator.ProfileEN16931,
		RawXML:         []byte("<invoice id=\"" + suffix + "\"/>"),
		Actor:          "integration-test",
	})

	if !errors.Is(ingestErr, errInjectedAfterIdempotencySave) {
		t.Fatalf(
			"expected injected error after idempotency insert, got result=%+v err=%v",
			result,
			ingestErr,
		)
	}
	if result != nil {
		t.Fatalf("expected nil result after rollback, got %+v", result)
	}

	var invoiceCount, eventCount, idempotencyCount int64

	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM invoices
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, invoiceID).Scan(&invoiceCount)
	if err != nil {
		t.Fatalf("count invoices: %v", err)
	}

	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM invoice_events
		WHERE tenant_id = $1 AND invoice_id = $2
	`, tenantID, invoiceID).Scan(&eventCount)
	if err != nil {
		t.Fatalf("count audit events: %v", err)
	}

	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM idempotency_keys
		WHERE tenant_id = $1 AND key = $2
	`, tenantID, idemKey).Scan(&idempotencyCount)
	if err != nil {
		t.Fatalf("count idempotency keys: %v", err)
	}

	if invoiceCount != 0 || eventCount != 0 || idempotencyCount != 0 {
		t.Errorf(
			"transaction rollback failed: invoices=%d audit_events=%d idempotency_keys=%d",
			invoiceCount,
			eventCount,
			idempotencyCount,
		)
		return
	}

	// Nettoyage limité au tenant créé par ce test, après vérification du rollback.
	if _, err := db.ExecContext(
		ctx,
		`DELETE FROM tenants WHERE id = $1`,
		tenantID,
	); err != nil {
		t.Errorf("clean up isolated test tenant: %v", err)
	}

	t.Log("PASS: invoice, audit event and idempotency key were all rolled back")
}
