package postgres

import (
"context"
"database/sql"
"embed"
"fmt"
"io/fs"
"path"
"sort"
"strings"
"time"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    filename TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// ApplyMigrations applique de façon idempotente les scripts SQL non encore enregistrés
func ApplyMigrations(ctx context.Context, db *sql.DB) error {
if _, err := db.ExecContext(ctx, createMigrationsTable); err != nil {
return fmt.Errorf("création schema_migrations: %w", err)
}

rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
if err != nil {
return fmt.Errorf("lecture schema_migrations: %w", err)
}
defer rows.Close()

applied := make(map[string]bool)
for rows.Next() {
var v string
if err := rows.Scan(&v); err != nil {
return err
}
applied[v] = true
}
if err := rows.Err(); err != nil {
return err
}

entries, err := fs.ReadDir(migrationFS, "migrations")
if err != nil {
return fmt.Errorf("lecture migrationFS: %w", err)
}

var upFiles []string
for _, entry := range entries {
if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
upFiles = append(upFiles, entry.Name())
}
}
sort.Strings(upFiles)

for _, filename := range upFiles {
version := strings.Split(filename, "_")[0]
if applied[version] {
continue
}

rawContent, err := migrationFS.ReadFile(path.Join("migrations", filename))
if err != nil {
return fmt.Errorf("lecture %s: %w", filename, err)
}

sqlContent := strings.TrimPrefix(string(rawContent), "\ufeff")

if err := applySingleMigration(ctx, db, version, filename, sqlContent); err != nil {
return fmt.Errorf("échec migration %s: %w", filename, err)
}
}

return nil
}

func applySingleMigration(ctx context.Context, db *sql.DB, version, filename, sqlScript string) error {
tx, err := db.BeginTx(ctx, nil)
if err != nil {
return err
}
defer tx.Rollback()

if _, err := tx.ExecContext(ctx, sqlScript); err != nil {
return err
}

query := `INSERT INTO schema_migrations (version, filename, applied_at) VALUES ($1, $2, $3)`
if _, err := tx.ExecContext(ctx, query, version, filename, time.Now().UTC()); err != nil {
return err
}

return tx.Commit()
}
