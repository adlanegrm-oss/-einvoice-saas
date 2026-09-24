package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
)

func main() {
	// Initialisation de la base SQLite
	db, err := sql.Open("sqlite", "einvoice.db")
	if err != nil {
		log.Fatalf("Erreur d'ouverture de la base de données : %v", err)
	}
	defer db.Close()

	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		log.Fatalf("Erreur d'initialisation du repository SQLite : %v", err)
	}

	// Initialisation du Worker Pool (3 workers en parallèle, file de 100 tâches)
	pool := worker.NewPool(3, 100)
	defer pool.Stop()

	h := handler.NewInvoiceHandler(repo, pool)

	// Routes HTTP
	http.HandleFunc("/health", h.Health)
	http.HandleFunc("/invoices/validate", h.Validate)
	http.HandleFunc("/invoices", h.List)
	http.HandleFunc("/invoices/export/xml", h.ExportXML)

	// Routes Rapports & Traitements Cron
	http.HandleFunc("/reports/daily", h.GetDailyReport)
	http.HandleFunc("/reports/daily/async", h.TriggerAsyncCronTask)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur l'API e-Invoice SaaS avec moteur de traitement asynchrone !")
	})

	log.Println("Serveur démarré sur le port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage : %v", err)
	}
}
