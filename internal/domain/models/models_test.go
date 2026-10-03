package models

import (
	"testing"
	"time"
)

func TestInvoiceModel_StructureAndAmounts(t *testing.T) {
	now := time.Now().UTC()

	inv := Invoice{
		ID:            "inv-uuid-001",
		InvoiceNumber: "INV-2026-0001",
		IssueDate:     now,
		Seller: Party{
			Name:    "Societe Alpha",
			TaxID:   "FR12345678901",
			Country: "FR",
			Address: "10 Rue de Paris",
		},
		Buyer: Party{
			Name:    "Societe Beta",
			TaxID:   "FR98765432109",
			Country: "FR",
			Address: "20 Avenue de Lyon",
		},
		Amounts: Amounts{
			Net:   100.00,
			Tax:   20.00,
			Total: 120.00,
		},
		Currency: "EUR",
		Status:   "DEPOSITED",
	}

	t.Run("Contrôle cohérence arithmétique montants", func(t *testing.T) {
		expectedTotal := inv.Amounts.Net + inv.Amounts.Tax
		if inv.Amounts.Total != expectedTotal {
			t.Errorf("Total = %.2f, attendu %.2f (Net + Tax)", inv.Amounts.Total, expectedTotal)
		}
	})

	t.Run("Présence identifiants obligatoires", func(t *testing.T) {
		if inv.InvoiceNumber == "" {
			t.Error("InvoiceNumber ne doit pas être vide")
		}
		if inv.Seller.TaxID == "" || inv.Buyer.TaxID == "" {
			t.Error("Seller.TaxID et Buyer.TaxID sont obligatoires")
		}
		if inv.Currency != "EUR" {
			t.Errorf("Currency attendue EUR, obtenu: %s", inv.Currency)
		}
	})
}