package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
	_ "modernc.org/sqlite"
)

type mockDir struct{}

func (m *mockDir) Lookup(ctx context.Context, participantID string) (*dispatcher.TargetEndpoint, error) {
	return &dispatcher.TargetEndpoint{
		ReceiverID:   participantID,
		PlatformName: "PDP Cible Test",
		AS4Endpoint:  "https://mock-as4.fr",
	}, nil
}

type mockAS4 struct{}

func (m *mockAS4) SendPayload(ctx context.Context, ep *dispatcher.TargetEndpoint, p []byte) (string, error) {
	return "ACK-AS4-OK-12345", nil
}

func setupTestPipeline(t *testing.T) (*InvoicePipeline, *repository.SQLiteInvoiceRepository) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("échec DB mémoire : %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		t.Fatalf("échec repo : %v", err)
	}

	val := en16931.NewValidator()
	sm := status.NewStateMachine()
	disp := dispatcher.NewDispatcher(&mockDir{}, &mockAS4{})

	pipe := NewInvoicePipeline(repo, val, sm, disp)
	return pipe, repo
}

func TestPipeline_ConformeEtAcheminee(t *testing.T) {
	pipe, repo := setupTestPipeline(t)
	ctx := context.Background()

	inv := invoice.Invoice{
		ID:        "INV-SUCCESS-01",
		Number:    "FA-2026-9999",
		Currency:  "EUR",
		IssueDate: time.Now(),
		Seller: invoice.Party{
			Name:  "Tech Fournisseur",
			SIRET: "11111111111111",
			Address: invoice.PostalAddress{
				StreetName:  "Rue A",
				PostalZone:  "75001",
				CityName:    "Paris",
				CountryCode: "FR",
			},
		},
		Customer: invoice.Party{
			Name:  "Client Final",
			SIRET: "22222222222222",
		},
		Items: []invoice.InvoiceItem{
			{Description: "Abonnement SaaS", Quantity: 1, UnitPrice: invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR), VATRate: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR)},
		},
		TotalHT: invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR),
		TotalVAT: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR),
		TotalTTC: invoice.NewMoneyFromFloat(120.0, 2, invoice.CurrencyEUR),
	}

	res, err := pipe.ProcessAndEmit(ctx, "tenant-a", inv)
	if err != nil {
		t.Fatalf("le pipeline aurait dû réussir : %v", err)
	}

	if res.Status != status.StateTransmitted {
		t.Errorf("statut attendu %s, obtenu %s", status.StateTransmitted, res.Status)
	}

	// Vérifier la piste d'audit générée en base
	history, err := repo.GetStatusHistory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("échec lecture historique : %v", err)
	}
	if len(history) != 3 { // DEPOSITED -> ISSUED -> TRANSMITTED
		t.Fatalf("attendu 3 événements dans la PAF, obtenu %d", len(history))
	}
}

func TestPipeline_RejetReglementaire(t *testing.T) {
	pipe, repo := setupTestPipeline(t)
	ctx := context.Background()

	// Facture sans devise et avec calcul TTC erroné
	inv := invoice.Invoice{
		ID:        "INV-FAIL-01",
		Number:    "FA-2026-0000",
		IssueDate: time.Now(),
		Seller: invoice.Party{
			Name:  "Tech Fournisseur",
			SIRET: "11111111111111",
		},
		Customer: invoice.Party{
			Name: "Client Invalide",
		},
		Items: []invoice.InvoiceItem{
			{Description: "Article", Quantity: 1, UnitPrice: invoice.NewMoneyFromFloat(50.0, 2, invoice.CurrencyEUR), VATRate: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR)},
		},
		TotalHT: invoice.NewMoneyFromFloat(50.0, 2, invoice.CurrencyEUR),
		TotalVAT: invoice.NewMoneyFromFloat(10.0, 2, invoice.CurrencyEUR),
		TotalTTC: invoice.NewMoneyFromFloat(999.0, 2, invoice.CurrencyEUR), // Erreur flagrante
	}

	res, err := pipe.ProcessAndEmit(ctx, "tenant-a", inv)
	if err != nil {
		t.Fatalf("le pipeline ne doit pas crasher sur une facture non conforme : %v", err)
	}

	if res.Status != status.StateRejected {
		t.Errorf("la facture aurait dû être rejetée, statut reçu : %s", res.Status)
	}

	history, err := repo.GetStatusHistory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("échec historique : %v", err)
	}
	if len(history) != 2 || history[1].ToState != status.StateRejected {
		t.Errorf("l'historique doit comporter le rejet : %+v", history)
	}
}