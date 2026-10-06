package fr_test

import (
"testing"
"time"

"einvoice-saas/internal/compliance/validators/fr"
"einvoice-saas/internal/model"
)

func validBaseInvoice() *model.CanonicalInvoice {
return &model.CanonicalInvoice{
InvoiceNumber:     "INV-2026-0001",
InvoiceTypeCode:   "380",
IssueDate:         time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC),
OperationCategory: "services",
Seller: model.Party{
Name:       "Fournisseur FR SAS",
TaxID:      "FR12345678901",
NationalID: "12345678900012",
Country:    "FR",
},
Buyer: model.Party{
Name:       "Client FR SARL",
NationalID: "98765432100019",
Country:    "FR",
},
TaxSubtotals: []model.TaxSubtotal{
{
TaxableAmount: 1000.0,
TaxAmount:     200.0,
Percent:       20.0,
CategoryCode:  "S",
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
inv.InvoiceTypeCode = "381"
inv.PrecedingInvoiceReference = ""

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
t.Fatalf("attendu règle FR-RULE-CREDIT-NOTE-REF-01, obtenu: %+v", rep.Issues)
}
})

t.Run("Avoir_Conforme_Avec_Reference", func(t *testing.T) {
inv := validBaseInvoice()
inv.InvoiceTypeCode = "381"
inv.PrecedingInvoiceReference = "INV-2026-0001"

rep := v.Validate(inv)
if !rep.Valid {
t.Fatalf("attendu valide, obtenu: %+v", rep.Issues)
}
})

t.Run("Rejet_Exoneration_Sans_Motif", func(t *testing.T) {
inv := validBaseInvoice()
inv.TaxSubtotals = []model.TaxSubtotal{
{
TaxableAmount:   500.0,
TaxAmount:       0.0,
Percent:         0.0,
CategoryCode:    "E",
ExemptionReason: "", // Manquant
},
}

rep := v.Validate(inv)
if rep.Valid {
t.Fatalf("attendu rejet pour exonération sans motif légal")
}
var found bool
for _, issue := range rep.Issues {
if issue.RuleID == "FR-RULE-VAT-EXEMPT-REASON-01" {
found = true
break
}
}
if !found {
t.Fatalf("attendu règle FR-RULE-VAT-EXEMPT-REASON-01, obtenu: %+v", rep.Issues)
}
})

t.Run("Exoneration_Conforme_Avec_Motif", func(t *testing.T) {
inv := validBaseInvoice()
inv.TaxSubtotals = []model.TaxSubtotal{
{
TaxableAmount:   500.0,
TaxAmount:       0.0,
Percent:         0.0,
CategoryCode:    "E",
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
inv.Buyer.NationalID = ""
inv.Buyer.TaxID = ""

rep := v.Validate(inv)
if rep.Valid {
t.Fatalf("attendu rejet acheteur français sans identifiant")
}
var found bool
for _, issue := range rep.Issues {
if issue.RuleID == "FR-RULE-BUYER-ID-01" {
found = true
break
}
}
if !found {
t.Fatalf("attendu règle FR-RULE-BUYER-ID-01, obtenu: %+v", rep.Issues)
}
})
}
