package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
	_ "modernc.org/sqlite"
)

func setupTestHandler(t *testing.T) (*handler.InvoiceHandler, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Impossible d'ouvrir la base en mémoire : %v", err)
	}

	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		db.Close()
		t.Fatalf("Impossible d'initialiser le repository : %v", err)
	}

	pool := worker.NewPool(1, 10)

	h := handler.NewInvoiceHandler(repo, pool)

	cleanup := func() {
		pool.Stop()
		db.Close()
	}

	return h, cleanup
}

func TestHealthHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Impossible de créer la requête : %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(h.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Statut incorrect : reçu %v, attendu %v", status, http.StatusOK)
	}
}

func TestValidateAndListInvoices(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

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
	http.HandlerFunc(h.Validate).ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Facture valide : code attendu 200, reçu %v", rr1.Code)
	}

	// 2. Tenter de valider une facture invalide
	invalidJSON := `{"id":"2", "number":"INV-002", "customer":"Client B", "items": []}`
	req2, _ := http.NewRequest("POST", "/invoices/validate", strings.NewReader(invalidJSON))
	rr2 := httptest.NewRecorder()
	http.HandlerFunc(h.Validate).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusUnprocessableEntity {
		t.Errorf("Facture invalide : code attendu 422, reçu %v", rr2.Code)
	}

	// 3. Récupérer la liste des factures
	req3, _ := http.NewRequest("GET", "/invoices", nil)
	rr3 := httptest.NewRecorder()
	http.HandlerFunc(h.List).ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("GET /invoices : code attendu 200, reçu %v", rr3.Code)
	}
}
