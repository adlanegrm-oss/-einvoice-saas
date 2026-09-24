package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
	_ "modernc.org/sqlite"
)

func setupTestHandler(t *testing.T) (*handler.InvoiceHandler, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Impossible d'ouvrir la base : %v", err)
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

	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(h.Health).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Statut incorrect : reçu %v, attendu %v", status, http.StatusOK)
	}
}

func TestProtectedEndpoints(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	protectedValidate := middleware.AuthMiddleware(h.Validate, APIKey)

	// 1. Rejet sans clé
	req1, _ := http.NewRequest("POST", "/invoices/validate", strings.NewReader(`{}`))
	rr1 := httptest.NewRecorder()
	protectedValidate.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusUnauthorized {
		t.Errorf("Attendu 401 Unauthorized, reçu %d", rr1.Code)
	}

	// 2. Succès avec clé API
	validJSON := `{
		"id": "1",
		"number": "INV-001",
		"customer": "Client A",
		"items": [
			{"description": "Prestation", "quantity": 1, "unit_price": 200.0, "vat_rate": 20.0}
		]
	}`
	req2, _ := http.NewRequest("POST", "/invoices/validate", strings.NewReader(validJSON))
	req2.Header.Set("X-API-Key", APIKey)
	rr2 := httptest.NewRecorder()
	protectedValidate.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Facture valide : code attendu 200, reçu %v", rr2.Code)
	}
}
