go

mpackage main

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

func TestValidateAndListInvoices(t *testing.T) {
	validateHandler := http.HandlerFunc(ValidateInvoiceHandler)
	listHandler := http.HandlerFunc(ListInvoicesHandler)

	// 1. Valider et enregistrer une facture valide
	validJSON := `{
		"id": "1",
		"number": "INV-001",
		"customer": "Client A",
		"items": [
			{"description": "Prestation", "quantity": 1, "unit_price": 200.0, "vat_rate": 20.0}
		]
	}`
	req1, _ := http.NewRequest("POST", "/invoices/validate", strings.NewReader(validJSON))
	rr1 := httptest.NewRecorder()
	validateHandler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Facture valide : code attendu 200, reçu %v", rr1.Code)
	}

	// 2. Tenter de valider une facture invalide (elle ne doit pas être enregistrée)
	invalidJSON := `{"id":"2", "number":"INV-002", "customer":"Client B", "items": []}`
	req2, _ := http.NewRequest("POST", "/invoices/validate", strings.NewReader(invalidJSON))
	rr2 := httptest.NewRecorder()
	validateHandler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusUnprocessableEntity {
		t.Errorf("Facture invalide : code attendu 422, reçu %v", rr2.Code)
	}

	// 3. Récupérer la liste des factures
	req3, _ := http.NewRequest("GET", "/invoices", nil)
	rr3 := httptest.NewRecorder()
	listHandler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("GET /invoices : code attendu 200, reçu %v", rr3.Code)
	}
}
