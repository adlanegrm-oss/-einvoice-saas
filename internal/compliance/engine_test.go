package compliance

import (
"testing"

"github.com/shopspring/decimal"

"einvoice-saas/internal/model"
)

func TestDispatcher_ValidateMoroccoInvoice(t *testing.T) {
inv := &model.CanonicalInvoice{
ID:               "MA-001",
IssueDate:        "2026-10-08",
DocumentCurrency: "MAD",
Seller: model.Party{
Name:        "Seller MA",
LegalID:     "001234567000012",
CountryCode: "MA",
},
Buyer: model.Party{
Name:        "Buyer MA",
LegalID:     "001234567000013",
CountryCode: "MA",
},
Lines: []model.Line{{
ID: "1", Quantity: decimal.NewFromFloat(1),
NetPrice: decimal.NewFromFloat(500), LineTotalAmount: decimal.NewFromFloat(500),
}},
Taxes: []model.TaxSubtotal{{
TaxableAmount: decimal.NewFromFloat(500), TaxAmount: decimal.NewFromFloat(100),
Percent: decimal.NewFromFloat(20), TaxCategoryCode: "S",
}},
Totals: model.Totals{
LineTotalAmount: decimal.NewFromFloat(500), TaxExclusiveAmount: decimal.NewFromFloat(500),
TaxInclusiveAmount: decimal.NewFromFloat(600), PayableAmount: decimal.NewFromFloat(600),
},
}

d := NewDispatcher()
report, err := d.Validate(inv)
if err != nil {
t.Fatalf("Validate error: %v", err)
}
if !report.Valid {
t.Fatalf("La facture MA devrait être valide, anomalies: %+v", report.Issues)
}
}

func TestDispatcher_ValidateFranceInvoice(t *testing.T) {
inv := &model.CanonicalInvoice{
ID: "FR-001", IssueDate: "2026-10-08", DocumentCurrency: "EUR",
TypeCode: "380", OperationCategory: "goods",
Seller: model.Party{
Name: "Seller FR", VATID: "FR12345678901",
LegalID: "123456789", CountryCode: "FR",
},
Buyer: model.Party{
Name: "Buyer FR", LegalID: "987654321", CountryCode: "FR",
},
Lines: []model.Line{{
ID: "1", Quantity: decimal.NewFromFloat(1),
NetPrice: decimal.NewFromFloat(1000), LineTotalAmount: decimal.NewFromFloat(1000),
}},
Taxes: []model.TaxSubtotal{{
TaxableAmount: decimal.NewFromFloat(1000), TaxAmount: decimal.NewFromFloat(200),
Percent: decimal.NewFromFloat(20), TaxCategoryCode: "S",
}},
Totals: model.Totals{
LineTotalAmount: decimal.NewFromFloat(1000), TaxExclusiveAmount: decimal.NewFromFloat(1000),
TaxInclusiveAmount: decimal.NewFromFloat(1200), PayableAmount: decimal.NewFromFloat(1200),
},
}

d := NewDispatcher()
report, err := d.Validate(inv)
if err != nil {
t.Fatalf("Validate error: %v", err)
}
if !report.Valid {
t.Fatalf("La facture FR devrait être valide, anomalies: %+v", report.Issues)
}
}
