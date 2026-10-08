package en16931

import (
	"github.com/shopspring/decimal"
	"testing"
	"einvoice-saas/internal/model"
)

func TestValidator_ValidCanonicalInvoice(t *testing.T) {
	v := NewValidator()

	inv := &model.CanonicalInvoice{
		ID:            "INV-2026-001",
		IssueDate: "2026-10-08",
		DocumentCurrency:      "EUR",
		Seller: model.Party{
			Name:    "Fournisseur SAS",
			CountryCode: "FR",
		},
		Buyer: model.Party{
			Name:    "Client SA",
			CountryCode: "FR",
		},
		Lines: []model.Line{
			{ID: "1", LineTotalAmount: decimal.NewFromFloat(100.00), Quantity: decimal.NewFromFloat(1), NetPrice: decimal.NewFromFloat(100.00)},
			{ID: "2", LineTotalAmount: decimal.NewFromFloat(50.00), Quantity: decimal.NewFromFloat(1), NetPrice: decimal.NewFromFloat(50.00)},
		},
		Taxes: []model.TaxSubtotal{
			{TaxableAmount: decimal.NewFromFloat(150.00), TaxAmount: decimal.NewFromFloat(30.00), Percent: decimal.NewFromFloat(20.0)},
		},
		Totals: model.Totals{
			LineTotalAmount: decimal.NewFromFloat(150.00),
			TaxExclusiveAmount:  decimal.NewFromFloat(150.00),
			TaxInclusiveAmount:  decimal.NewFromFloat(180.00),
			PayableAmount:       decimal.NewFromFloat(180.00),
		},
	}

	violations := v.Validate(inv)
	if len(violations) > 0 {
		t.Fatalf("expected 0 violations, got %d: %+v", len(violations), violations)
	}
}

func TestValidator_MissingMandatoryPartiesAndHeaders(t *testing.T) {
	v := NewValidator()

	inv := &model.CanonicalInvoice{
		DocumentCurrency: "EUR",
	}

	violations := v.Validate(inv)
	if len(violations) < 6 {
		t.Errorf("expected at least 6 violations, got %d: %+v", len(violations), violations)
	}
}

func TestValidator_MathDiscrepancy(t *testing.T) {
	v := NewValidator()

	inv := &model.CanonicalInvoice{
		ID:            "INV-ERR",
		IssueDate: "2026-10-08",
		DocumentCurrency:      "EUR",
		Seller:        model.Party{Name: "Seller", CountryCode: "FR"},
		Buyer:         model.Party{Name: "Buyer", CountryCode: "FR"},
		Lines: []model.Line{
			{ID: "1", LineTotalAmount: decimal.NewFromFloat(100.00)},
		},
		Taxes: []model.TaxSubtotal{
			{TaxableAmount: decimal.NewFromFloat(100.00), TaxAmount: decimal.NewFromFloat(20.00), Percent: decimal.NewFromFloat(20.0)},
		},
		Totals: model.Totals{
			LineTotalAmount: decimal.NewFromFloat(90.00),
			TaxExclusiveAmount:  decimal.NewFromFloat(100.00),
			TaxInclusiveAmount:  decimal.NewFromFloat(110.00),
			PayableAmount:       decimal.NewFromFloat(110.00),
		},
	}

	violations := v.Validate(inv)
	foundBRCO10 := false
	foundBRCO15 := false

	for _, violation := range violations {
		if violation.RuleID == "BR-CO-10" {
			foundBRCO10 = true
		}
		if violation.RuleID == "BR-CO-15" {
			foundBRCO15 = true
		}
	}

	if !foundBRCO10 || !foundBRCO15 {
		t.Errorf("expected BR-CO-10 and BR-CO-15 to trigger, got: %+v", violations)
	}
}
