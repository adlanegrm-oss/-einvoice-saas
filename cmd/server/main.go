package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
)

func main() {
	// Initialisation de la couche données et HTTP
	repo := repository.NewMemoryInvoiceRepository()
	h := handler.NewInvoiceHandler(repo)

	// Déclaration des routes
	http.HandleFunc("/health", h.Health)
	http.HandleFunc("/invoices/validate", h.Validate)
	http.HandleFunc("/invoices", h.List)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur l'API e-Invoice SaaS !")
	})

	log.Println("Serveur démarré sur le port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage : %v", err)
	}
}
