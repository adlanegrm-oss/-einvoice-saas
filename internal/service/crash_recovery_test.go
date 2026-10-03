package service_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/connector"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
	_ "modernc.org/sqlite"
)

func TestOutbox_CrashRecoveryScenario(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	schema, err := os.ReadFile("../../migrations/001_initial_schema.sql")
	if err != nil {
		t.Fatalf("lecture schéma : %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}

	tenantID := "tenant-resilience"
	invoiceID := "inv-crash-proof-101"
	invoiceNumber := "FAC-CRASH-001"
	now := time.Now().UTC()

	// 1. Transaction métier atomique simulant POST /emit
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO invoices (id, tenant_id, invoice_number, compliance_status, transmission_status, payment_status, total_cents, currency, payload, created_at, updated_at)
VALUES (?, ?, ?, 'VALID', 'PENDING', 'NOT_DUE', 10000, 'EUR', '<xml>data</xml>', ?, ?)
`, invoiceID, tenantID, invoiceNumber, now, now)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO outbox_events (id, tenant_id, aggregate_id, event_type, payload_json, status, retry_count, max_retries, next_attempt_at, created_at, updated_at)
VALUES ('outbox-evt-1', ?, ?, 'SUBMISSION_REQUESTED', '<xml>data</xml>', 'PENDING', 0, 3, ?, ?, ?)
`, tenantID, invoiceID, now, now, now)
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}

	// COMMIT en base réussi
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// --- 2. SIMULATION CRASH DU SERVEUR / REBOOT ---
	// Aucun worker n'a tourné, l'application s'arrête brutalement ici.
	// La mémoire vive est purgée.

	// --- 3. REDÉMARRAGE DU PROCESSUS & DU WORKER ---
	// Une nouvelle instance démarre et inspecte la base de données persistante
	freshConnector := connector.NewSandboxPDPConnector("PDP-RECOVERY-NODE")
	freshWorker := worker.NewOutboxWorker(db, freshConnector)

	// Le worker dépile les événements PENDING laissés avant le crash
	processed, err := freshWorker.ProcessNextBatch(ctx)
	if err != nil {
		t.Fatalf("Échec lors du redémarrage du worker : %v", err)
	}
	if processed != 1 {
		t.Fatalf("Le worker aurait dû récupérer 1 événement non traité, obtenu %d", processed)
	}

	// 4. Vérification que la facture a bien été transmise et acceptée après reprise
	var finalStatus string
	err = db.QueryRowContext(ctx, "SELECT transmission_status FROM invoices WHERE id = ?", invoiceID).Scan(&finalStatus)
	if err != nil || finalStatus != "ACCEPTED" {
		t.Fatalf("Après crash et redémarrage, la facture aurait dû être 'ACCEPTED', statut actuel: %s", finalStatus)
	}

	t.Log("✓ Succès : Preuve de résilience au crash validée, aucune facture perdue.")
}
