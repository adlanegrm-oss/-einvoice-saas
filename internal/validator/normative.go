package validator

import (
"fmt"
"math"

"einvoice-saas/internal/model"
)

const floatTolerance = 0.02

type NormativeValidator struct {
StrictMode bool
}

func NewNormativeValidator(strictMode bool) *NormativeValidator {
return &NormativeValidator{
StrictMode: strictMode,
}
}

type ValidationResult struct {
Valid      bool
RuleErrors []string
}

// ValidateCanonical contrôle l'intégrité arithmétique EN 16931 du modèle canonique
func (v *NormativeValidator) ValidateCanonical(inv *model.CanonicalInvoice) (*ValidationResult, error) {
if inv == nil {
return nil, fmt.Errorf("normative: invoice is nil")
}

result := &ValidationResult{Valid: true, RuleErrors: make([]string, 0)}

// BR-LINE-NET-AMOUNT : Quantité * Prix unitaire = LineTotal
var sumLines float64
for _, line := range inv.Lines {
expectedLineTotal := math.Round(line.Quantity*line.UnitPrice*100) / 100
if math.Abs(expectedLineTotal-line.LineTotal) > floatTolerance {
result.Valid = false
result.RuleErrors = append(result.RuleErrors,
fmt.Sprintf("BR-LINE-NET-AMOUNT: ligne %s calcul incohérent (quantité: %.2f, prix: %.2f, attendu: %.2f, reçu: %.2f)",
line.ID, line.Quantity, line.UnitPrice, expectedLineTotal, line.LineTotal))
}
sumLines += line.LineTotal
}

// BR-CO-10 : Somme des lignes = LineExtensionAmount
sumLines = math.Round(sumLines*100) / 100
if math.Abs(sumLines-inv.Totals.LineExtensionAmount) > floatTolerance {
result.Valid = false
result.RuleErrors = append(result.RuleErrors,
fmt.Sprintf("BR-CO-10: somme des lignes (%.2f) != LineExtensionAmount (%.2f)",
sumLines, inv.Totals.LineExtensionAmount))
}

// BR-TAX-CALCULATION & Somme de la taxe
var sumTaxCalculated float64
for i, sub := range inv.TaxSubtotals {
expectedTax := math.Round(sub.TaxableAmount*(sub.Percent/100.0)*100) / 100
if math.Abs(expectedTax-sub.TaxAmount) > floatTolerance {
result.Valid = false
result.RuleErrors = append(result.RuleErrors,
fmt.Sprintf("BR-TAX-CALCULATION: sous-total TVA #%d incohérent (base: %.2f, taux: %.2f%%, attendu: %.2f, déclaré: %.2f)",
i+1, sub.TaxableAmount, sub.Percent, expectedTax, sub.TaxAmount))
}
sumTaxCalculated += sub.TaxAmount
}
sumTaxCalculated = math.Round(sumTaxCalculated*100) / 100

// BR-CO-15 : HT + TVA = TTC
expectedTTC := math.Round((inv.Totals.TaxExclusiveAmount+sumTaxCalculated)*100) / 100
if math.Abs(expectedTTC-inv.Totals.TaxInclusiveAmount) > floatTolerance {
result.Valid = false
result.RuleErrors = append(result.RuleErrors,
fmt.Sprintf("BR-CO-15: équilibre financier rompu (HT: %.2f + TVA: %.2f = %.2f != TTC: %.2f)",
inv.Totals.TaxExclusiveAmount, sumTaxCalculated, expectedTTC, inv.Totals.TaxInclusiveAmount))
}

// BR-CO-16 : Net à payer
if inv.Totals.PayableAmount < 0 && v.StrictMode {
result.Valid = false
result.RuleErrors = append(result.RuleErrors, "BR-CO-16: PayableAmount ne peut pas être strictement négatif")
}

return result, nil
}

// ValidateEN16931AndPeppol normalise le flux XML CII puis applique les règles normatives
func (v *NormativeValidator) ValidateEN16931AndPeppol(xmlData []byte) (*ValidationResult, error) {
canonical, err := model.NormalizeCIIToCanonical(xmlData)
if err != nil {
return nil, fmt.Errorf("échec de normalisation amont : %w", err)
}
return v.ValidateCanonical(canonical)
}
