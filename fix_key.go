package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://postgres:@127.0.0.1:5432/einvoice_saas?sslmode=disable"
    }
    db, err := sql.Open("pgx", dbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    _, err = db.Exec(`INSERT INTO api_keys (key, name, is_active, created_at) VALUES ('test_api_key_123456789012345678901234567890', 'Test Key', true, NOW()) ON CONFLICT DO NOTHING`)
    if err != nil {
        _, err = db.Exec(`INSERT INTO api_keys (key_hash, name, is_active, created_at) VALUES ('test_api_key_123456789012345678901234567890', 'Test Key', true, NOW()) ON CONFLICT DO NOTHING`)
        if err != nil {
            log.Fatal("Erreur insertion clé : ", err)
        }
    }
    fmt.Println("Clé API de test insérée avec succès.")
}