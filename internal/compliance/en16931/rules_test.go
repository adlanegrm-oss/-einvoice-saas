package en16931

import (
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

func validParty(name, siret string) invoice.Party {
	return invoice.Party{
		Name:  name,
		SIRET: siret,
		Address: invoice.PostalAddress{
			StreetName:  "10 Rue de la Paix",
			PostalZone:  "75001",
			CityName:    "Paris",
			CountryCode: "FR",
		},
	}
}

func TestValidator_EN16931(t *testing.T) {
	validator := NewValidator()

	// Cas 1 : Facture conforme
	validInv := &invoice.Invoice{
		ID:        "INV-001",
		Number:    "FA-2026-0001",
		Currency:  "EUR",
		Seller:    validParty("Fournisseur Tech", "12345678901234"),
		Customer:  validParty("Client Entreprise", "98765432109876"),
		IssueDate: time.Now(),
		TotalHT:   invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR),
		TotalVAT:  invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR),
		TotalTTC:  invoice.NewMoneyFromFloat(120.0, 2, invoice.CurrencyEUR),
		Items: []invoice.InvoiceItem{
			{Description: "Service Cloud", Quantity: 1, UnitPrice: invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR), VATRate: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR)},
		},
	}
	errs := validator.ValidateInvoice(validInv)
	if len(errs) != 0 {
		t.Fatalf("Facture valide refusée à tort : %+v", errs)
	}

	// Cas 2 : SIRET vendeur invalide
	badSiretInv := *validInv
	badSiretInv.Seller.SIRET = "123" // Pas 14 chiffres
	errs = validator.ValidateInvoice(&badSiretInv)
	if len(errs) == 0 || errs[0].RuleID != "BR-FR-01" {
		t.Fatalf("Le SIRET invalide aurait dû déclencher BR-FR-01 : %+v", errs)
	}

	// Cas 3 : Devise manquante
	noCurrInv := *validInv
	noCurrInv.Currency = ""
	errs = validator.ValidateInvoice(&noCurrInv)
	if len(errs) == 0 || errs[0].RuleID != "BR-CO-26" {
		t.Fatalf("La devise manquante aurait dû déclencher BR-CO-26 : %+v", errs)
	}
}