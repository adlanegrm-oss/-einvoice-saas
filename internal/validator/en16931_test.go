package validator_test

import (
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain"
	"github.com/adlanegrm-oss/einvoice-saas/internal/validator"
)

func TestEN16931Validation(t *testing.T) {
	engine := validator.NewEN16931Engine("2026.1")

	totals := domain.MonetaryTotals{
		NetTotal:   domain.NewAmount(100, 0),
		TaxTotal:   domain.NewAmount(20, 0),
		GrossTotal: domain.NewAmount(120, 0),
		Prepaid:    0,
		Payable:    domain.NewAmount(120, 0),
	}

	// Cas valide avec SIRET 14 chiffres
	report := engine.ValidateFacture("INV-2026-001", "12345678901234", "98765432109876", totals, 1)
	if !report.Valid {
		t.Fatalf("La facture devrait être valide : %+v", report.Diagnostics)
	}

	// Cas invalide : numéro manquant et SIRET invalide
	reportInvalid := engine.ValidateFacture("", "12345678901234", "123", totals, 0)
	if reportInvalid.Valid {
		t.Fatal("La facture devrait être déclarée invalide")
	}
	if len(reportInvalid.Diagnostics) < 3 {
		t.Fatalf("Attendu au moins 3 diagnostics (BR-01, BR-16, FR-R-01), obtenu %d", len(reportInvalid.Diagnostics))
	}
}
