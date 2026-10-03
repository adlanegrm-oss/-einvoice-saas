package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "./data/einvoice.db"
	if envDb := os.Getenv("DATABASE_PATH"); envDb != "" {
		dbPath = envDb
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Erreur ouverture DB : %v\n", err)
		return
	}
	defer db.Close()

	// Initialiser la table si absente
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	email := "admin@mondomaine.com"
	pass := "MonSuperPasswordDeTest2026!"

	hash, err := auth.HashPassword(pass)
	if err != nil {
		fmt.Printf("Erreur hash : %v\n", err)
		return
	}

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash, role) 
		VALUES (?, ?, ?, 'admin') 
		ON CONFLICT(email) DO UPDATE SET password_hash = excluded.password_hash`,
		"admin-01", email, hash)

	if err != nil {
		fmt.Printf("Erreur SQL : %v\n", err)
		return
	}

	fmt.Printf("[OK] Compte administrateur synchronisé :\n  Email : %s\n  Pass  : %s\n", email, pass)
}