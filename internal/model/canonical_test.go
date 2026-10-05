package model_test

import (
	"testing"
	"time"

	"einvoice-saas/internal/model"
)

func TestCanonicalInvoice_Structure(t *testing.T) {
	inv := model.CanonicalInvoice{
		ID:            "INV-2026-001",
		InvoiceNumber: "FAC-001",
		IssueDate:     time.Now(),
		Currency:      "MAD",
		SourceSyntax:  "UBL-2.1",
		Seller: model.Party{
			Name:       "Fournisseur SA",
			NationalID: "001524368000045",
			Country:    "MA",
		},
		Buyer: model.Party{
			Name:       "Client SARL",
			NationalID: "002879541000088",
			Country:    "MA",
		},
		Totals: model.MonetaryTotals{
			LineExtensionAmount: 1000.0,
			TaxExclusiveAmount:  1000.0,
			TaxInclusiveAmount:  1200.0,
			PayableAmount:       1200.0,
		},
	}

	if inv.Seller.NationalID != "001524368000045" {
		t.Errorf("NationalID attendu 001524368000045, obtenu %s", inv.Seller.NationalID)
	}

	if inv.Currency != "MAD" {
		t.Errorf("Currency attendue MAD, obtenu %s", inv.Currency)
	}
}