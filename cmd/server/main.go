package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
)

func main() {
	// Initialisation de la base SQLite
	db, err := sql.Open("sqlite", "einvoice.db")
	if err != nil {
		log.Fatalf("Erreur d'ouverture de la base de données : %v", err)
	}
	defer db.Close()

	// Utilisation du repository SQLite
	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		log.Fatalf("Erreur d'initialisation du repository SQLite : %v", err)
	}

	h := handler.NewInvoiceHandler(repo)

	// Routes HTTP
	http.HandleFunc("/health", h.Health)
	http.HandleFunc("/invoices/validate", h.Validate)
	http.HandleFunc("/invoices", h.List)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur l'API e-Invoice SaaS avec persistance SQLite !")
	})

	log.Println("Serveur démarré sur le port 8080 avec base SQLite...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage : %v", err)
	}
}
