package model_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"einvoice-saas/internal/model"
)

func TestCanonicalInvoice_Structure(t *testing.T) {
	inv := model.CanonicalInvoice{
		ID:               "FAC-001",
		IssueDate:        "2026-10-08",
		DocumentCurrency: "MAD",
		Seller: model.Party{
			Name:        "Fournisseur SA",
			LegalID:     "001524368000045",
			CountryCode: "MA",
		},
		Buyer: model.Party{
			Name:        "Client SARL",
			LegalID:     "002879541000088",
			CountryCode: "MA",
		},
		Totals: model.Totals{
			LineTotalAmount:    decimal.NewFromFloat(1000),
			TaxExclusiveAmount: decimal.NewFromFloat(1000),
			TaxInclusiveAmount: decimal.NewFromFloat(1200),
			PayableAmount:      decimal.NewFromFloat(1200),
		},
	}

	if inv.Seller.LegalID != "001524368000045" {
		t.Errorf("LegalID attendu 001524368000045, obtenu %s", inv.Seller.LegalID)
	}
	if inv.DocumentCurrency != "MAD" {
		t.Errorf("DocumentCurrency attendue MAD, obtenu %s", inv.DocumentCurrency)
	}
}
