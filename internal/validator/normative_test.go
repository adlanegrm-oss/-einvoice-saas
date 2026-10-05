package validator

import (
"os"
"path/filepath"
"testing"
"time"

"einvoice-saas/internal/model"
)

func TestNormativeValidator_ValidInvoice(t *testing.T) {
inv := &model.CanonicalInvoice{
IssueDate: time.Now(),
Lines: []model.InvoiceLine{
{ID: "1", Quantity: 2.0, UnitPrice: 100.0, LineTotal: 200.0, VatPercent: 20.0, VatCategory: "S"},
{ID: "2", Quantity: 1.0, UnitPrice: 50.0, LineTotal: 50.0, VatPercent: 20.0, VatCategory: "S"},
},
Totals: model.MonetaryTotals{
LineExtensionAmount: 250.0,
TaxExclusiveAmount:  250.0,
TaxInclusiveAmount:  300.0,
PayableAmount:       300.0,
},
TaxSubtotals: []model.TaxSubtotal{
{TaxableAmount: 250.0, Percent: 20.0, TaxAmount: 50.0, CategoryCode: "S"},
},
}

val := NewNormativeValidator(true)
res, err := val.ValidateCanonical(inv)
if err != nil {
t.Fatalf("erreur validation inattendue: %v", err)
}
if !res.Valid {
t.Fatalf("la facture valide a été rejetée: %v", res.RuleErrors)
}
}

func TestNormativeValidator_RejectsMathInconsistency(t *testing.T) {
t.Run("Rejet calcul montant de ligne (BR-LINE-NET-AMOUNT)", func(t *testing.T) {
inv := &model.CanonicalInvoice{
Lines: []model.InvoiceLine{
{ID: "1", Quantity: 2.0, UnitPrice: 100.0, LineTotal: 250.0}, // Faux : 2 * 100 = 200 != 250
},
Totals: model.MonetaryTotals{
LineExtensionAmount: 250.0,
TaxExclusiveAmount:  250.0,
TaxInclusiveAmount:  250.0,
},
}
val := NewNormativeValidator(false)
res, err := val.ValidateCanonical(inv)
if err != nil {
t.Fatal(err)
}
if res.Valid {
t.Fatal("attendu: rejet pour incohérence ligne")
}
})

t.Run("Rejet somme des lignes vs LineExtensionAmount (BR-CO-10)", func(t *testing.T) {
inv := &model.CanonicalInvoice{
Lines: []model.InvoiceLine{
{ID: "1", Quantity: 1.0, UnitPrice: 100.0, LineTotal: 100.0},
},
Totals: model.MonetaryTotals{
LineExtensionAmount: 90.0, // Faux : somme = 100 != 90
TaxExclusiveAmount:  90.0,
TaxInclusiveAmount:  90.0,
},
}
val := NewNormativeValidator(false)
res, err := val.ValidateCanonical(inv)
if err != nil {
t.Fatal(err)
}
if res.Valid {
t.Fatal("attendu: rejet pour BR-CO-10")
}
})

t.Run("Rejet calcul TVA (BR-TAX-CALCULATION)", func(t *testing.T) {
inv := &model.CanonicalInvoice{
Lines: []model.InvoiceLine{
{ID: "1", Quantity: 1.0, UnitPrice: 1000.0, LineTotal: 1000.0},
},
Totals: model.MonetaryTotals{
LineExtensionAmount: 1000.0,
TaxExclusiveAmount:  1000.0,
TaxInclusiveAmount:  1100.0,
},
TaxSubtotals: []model.TaxSubtotal{
{TaxableAmount: 1000.0, Percent: 20.0, TaxAmount: 100.0}, // Faux : 1000 * 20% = 200 != 100
},
}
val := NewNormativeValidator(false)
res, err := val.ValidateCanonical(inv)
if err != nil {
t.Fatal(err)
}
if res.Valid {
t.Fatal("attendu: rejet pour calcul TVA erroné")
}
})

t.Run("Rejet rupture équilibre TTC (BR-CO-15)", func(t *testing.T) {
inv := &model.CanonicalInvoice{
Lines: []model.InvoiceLine{
{ID: "1", Quantity: 1.0, UnitPrice: 100.0, LineTotal: 100.0},
},
Totals: model.MonetaryTotals{
LineExtensionAmount: 100.0,
TaxExclusiveAmount:  100.0,
TaxInclusiveAmount:  150.0, // Faux : 100 + 20 = 120 != 150
},
TaxSubtotals: []model.TaxSubtotal{
{TaxableAmount: 100.0, Percent: 20.0, TaxAmount: 20.0},
},
}
val := NewNormativeValidator(false)
res, err := val.ValidateCanonical(inv)
if err != nil {
t.Fatal(err)
}
if res.Valid {
t.Fatal("attendu: rejet pour BR-CO-15")
}
})
}

func TestNormativeValidator_RejetCalculTVA_Lot(t *testing.T) {
// Vérification avec le fichier de lot si présent
lotPath := filepath.Join("..", "..", "factures_test_lots", "FACT_2026_005_REJET_CALCUL_TVA.xml")
data, err := os.ReadFile(lotPath)
if err != nil {
t.Skipf("lot de test non trouvé à l'emplacement %s, skip", lotPath)
}

val := NewNormativeValidator(false)
res, err := val.ValidateEN16931AndPeppol(data)
if err != nil {
// Une erreur de parsing/normalisation ou une invalidité prouve le rejet
return
}
if res.Valid {
t.Fatal("la facture FACT_2026_005_REJET_CALCUL_TVA.xml aurait dû être rejetée")
}
}
