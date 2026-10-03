package service_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/connector"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
	_ "modernc.org/sqlite"
)

func TestOutbox_RealDiskCrashRecovery(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "einvoice-crash-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dbFile := filepath.Join(tempDir, "durable_outbox.db")
	dsn := "file:" + dbFile + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}

	schema, err := os.ReadFile("../../migrations/001_initial_schema.sql")
	if err != nil {
		db.Close()
		t.Fatalf("lecture schéma : %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		t.Fatal(err)
	}

	tenantID := "tenant-disk-proof"
	invoiceID := "inv-disk-001"
	invoiceNumber := "FAC-DISK-001"
	now := time.Now().UTC()

	// Ingestion transactionnelle atomique
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	_, _ = tx.ExecContext(ctx, `
INSERT INTO invoices (id, tenant_id, invoice_number, compliance_status, transmission_status, payment_status, total_cents, currency, payload, created_at, updated_at)
VALUES (?, ?, ?, 'VALID', 'PENDING', 'NOT_DUE', 50000, 'EUR', '<xml>durable</xml>', ?, ?)
`, invoiceID, tenantID, invoiceNumber, now, now)

	_, _ = tx.ExecContext(ctx, `
INSERT INTO outbox_events (id, tenant_id, aggregate_id, event_type, payload_json, status, retry_count, max_retries, next_attempt_at, created_at, updated_at)
VALUES ('outbox-disk-evt-1', ?, ?, 'SUBMISSION_REQUESTED', '<xml>durable</xml>', 'PENDING', 0, 3, ?, ?, ?)
`, tenantID, invoiceID, now, now, now)

	if err := tx.Commit(); err != nil {
		db.Close()
		t.Fatal(err)
	}

	// SIMULATION DU CRASH : Fermeture brutale du pool de connexions (mort du processus)
	db.Close()

	// REDÉMARRAGE SUR NOUVEAU PROCESSUS : Réouverture physique de la BDD disque
	recoveredDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("reconnexion impossible post-crash: %v", err)
	}
	defer recoveredDB.Close()

	connectorMock := connector.NewSandboxPDPConnector("PDP-DISK-NODE")
	freshWorker := worker.NewOutboxWorker(recoveredDB, connectorMock)

	processed, err := freshWorker.ProcessNextBatch(ctx)
	if err != nil || processed != 1 {
		t.Fatalf("Reprise post-crash échouée : err=%v, processed=%d", err, processed)
	}

	var status string
	err = recoveredDB.QueryRowContext(ctx, "SELECT transmission_status FROM invoices WHERE id = ?", invoiceID).Scan(&status)
	if err != nil || status != "ACCEPTED" {
		t.Fatalf("Facture non transmise après reprise sur disque: statut=%s, err=%v", status, err)
	}
}
