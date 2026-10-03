package idempotency

import (
"context"
"database/sql"
"testing"
"time"

_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
t.Helper()
db, err := sql.Open("sqlite", ":memory:")
if err != nil {
t.Fatalf("ouverture sqlite: %v", err)
}
t.Cleanup(func() { db.Close() })

schema := `
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
);`
if _, err := db.Exec(schema); err != nil {
t.Fatalf("init schema: %v", err)
}
return db
}

func TestIdempotencyStore_TenantIsolation(t *testing.T) {
db := setupTestDB(t)
store := NewStore(db)
ctx := context.Background()

key := "req-abc-001"
hash := ComputeHash([]byte(`{"number":"FAC-1"}`))

// 1. Tenant 1 pose un verrou
tx1, _ := db.BeginTx(ctx, nil)
rec, err := store.LockKey(ctx, tx1, "tenant-1", key, hash, time.Hour)
if err != nil || rec != nil {
t.Fatalf("lock tenant-1 échoué: %v", err)
}
_ = store.Complete(ctx, tx1, "tenant-1", key, 201, `{"id":"inv-1"}`)
_ = tx1.Commit()

// 2. Tenant 2 utilise la MEME clé avec un contenu différent -> Doit passer car tenant séparé
hash2 := ComputeHash([]byte(`{"number":"FAC-2"}`))
tx2, _ := db.BeginTx(ctx, nil)
rec2, err := store.LockKey(ctx, tx2, "tenant-2", key, hash2, time.Hour)
if err != nil || rec2 != nil {
t.Fatalf("lock tenant-2 avec même clé doit réussir grâce à l'isolation: %v", err)
}
_ = tx2.Rollback()

// 3. Tenant 1 rejoue la même requête -> Doit recevoir la réponse en cache
tx3, _ := db.BeginTx(ctx, nil)
cached, err := store.LockKey(ctx, tx3, "tenant-1", key, hash, time.Hour)
if err != nil {
t.Fatalf("rejeu tenant-1 échoué: %v", err)
}
if cached == nil || cached.ResponseCode != 201 || cached.ResponseBody != `{"id":"inv-1"}` {
t.Fatalf("réponse en cache invalide: %+v", cached)
}
_ = tx3.Rollback()
}
