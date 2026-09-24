package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// AuthMiddleware vérifie la présence et la validité de la clé d'API
func AuthMiddleware(next http.HandlerFunc, validAPIKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		apiKeyHeader := r.Header.Get("X-API-Key")

		token := ""
		if apiKeyHeader != "" {
			token = apiKeyHeader
		} else if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Si aucune clé valide n'est fournie
		if token == "" || token != validAPIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Accès non autorisé : Clé API manquante ou invalide",
			})
			return
		}

		next(w, r)
	}
}
