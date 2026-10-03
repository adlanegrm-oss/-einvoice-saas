package exporter

import (
	"strings"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

func TestGenerateFacturXXML(t *testing.T) {
	inv := invoice.Invoice{
		ID:        "INV-100",
		Number:    "INV-2026-001",
		Customer:  invoice.Party{Name: "Entreprise ACME"},
		IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
		TotalHT:   invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR),
		TotalVAT:  invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR),
		TotalTTC:  invoice.NewMoneyFromFloat(120.0, 2, invoice.CurrencyEUR),
	}

	xmlData, err := GenerateFacturXXML(inv)
	if err != nil {
		t.Fatalf("Erreur de génération XML : %v", err)
	}

	if !strings.Contains(string(xmlData), "INV-2026-001") {
		t.Errorf("Le XML généré ne contient pas le numéro de facture")
	}
}
