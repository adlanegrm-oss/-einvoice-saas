package retention

import (
"context"
"database/sql"
"fmt"
"log/slog"
"os"
"path/filepath"
"time"
)

type Policy struct {
InstanceID      string
StorageRoot     string
RetentionMonths int
PurgeHourUTC    int
}

type Service struct {
db     *sql.DB
policy Policy
}

func NewService(db *sql.DB, policy Policy) *Service {
return &Service{db: db, policy: policy}
}

func ComputeCutoffDate(targetDate time.Time, months int) time.Time {
return time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC).
AddDate(0, -months, 0)
}

func NextPurgeDelay(now time.Time, targetHourUTC int) time.Duration {
next := time.Date(now.Year(), now.Month(), now.Day(), targetHourUTC, 0, 0, 0, time.UTC)
if !now.Before(next) {
next = next.AddDate(0, 0, 1)
}
return next.Sub(now)
}

func (s *Service) PurgeCycle(ctx context.Context, period string) error {
jobID := fmt.Sprintf("purge_%s_%s_%d", s.policy.InstanceID, period, time.Now().UnixNano())

var existingStatus string
err := s.db.QueryRowContext(ctx, "SELECT status FROM retention_purge_jobs WHERE accounting_period = ?", period).
Scan(&existingStatus)

if err == nil {
if existingStatus == "COMPLETED" {
slog.Info("période déjà purgée, opération ignorée", "period", period)
return nil
}
} else if err != sql.ErrNoRows {
return fmt.Errorf("contrôle statut purge: %w", err)
}

periodPath := filepath.Join(s.policy.StorageRoot, "HISTORY", period)

if existingStatus == "DB_PURGED" {
slog.Warn("reprise de purge : DB déjà nettoyée, tentative filesystem", "period", period)
return s.purgeFilesystemAndComplete(ctx, period, periodPath, 0)
}

var unclosedCount int
err = s.db.QueryRowContext(ctx, `
SELECT COUNT(1) FROM invoices 
WHERE accounting_period = ? AND status != 'CLOSED'`, period).Scan(&unclosedCount)
if err != nil {
return fmt.Errorf("vérification clôture documents: %w", err)
}
if unclosedCount > 0 {
return fmt.Errorf("purge impossible pour %s : %d documents non clôturés", period, unclosedCount)
}

_, err = s.db.ExecContext(ctx, `
INSERT INTO retention_purge_jobs (id, instance_id, accounting_period, status, started_at)
VALUES (?, ?, ?, 'IN_PROGRESS', ?)
ON CONFLICT(accounting_period) DO UPDATE SET status='IN_PROGRESS', error_message=NULL`,
jobID, s.policy.InstanceID, period, time.Now().UTC())
if err != nil {
return fmt.Errorf("initialisation job: %w", err)
}

tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
return s.failJob(ctx, period, fmt.Errorf("init transaction: %w", err))
}

_, err = tx.ExecContext(ctx, "DELETE FROM invoice_events WHERE invoice_id IN (SELECT id FROM invoices WHERE accounting_period = ?)", period)
if err != nil {
_ = tx.Rollback()
return s.failJob(ctx, period, fmt.Errorf("purge events DB: %w", err))
}

res, err := tx.ExecContext(ctx, "DELETE FROM invoices WHERE accounting_period = ?", period)
if err != nil {
_ = tx.Rollback()
return s.failJob(ctx, period, fmt.Errorf("purge invoices DB: %w", err))
}

rowsDeleted, err := res.RowsAffected()
if err != nil {
_ = tx.Rollback()
return s.failJob(ctx, period, fmt.Errorf("lecture rows affected: %w", err))
}

if err := tx.Commit(); err != nil {
return s.failJob(ctx, period, fmt.Errorf("commit purge DB: %w", err))
}

if _, err = s.db.ExecContext(ctx, "UPDATE retention_purge_jobs SET status = 'DB_PURGED', rows_deleted = ? WHERE accounting_period = ?", rowsDeleted, period); err != nil {
slog.Error("échec mise à jour statut DB_PURGED", "err", err, "period", period)
}

return s.purgeFilesystemAndComplete(ctx, period, periodPath, rowsDeleted)
}

func (s *Service) purgeFilesystemAndComplete(ctx context.Context, period, periodPath string, rowsDeleted int64) error {
if _, err := os.Stat(periodPath); err == nil {
if err := os.RemoveAll(periodPath); err != nil {
return s.failJob(ctx, period, fmt.Errorf("purge filesystem (%s): %w", periodPath, err))
}
}

_, err := s.db.ExecContext(ctx, `
UPDATE retention_purge_jobs 
SET status = 'COMPLETED', completed_at = ?, error_message = NULL 
WHERE accounting_period = ?`, time.Now().UTC(), period)
if err != nil {
return fmt.Errorf("finalisation job: %w", err)
}

slog.Info("purge cycle achevée avec succès", "period", period, "rows_db", rowsDeleted)
return nil
}

func (s *Service) failJob(ctx context.Context, period string, err error) error {
if _, dbErr := s.db.ExecContext(ctx, `
UPDATE retention_purge_jobs 
SET status = 'FAILED', error_message = ? 
WHERE accounting_period = ?`, err.Error(), period); dbErr != nil {
slog.Error("échec enregistrement état FAILED", "dbErr", dbErr, "period", period)
}
return err
}

func (s *Service) RunScheduler(ctx context.Context) {
for {
delay := NextPurgeDelay(time.Now().UTC(), s.policy.PurgeHourUTC)
timer := time.NewTimer(delay)
select {
case <-ctx.Done():
timer.Stop()
return
case now := <-timer.C:
if _, err := s.RunBatch(ctx, now); err != nil {
slog.Error("échec exécution scheduled batch", "err", err)
}
}
}
}

// RunBatch découvre et purge l'ensemble des périodes HISTORY antérieures au cutoff.
func (s *Service) RunBatch(ctx context.Context, now time.Time) ([]string, error) {
cutoff := ComputeCutoffDate(now, s.policy.RetentionMonths)
historyDir := filepath.Join(s.policy.StorageRoot, "HISTORY")
entries, err := os.ReadDir(historyDir)
if err != nil {
if os.IsNotExist(err) {
return nil, nil
}
return nil, fmt.Errorf("lecture HISTORY: %w", err)
}

var purged []string
for _, e := range entries {
if !e.IsDir() {
continue
}
pDate, err := time.Parse("2006-01", e.Name())
if err != nil {
continue
}
if pDate.Before(cutoff) {
if err := s.PurgeCycle(ctx, e.Name()); err != nil {
slog.Error("échec purge cycle", "period", e.Name(), "err", err)
return purged, err
}
purged = append(purged, e.Name())
}
}
return purged, nil
}
