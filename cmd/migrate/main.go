package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"einvoice-saas/internal/repository/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL est requis pour exécuter les migrations")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Connexion DB échouée: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Println("Vérification et application des migrations...")
	if err := postgres.ApplyMigrations(ctx, db); err != nil {
		log.Fatalf("Échec des migrations: %v", err)
	}

	log.Println("Base de données synchronisée avec succès.")
}
