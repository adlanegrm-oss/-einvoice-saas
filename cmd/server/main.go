package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Route de vérification de santé (Healthcheck)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
	})

	// Route d'accueil
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur l'API e-Invoice SaaS !")
	})

	log.Println("Serveur démarré sur le port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage : %v", err)
	}
}
