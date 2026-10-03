package retention_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/retention"
	_ "modernc.org/sqlite"
)

func TestRetentionCutoff(t *testing.T) {
	refDate := time.Date(2026, 10, 3, 2, 0, 0, 0, time.UTC)
	cutoff := retention.ComputeCutoffDate(refDate, 3)

	expected := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if !cutoff.Equal(expected) {
		t.Fatalf("attendu %v, obtenu %v", expected, cutoff)
	}

	pJune, _ := time.Parse("2006-01", "2026-06")
	if !pJune.Before(cutoff) {
		t.Errorf("2026-06 devrait être éligible à la purge")
	}

	pJuly, _ := time.Parse("2006-01", "2026-07")
	if pJuly.Before(cutoff) {
		t.Errorf("2026-07 ne devrait pas être purgé")
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
CREATE TABLE invoices (id TEXT PRIMARY KEY, accounting_period TEXT, status TEXT);
CREATE TABLE invoice_events (invoice_id TEXT);
CREATE TABLE retention_purge_jobs (
id TEXT PRIMARY KEY, instance_id TEXT, accounting_period TEXT UNIQUE,
status TEXT, started_at DATETIME, completed_at DATETIME,
files_deleted INTEGER DEFAULT 0, rows_deleted INTEGER DEFAULT 0, error_message TEXT
);
`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPurgeCycle_UnclosedGuard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	_, _ = db.Exec("INSERT INTO invoices VALUES ('INV-001', '2026-06', 'PROCESSING')")

	policy := retention.Policy{
		InstanceID:      "GLOBAL-A",
		StorageRoot:     t.TempDir(),
		RetentionMonths: 3,
	}

	svc := retention.NewService(db, policy)
	err := svc.PurgeCycle(ctx, "2026-06")
	if err == nil {
		t.Fatal("la purge aurait dû être rejetée en raison de documents non CLOSED")
	}
}

func TestPurgeCycle_ExecutionSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	_, _ = db.Exec("INSERT INTO invoices VALUES ('INV-002', '2026-06', 'CLOSED')")

	tmpDir := t.TempDir()
	periodDir := filepath.Join(tmpDir, "HISTORY", "2026-06")
	_ = os.MkdirAll(periodDir, 0755)
	_ = os.WriteFile(filepath.Join(periodDir, "invoice.xml"), []byte("<data/>"), 0644)

	policy := retention.Policy{
		InstanceID:      "GLOBAL-A",
		StorageRoot:     tmpDir,
		RetentionMonths: 3,
	}

	svc := retention.NewService(db, policy)
	if err := svc.PurgeCycle(ctx, "2026-06"); err != nil {
		t.Fatalf("échec de purge: %v", err)
	}

	if _, err := os.Stat(periodDir); !os.IsNotExist(err) {
		t.Errorf("le répertoire physique %s aurait dû être supprimé", periodDir)
	}

	var status string
	_ = db.QueryRow("SELECT status FROM retention_purge_jobs WHERE accounting_period = '2026-06'").Scan(&status)
	if status != "COMPLETED" {
		t.Errorf("statut attendu: COMPLETED, obtenu: %s", status)
	}
}

func TestPurgeCycle_Idempotence(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	_, _ = db.Exec(`INSERT INTO retention_purge_jobs (id, instance_id, accounting_period, status, started_at) 
VALUES ('p-done', 'GLOBAL-A', '2026-05', 'COMPLETED', CURRENT_TIMESTAMP)`)

	policy := retention.Policy{
		InstanceID:      "GLOBAL-A",
		StorageRoot:     t.TempDir(),
		RetentionMonths: 3,
	}

	svc := retention.NewService(db, policy)
	if err := svc.PurgeCycle(ctx, "2026-05"); err != nil {
		t.Fatalf("un job déjà complété doit retourner nil sans erreur: %v", err)
	}
}

func TestPurgeCycle_CrashRecoveryDBPurged(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tmpDir := t.TempDir()
	periodDir := filepath.Join(tmpDir, "HISTORY", "2026-05")
	_ = os.MkdirAll(periodDir, 0755)
	_ = os.WriteFile(filepath.Join(periodDir, "residual.xml"), []byte("<orphan/>"), 0644)

	// Simuler un crash antérieur : la DB a été purgée mais le dossier est resté
	_, _ = db.Exec(`INSERT INTO retention_purge_jobs (id, instance_id, accounting_period, status, started_at) 
VALUES ('p-crashed', 'GLOBAL-A', '2026-05', 'DB_PURGED', CURRENT_TIMESTAMP)`)

	policy := retention.Policy{
		InstanceID:      "GLOBAL-A",
		StorageRoot:     tmpDir,
		RetentionMonths: 3,
	}

	svc := retention.NewService(db, policy)
	if err := svc.PurgeCycle(ctx, "2026-05"); err != nil {
		t.Fatalf("la reprise sur statut DB_PURGED a échoué: %v", err)
	}

	// Vérifier que le dossier résiduel a bien été supprimé
	if _, err := os.Stat(periodDir); !os.IsNotExist(err) {
		t.Errorf("le dossier résiduel %s aurait dû être supprimé lors de la reprise", periodDir)
	}

	// Vérifier que le statut est passé à COMPLETED
	var status string
	_ = db.QueryRow("SELECT status FROM retention_purge_jobs WHERE accounting_period = '2026-05'").Scan(&status)
	if status != "COMPLETED" {
		t.Errorf("statut final attendu: COMPLETED, obtenu: %s", status)
	}
}
