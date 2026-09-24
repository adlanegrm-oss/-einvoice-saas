package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
)

const APIKey = "secret-api-key-123"

func main() {
	db, err := sql.Open("sqlite", "einvoice.db")
	if err != nil {
		log.Fatalf("Erreur d'ouverture de la base : %v", err)
	}
	defer db.Close()

	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		log.Fatalf("Erreur d'initialisation SQLite : %v", err)
	}

	pool := worker.NewPool(3, 100)
	defer pool.Stop()

	h := handler.NewInvoiceHandler(repo, pool)

	// Servir le Front-Office / Back-Office Web
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// Route publique de santé
	http.HandleFunc("/health", h.Health)

	// Routes sécurisées par Middleware API Key
	http.HandleFunc("/invoices/validate", middleware.AuthMiddleware(h.Validate, APIKey))
	http.HandleFunc("/invoices", middleware.AuthMiddleware(h.List, APIKey))
	http.HandleFunc("/invoices/export/xml", middleware.AuthMiddleware(h.ExportXML, APIKey))
	http.HandleFunc("/reports/daily", middleware.AuthMiddleware(h.GetDailyReport, APIKey))
	http.HandleFunc("/reports/daily/async", middleware.AuthMiddleware(h.TriggerAsyncCronTask, APIKey))

	log.Println("Serveur web démarré sur http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage : %v", err)
	}
}
