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
		Customer:  "Entreprise ACME",
		IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
		TotalHT:   100.0,
		TotalVAT:  20.0,
		TotalTTC:  120.0,
	}

	xmlData, err := GenerateFacturXXML(inv)
	if err != nil {
		t.Fatalf("Erreur de génération XML : %v", err)
	}

	xmlStr := string(xmlData)

	// Vérifications de la présence des balises Factur-X clés
	expectedTokens := []string{
		"CrossIndustryInvoice",
		"INV-2026-001",
		"Entreprise ACME",
		"100.00",
		"20.00",
		"120.00",
	}

	for _, token := range expectedTokens {
		if !strings.Contains(xmlStr, token) {
			t.Errorf("Le XML généré ne contient pas l'élément attendu : %s", token)
		}
	}
}
