package main

import (
	"fmt"
	"log"
	"net/http"
)

// HealthHandler gère la vérification d'état de l'API
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "ok"}`)
}

func main() {
	http.HandleFunc("/health", HealthHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bienvenue sur l'API e-Invoice SaaS !")
	})

	log.Println("Serveur démarré sur le port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage : %v", err)
	}
}
