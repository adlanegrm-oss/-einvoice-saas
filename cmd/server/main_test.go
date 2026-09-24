package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Impossible de créer la requête : %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)

	handler.ServeHTTP(rr, req)

	// Vérification du code HTTP 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Statut incorrect : reçu %v, attendu %v", status, http.StatusOK)
	}

	// Vérification du contenu JSON
	expected := `{"status": "ok"}`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("Réponse incorrecte : reçu %v, attendu %v", rr.Body.String(), expected)
	}
}
