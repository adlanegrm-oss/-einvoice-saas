package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	_ "modernc.org/sqlite"
)

func setupTestEnvironment(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	ddl := `
CREATE TABLE idempotency_keys (
tenant_id TEXT NOT NULL,
idempotency_key TEXT NOT NULL,
request_hash TEXT NOT NULL,
status TEXT NOT NULL,
response_code INTEGER,
response_body TEXT,
created_at TIMESTAMP NOT NULL,
expires_at TIMESTAMP NOT NULL,
PRIMARY KEY (tenant_id, idempotency_key)
);

CREATE TABLE invoices_v2 (
id TEXT PRIMARY KEY,
tenant_id TEXT NOT NULL,
invoice_number TEXT NOT NULL,
status TEXT NOT NULL,
currency TEXT NOT NULL DEFAULT 'EUR',
issue_date TIMESTAMP NOT NULL,
seller_json TEXT NOT NULL,
customer_json TEXT NOT NULL,
total_ht_cents BIGINT NOT NULL,
total_vat_cents BIGINT NOT NULL,
total_ttc_cents BIGINT NOT NULL,
is_validated BOOLEAN NOT NULL DEFAULT FALSE,
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
CONSTRAINT uq_tenant_invoice_number UNIQUE (tenant_id, invoice_number)
);

CREATE TABLE outbox_events (
id TEXT PRIMARY KEY,
tenant_id TEXT NOT NULL,
aggregate_id TEXT NOT NULL,
event_type TEXT NOT NULL,
payload_json TEXT NOT NULL,
status TEXT NOT NULL,
created_at TIMESTAMP NOT NULL
);`

	if _, err := db.Exec(ddl); err != nil {
		t.Fatalf("ddl init: %v", err)
	}
	return db
}

func TestMultiTenantEmissionPipeline(t *testing.T) {
	db := setupTestEnvironment(t)
	svc := NewMultiTenantEmissionService(db)
	ctx := context.Background()

	makeInvoice := func(num string) *invoice.Invoice {
		return &invoice.Invoice{
			Number:    num,
			Customer:  invoice.Party{Name: "Client Test"},
			IssueDate: time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC),
			Items: []invoice.InvoiceItem{
				{
					Description: "Abonnement SaaS",
					Quantity:    1,
					UnitPrice:   invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR),
					VATRate:     invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR),
				},
			},
		}
	}

	// Test 1: Émission normale
	invA := makeInvoice("FAC-2026-001")
	res1, err := svc.EmitInvoiceTx(ctx, "tenant-A", "idemp-1", invA)
	if err != nil || res1.Status != "SUBMISSION_PENDING" || res1.Cached {
		t.Fatalf("emission 1 échouée: %v, res: %+v", err, res1)
	}

	// Test 2: Même numéro sur un tenant différent (tenant-B) -> Doit réussir grâce au multi-tenant strict
	invB := makeInvoice("FAC-2026-001")
	res2, err := svc.EmitInvoiceTx(ctx, "tenant-B", "idemp-2", invB)
	if err != nil || res2.Status != "SUBMISSION_PENDING" {
		t.Fatalf("tenant-B doit pouvoir émettre son propre FAC-2026-001: %v", err)
	}

	// Test 3: Doublon de numéro au sein du MÊME tenant (tenant-A) -> Rejet immédiat
	invA2 := makeInvoice("FAC-2026-001")
	_, err = svc.EmitInvoiceTx(ctx, "tenant-A", "idemp-3", invA2)
	if err == nil {
		t.Fatalf("l'émission d'un numéro en doublon pour le même tenant doit échouer")
	}

	// Test 4: Rejeu idempotence (même tenant, même clé) -> Renvoie le résultat en cache sans réinsérer
	replayed, err := svc.EmitInvoiceTx(ctx, "tenant-A", "idemp-1", invA)
	if err != nil || !replayed.Cached {
		t.Fatalf("le rejeu doit renvoyer le résultat mis en cache: err=%v, res=%+v", err, replayed)
	}
}
