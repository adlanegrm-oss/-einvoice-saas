package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
)

const defaultPort = "8080"
const APIKey = "secret-api-key-123"
const validAPIKey = APIKey

// apiKeyAuthMiddleware protège les routes API en vérifiant la présence de la clé API
func apiKeyAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey != APIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Accès non autorisé : Clé API manquante ou invalide",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

// loggingMiddleware enregistre les requêtes entrantes
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	mux := http.NewServeMux()

	// Initialisation des handlers
	valHandler := handler.NewValidationHandler()

	// 1. Service des fichiers statiques de l'interface Web (Front/Back office & Swagger Docs)
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)

	// Exposition du contrat d'API Swagger/OpenAPI
	mux.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger.yaml")
	})

	// 2. Health check (Publique)
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "UP",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 3. Routes API Protégées par Clé API
	// Route de validation multi-formats (Factur-X, UBL, EDIFACT, IDDoc, PDF Signé)
	mux.HandleFunc("/api/v1/validate", apiKeyAuthMiddleware(valHandler.HandleValidateDocument))

	// Endpoint Factures
	mux.HandleFunc("/api/v1/invoices", apiKeyAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Endpoint Invoices opérationnel",
		})
	}))

	// Endpoint Rapports
	mux.HandleFunc("/api/v1/reports", apiKeyAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Endpoint Reports opérationnel",
		})
	}))

	// Configuration du serveur HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Serveur démarré sur le port %s (http://localhost:%s)", port, port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Erreur de démarrage : %v", err)
	}
}
