package fr_test

import (
	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/model"
	"github.com/shopspring/decimal"
	"testing"
)

func validBaseInvoice() *model.CanonicalInvoice {
	return &model.CanonicalInvoice{
		ID:                "INV-2026-0001",
		TypeCode:          "380",
		IssueDate:         "2026-10-06",
		OperationCategory: "services",
		Seller: model.Party{
			Name:        "Fournisseur FR SAS",
			VATID:       "FR12345678901",
			LegalID:     "12345678900012",
			CountryCode: "FR",
		},
		Buyer: model.Party{
			Name:        "Client FR SARL",
			LegalID:     "98765432100019",
			CountryCode: "FR",
		},
		Taxes: []model.TaxSubtotal{
			{
				TaxableAmount:   decimal.NewFromFloat(1000.0),
				TaxAmount:       decimal.NewFromFloat(200.0),
				Percent:         decimal.NewFromFloat(20.0),
				TaxCategoryCode: "S",
			},
		},
	}
}

func TestFranceCanonicalValidator_FiscalCoverage(t *testing.T) {
	v := fr.NewFranceCanonicalValidator()

	t.Run("Facture_Conforme", func(t *testing.T) {
		inv := validBaseInvoice()
		rep := v.Validate(inv)
		if !rep.Valid {
			t.Fatalf("attendu valide, obtenu erreurs: %+v", rep.Issues)
		}
	})

	t.Run("Rejet_Avoir_Sans_Reference", func(t *testing.T) {
		inv := validBaseInvoice()
		inv.TypeCode = "381"
		inv.PrecedingInvoices = []model.PrecedingInvoiceRef{{ID: ""}}

		rep := v.Validate(inv)
		if rep.Valid {
			t.Fatalf("attendu rejet pour avoir sans reference")
		}
		var found bool
		for _, issue := range rep.Issues {
			if issue.RuleID == "FR-RULE-CREDIT-NOTE-REF-01" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("attendu rÃ¨gle FR-RULE-CREDIT-NOTE-REF-01, obtenu: %+v", rep.Issues)
		}
	})

	t.Run("Avoir_Conforme_Avec_Reference", func(t *testing.T) {
		inv := validBaseInvoice()
		inv.TypeCode = "381"
		inv.PrecedingInvoices = []model.PrecedingInvoiceRef{{ID: "INV-2026-0001"}}

		rep := v.Validate(inv)
		if !rep.Valid {
			t.Fatalf("attendu valide, obtenu: %+v", rep.Issues)
		}
	})

	t.Run("Rejet_Exoneration_Sans_Motif", func(t *testing.T) {
		inv := validBaseInvoice()
		inv.Taxes = []model.TaxSubtotal{
			{
				TaxableAmount:   decimal.NewFromFloat(500.0),
				TaxAmount:       decimal.NewFromFloat(0.0),
				Percent:         decimal.NewFromFloat(0.0),
				TaxCategoryCode: "E",
				ExemptionReason: "", // Manquant
			},
		}

		rep := v.Validate(inv)
		if rep.Valid {
			t.Fatalf("attendu rejet pour exonÃ©ration sans motif lÃ©gal")
		}
		var found bool
		for _, issue := range rep.Issues {
			if issue.RuleID == "FR-RULE-VAT-EXEMPT-REASON-01" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("attendu rÃ¨gle FR-RULE-VAT-EXEMPT-REASON-01, obtenu: %+v", rep.Issues)
		}
	})

	t.Run("Exoneration_Conforme_Avec_Motif", func(t *testing.T) {
		inv := validBaseInvoice()
		inv.Taxes = []model.TaxSubtotal{
			{
				TaxableAmount:   decimal.NewFromFloat(500.0),
				TaxAmount:       decimal.NewFromFloat(0.0),
				Percent:         decimal.NewFromFloat(0.0),
				TaxCategoryCode: "E",
				ExemptionReason: "Article 262 ter I du CGI",
			},
		}

		rep := v.Validate(inv)
		if !rep.Valid {
			t.Fatalf("attendu valide, obtenu: %+v", rep.Issues)
		}
	})

	t.Run("Rejet_Client_FR_Sans_SIREN_SIRET", func(t *testing.T) {
		inv := validBaseInvoice()
		inv.Buyer.LegalID = ""
		inv.Buyer.VATID = ""

		rep := v.Validate(inv)
		if rep.Valid {
			t.Fatalf("attendu rejet acheteur franÃ§ais sans identifiant")
		}
		var found bool
		for _, issue := range rep.Issues {
			if issue.RuleID == "FR-RULE-BUYER-ID-01" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("attendu rÃ¨gle FR-RULE-BUYER-ID-01, obtenu: %+v", rep.Issues)
		}
	})
}
