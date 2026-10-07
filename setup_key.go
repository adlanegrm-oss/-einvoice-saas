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

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS api_keys (
            id SERIAL PRIMARY KEY,
            key VARCHAR(255) UNIQUE,
            key_hash VARCHAR(255),
            name VARCHAR(255),
            is_active BOOLEAN DEFAULT TRUE,
            created_at TIMESTAMP DEFAULT NOW()
        );
    `)
    if err != nil {
        log.Fatal("Erreur création table api_keys : ", err)
    }

    _, err = db.Exec(`
        INSERT INTO api_keys (key, name, is_active, created_at) 
        VALUES ('test_api_key_123456789012345678901234567890', 'Test Key', true, NOW()) 
        ON CONFLICT (key) DO UPDATE SET is_active = TRUE;
    `)
    if err != nil {
        log.Fatal("Erreur insertion clé : ", err)
    }

    fmt.Println("Table api_keys créée et clé insérée avec succès.")
}