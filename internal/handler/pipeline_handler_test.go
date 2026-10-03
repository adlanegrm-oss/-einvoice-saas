package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
	"github.com/adlanegrm-oss/einvoice-saas/internal/service"
	_ "modernc.org/sqlite"
)

type mockDirectory struct{}

func (m *mockDirectory) Lookup(ctx context.Context, id string) (*dispatcher.TargetEndpoint, error) {
	return &dispatcher.TargetEndpoint{
		ReceiverID:   id,
		PlatformName: "PDP Partenaire",
		AS4Endpoint:  "https://as4.test",
	}, nil
}

type mockAS4 struct{}

func (m *mockAS4) SendPayload(ctx context.Context, ep *dispatcher.TargetEndpoint, p []byte) (string, error) {
	return "RECEIPT-AS4-999", nil
}

func setupTestServer(t *testing.T) (*PipelineHandler, *repository.SQLiteInvoiceRepository) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Erreur ouverture DB mémoire : %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		t.Fatalf("Erreur initialisation repo : %v", err)
	}

	val := en16931.NewValidator()
	sm := status.NewStateMachine()
	disp := dispatcher.NewDispatcher(&mockDirectory{}, &mockAS4{})
	pipeline := service.NewInvoicePipeline(repo, val, sm, disp)

	return NewPipelineHandler(pipeline, repo), repo
}

func TestPipelineHandler_EmitInvoice_SuccessAndAudit(t *testing.T) {
	handler, _ := setupTestServer(t)

	inv := invoice.Invoice{
		ID:        "INV-HTTP-001",
		Number:    "FA-2026-0001",
		Currency:  "EUR",
		IssueDate: time.Now(),
		Seller: invoice.Party{
			Name:  "Fournisseur SAS",
			SIRET: "12345678901234",
			Address: invoice.PostalAddress{
				StreetName:  "1 Rue de Paris",
				PostalZone:  "75001",
				CityName:    "Paris",
				CountryCode: "FR",
			},
		},
		Customer: invoice.Party{
			Name:  "Client SAS",
			SIRET: "98765432109876",
		},
		Items: []invoice.InvoiceItem{
			{Description: "Service Cloud", Quantity: 1, UnitPrice: invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR), VATRate: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR)},
		},
		TotalHT: invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR),
		TotalVAT: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR),
		TotalTTC: invoice.NewMoneyFromFloat(120.0, 2, invoice.CurrencyEUR),
	}

	body, _ := json.Marshal(inv)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/emit", bytes.NewReader(body))
	req.Header.Set("X-Tenant-ID", "tenant-123")
	rec := httptest.NewRecorder()

	handler.EmitInvoice(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Statut HTTP 200 attendu, obtenu : %d. Réponse : %s", rec.Code, rec.Body.String())
	}

	var res service.IngestionResult
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Status != status.StateTransmitted {
		t.Errorf("Statut attendu %s, obtenu : %s", status.StateTransmitted, res.Status)
	}

	// Consultation de la piste d'audit
	auditReq := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/INV-HTTP-001/audit-trail", nil)
	auditRec := httptest.NewRecorder()

	handler.GetAuditTrail(auditRec, auditReq)

	if auditRec.Code != http.StatusOK {
		t.Fatalf("Statut HTTP 200 attendu pour l'audit, obtenu : %d", auditRec.Code)
	}

	var history []status.StatusEvent
	_ = json.Unmarshal(auditRec.Body.Bytes(), &history)
	if len(history) != 3 {
		t.Errorf("Attendu 3 événements d'audit, obtenu : %d", len(history))
	}
}

func TestPipelineHandler_EmitInvoice_ValidationError(t *testing.T) {
	handler, _ := setupTestServer(t)

	// Facture non conforme (pas de devise, calcul TTC faux)
	inv := invoice.Invoice{
		ID:        "INV-HTTP-BAD",
		Number:    "FA-2026-BAD",
		IssueDate: time.Now(),
		Seller: invoice.Party{
			Name:  "Fournisseur SAS",
			SIRET: "12345678901234",
		},
		Customer: invoice.Party{Name: "Client SAS"},
		Items: []invoice.InvoiceItem{
			{Description: "Service Cloud", Quantity: 1, UnitPrice: invoice.NewMoneyFromFloat(50.0, 2, invoice.CurrencyEUR), VATRate: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR)},
		},
		TotalHT: invoice.NewMoneyFromFloat(50.0, 2, invoice.CurrencyEUR),
		TotalVAT: invoice.NewMoneyFromFloat(10.0, 2, invoice.CurrencyEUR),
		TotalTTC: invoice.NewMoneyFromFloat(999.0, 2, invoice.CurrencyEUR),
	}

	body, _ := json.Marshal(inv)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/emit", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.EmitInvoice(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Statut HTTP 422 Unprocessable Entity attendu, obtenu : %d", rec.Code)
	}
}