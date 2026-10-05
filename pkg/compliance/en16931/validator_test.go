package en16931

import (
"testing"
"time"

"einvoice-saas/internal/model"
)

func TestValidator_ValidCanonicalInvoice(t *testing.T) {
v := NewValidator()

inv := &model.CanonicalInvoice{
ID:            "INV-2026-001",
InvoiceNumber: "INV-2026-001",
IssueDate:     time.Now(),
Currency:      "EUR",
Seller: model.Party{
Name:    "Fournisseur SAS",
Country: "FR",
},
Buyer: model.Party{
Name:    "Client SA",
Country: "FR",
},
Lines: []model.InvoiceLine{
{ID: "1", LineTotal: 100.00, Quantity: 1, UnitPrice: 100.00},
{ID: "2", LineTotal: 50.00, Quantity: 1, UnitPrice: 50.00},
},
TaxSubtotals: []model.TaxSubtotal{
{TaxableAmount: 150.00, TaxAmount: 30.00, Percent: 20.0},
},
Totals: model.MonetaryTotals{
LineExtensionAmount: 150.00,
TaxExclusiveAmount:  150.00,
TaxInclusiveAmount:  180.00,
PayableAmount:       180.00,
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
Currency: "EUR",
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
InvoiceNumber: "INV-ERR",
IssueDate:     time.Now(),
Currency:      "EUR",
Seller:        model.Party{Name: "Seller", Country: "FR"},
Buyer:         model.Party{Name: "Buyer", Country: "FR"},
Lines: []model.InvoiceLine{
{ID: "1", LineTotal: 100.00},
},
TaxSubtotals: []model.TaxSubtotal{
{TaxableAmount: 100.00, TaxAmount: 20.00, Percent: 20.0},
},
Totals: model.MonetaryTotals{
LineExtensionAmount: 90.00,
TaxExclusiveAmount:  100.00,
TaxInclusiveAmount:  110.00,
PayableAmount:       110.00,
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
