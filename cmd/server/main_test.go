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

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Statut incorrect : reçu %v, attendu %v", status, http.StatusOK)
	}
}

func TestValidateInvoiceHandler(t *testing.T) {
	handler := http.HandlerFunc(ValidateInvoiceHandler)

	// Cas 1 : Facture valide (Doit retourner 200 OK)
	validJSON := `{"id":"1", "number":"INV-001", "customer":"Client A", "amount":150.50, "currency":"EUR"}`
	req1, _ := http.NewRequest("POST", "/invoices/validate", strings.NewReader(validJSON))
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Facture valide : code attendu 200, reçu %v", rr1.Code)
	}

	// Cas 2 : Facture invalide (Montant <= 0 -> Doit retourner 422 Unprocessable Entity)
	invalidJSON := `{"id":"2", "number":"INV-002", "customer":"Client B", "amount":0}`
	req2, _ := http.NewRequest("POST", "/invoices/validate", strings.NewReader(invalidJSON))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusUnprocessableEntity {
		t.Errorf("Facture invalide : code attendu 422, reçu %v", rr2.Code)
	}
}
