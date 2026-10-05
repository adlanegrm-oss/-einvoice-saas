package fr_test

import (
"testing"
"time"

"einvoice-saas/internal/compliance/validators/fr"
"einvoice-saas/internal/model"
)

func TestFranceCanonicalValidator_ValidInvoice(t *testing.T) {
inv := &model.CanonicalInvoice{
InvoiceNumber:      "FAC-FR-2026-0001",
IssueDate:          time.Now(),
TargetJurisdiction: "FR",
Seller: model.Party{
Name:    "Société Exemple SAS",
TaxID:   "FR12345678901",
Country: "FR",
},
TaxSubtotals: []model.TaxSubtotal{
{
TaxableAmount: 1000.0,
TaxAmount:     200.0,
Percent:       20.0,
},
},
}

validator := fr.NewFranceCanonicalValidator()
report := validator.Validate(inv)

if !report.Valid {
t.Fatalf("La facture FR devrait être valide, anomalies trouvées: %+v", report.Issues)
}

if report.Jurisdiction != "FR" {
t.Errorf("Juridiction attendue FR, obtenu: %s", report.Jurisdiction)
}
}

func TestFranceCanonicalValidator_InvalidSellerAndVAT(t *testing.T) {
inv := &model.CanonicalInvoice{
InvoiceNumber:      "",
TargetJurisdiction: "FR",
Seller: model.Party{
Name:  "Société Incomplète",
TaxID: "INVALID_TAX_ID",
},
TaxSubtotals: []model.TaxSubtotal{
{
TaxableAmount: 100.0,
TaxAmount:     12.0,
Percent:       12.0,
},
},
}

validator := fr.NewFranceCanonicalValidator()
report := validator.Validate(inv)

if report.Valid {
t.Fatalf("La facture aurait dû être rejetée")
}

if len(report.Issues) != 4 {
t.Errorf("Attendu 4 anomalies (TaxID, InvoiceNum, Date, TVA), obtenu: %d", len(report.Issues))
}
}
