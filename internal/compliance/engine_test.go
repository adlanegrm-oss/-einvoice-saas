package compliance_test

import (
	"os"
	"testing"
	"time"

	"einvoice-saas/internal/compliance"
	"einvoice-saas/internal/model"
)

func TestDispatcher_ValidateMoroccoInvoice(t *testing.T) {
	rawXML, err := os.ReadFile("../../factures_test_lots/ma/test_facture_maroc_ok.xml")
	if err != nil {
		t.Fatalf("Erreur lecture lot UBL: %v", err)
	}

	inv, err := model.NormalizeUBLToCanonical(rawXML)
	if err != nil {
		t.Fatalf("Erreur normalisation pivot: %v", err)
	}

	dispatcher := compliance.NewDispatcher()
	report, err := dispatcher.Validate(inv)
	if err != nil {
		t.Fatalf("Erreur execution dispatcher: %v", err)
	}

	if !report.Valid {
		t.Fatalf("La facture MA devrait être valide, anomalies trouvées: %+v", report.Issues)
	}

	if report.Jurisdiction != "MA" {
		t.Errorf("Juridiction attendue MA, obtenu: %s", report.Jurisdiction)
	}
}

func TestDispatcher_ValidateFranceInvoice(t *testing.T) {
	inv := &model.CanonicalInvoice{
		InvoiceNumber:      "FAC-FR-2026-0099",
		IssueDate:          time.Now(),
		Currency:           "EUR",
		TargetJurisdiction: "FR",
		Seller: model.Party{
			Name:    "Société France SAS",
			TaxID:   "FR12345678901",
			Country: "FR",
		},
		Buyer: model.Party{
			Name:    "Client France SAS",
			TaxID:   "FR98765432109",
			Country: "FR",
		},
		Lines: []model.InvoiceLine{
			{
				ID:        "1",
				LineTotal: 500.0,
			},
		},
		TaxSubtotals: []model.TaxSubtotal{
			{
				TaxableAmount: 500.0,
				TaxAmount:     100.0,
				Percent:       20.0,
			},
		},
		Totals: model.MonetaryTotals{
			LineExtensionAmount: 500.0,
			TaxExclusiveAmount:  500.0,
			TaxInclusiveAmount:  600.0,
			PayableAmount:       600.0,
		},
	}

	dispatcher := compliance.NewDispatcher()
	report, err := dispatcher.Validate(inv)
	if err != nil {
		t.Fatalf("Erreur execution dispatcher: %v", err)
	}

	if !report.Valid {
		t.Fatalf("La facture FR devrait être valide, anomalies: %+v", report.Issues)
	}

	if report.Jurisdiction != "FR" {
		t.Errorf("Juridiction attendue FR, obtenu: %s", report.Jurisdiction)
	}
}

func TestDispatcher_UnsupportedJurisdiction(t *testing.T) {
	inv := &model.CanonicalInvoice{
		ID:                 "TEST-UNKNOWN",
		InvoiceNumber:      "TEST-UNKNOWN",
		IssueDate:          time.Now(),
		Currency:           "EUR",
		TargetJurisdiction: "XX",
		Seller: model.Party{
			Name:    "Seller",
			Country: "XX",
		},
		Buyer: model.Party{
			Name:    "Buyer",
			Country: "XX",
		},
	}

	dispatcher := compliance.NewDispatcher()
	report, err := dispatcher.Validate(inv)
	if err != nil {
		t.Fatalf("Erreur inattendue: %v", err)
	}

	if report.Valid {
		t.Errorf("Le rapport devrait être invalide pour une juridiction inconnue")
	}

	foundUnsupported := false
	for _, issue := range report.Issues {
		if issue.RuleID == "SYS-JURISDICTION-UNSUPPORTED" {
			foundUnsupported = true
			break
		}
	}

	if !foundUnsupported {
		t.Errorf("Anomalie attendue SYS-JURISDICTION-UNSUPPORTED non trouvée dans: %+v", report.Issues)
	}
}
