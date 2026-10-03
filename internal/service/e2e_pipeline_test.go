package service_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
	"github.com/adlanegrm-oss/einvoice-saas/internal/connector"
	"github.com/adlanegrm-oss/einvoice-saas/internal/evidence"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	schema, err := os.ReadFile("../../migrations/001_initial_schema.sql")
	if err != nil {
		t.Fatalf("lecture schema: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("exec migration: %v", err)
	}
	return db
}

func TestEndToEndPipeline_Validation_Seal_Outbox_Worker_Evidence(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	tenantID := "tenant-alpha"
	invoiceID := "inv-e2e-100"
	invoiceNumber := "FAC-2026-999"
	sellerSIRET := "12345678901234"
	buyerSIRET := "98765432109876"
	netCents := int64(10000)
	taxCents := int64(2000)
	grossCents := int64(12000)

	// ETAPE 1 : VALIDATION REGLEMENTAIRE
	engine := en16931.NewComplianceEngine("2026.1")
	report := engine.Validate(invoiceNumber, sellerSIRET, buyerSIRET, netCents, taxCents, grossCents, 1)
	if !report.Valid {
		t.Fatalf("Validation échouée: %+v", report.Diagnostics)
	}

	// ETAPE 2 : TRANSACTION ATOMIQUE
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
INSERT INTO invoices (id, tenant_id, invoice_number, compliance_status, transmission_status, payment_status, total_cents, currency, payload, created_at, updated_at)
VALUES (?, ?, ?, 'VALID', 'PENDING', 'NOT_DUE', ?, 'EUR', '<xml>payload</xml>', ?, ?)
`, invoiceID, tenantID, invoiceNumber, grossCents, now, now)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	genesisPrev := "0000000000000000000000000000000000000000000000000000000000000000"
	payloadHash := evidence.ComputeEventHash(genesisPrev, "<xml>payload</xml>", "INVOICE_SEALED", now)
	_, err = tx.ExecContext(ctx, `
INSERT INTO audit_events (id, tenant_id, invoice_id, sequence_id, event_type, payload_hash, prev_hash, event_hash, recorded_at)
VALUES (?, ?, ?, 1, 'INVOICE_SEALED', ?, ?, ?, ?)
`, "evt-1", tenantID, invoiceID, payloadHash, genesisPrev, payloadHash, now)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO outbox_events (id, tenant_id, aggregate_id, event_type, payload_json, status, retry_count, max_retries, next_attempt_at, created_at, updated_at)
VALUES (?, ?, ?, 'SUBMISSION_REQUESTED', '<xml>payload</xml>', 'PENDING', 0, 3, ?, ?, ?)
`, "outbox-1", tenantID, invoiceID, now, now, now)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// ETAPE 3 : WORKER -> CONNECTEUR SANDBOX
	connectorMock := connector.NewSandboxPDPConnector("PDP-FR-DEMO")
	outboxWorker := worker.NewOutboxWorker(db, connectorMock)

	processed, err := outboxWorker.ProcessNextBatch(ctx)
	if err != nil || processed != 1 {
		t.Fatalf("worker error: %v, processed: %d", err, processed)
	}

	// ETAPE 4 : STATUT DE TRANSMISSION
	var transStatus string
	err = db.QueryRowContext(ctx, "SELECT transmission_status FROM invoices WHERE id = ?", invoiceID).Scan(&transStatus)
	if err != nil || transStatus != "ACCEPTED" {
		t.Fatalf("Statut attendu ACCEPTED, obtenu: %s", transStatus)
	}

	// ETAPE 5 : EVIDENCE GENERATION
	rows, err := db.QueryContext(ctx, "SELECT id, sequence_id, event_type, payload_hash, prev_hash, event_hash, recorded_at FROM audit_events WHERE invoice_id = ? ORDER BY sequence_id ASC", invoiceID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var chain []evidence.AuditEvent
	for rows.Next() {
		var e evidence.AuditEvent
		if err := rows.Scan(&e.ID, &e.SequenceID, &e.EventType, &e.PayloadHash, &e.PrevHash, &e.EventHash, &e.RecordedAt); err != nil {
			t.Fatal(err)
		}
		chain = append(chain, e)
	}

	if len(chain) < 2 {
		t.Fatalf("Attendu au moins 2 preuves dans la chaine, obtenu %d", len(chain))
	}
	t.Logf("Succes : Scénario E2E validé avec %d événements scellés", len(chain))
}
