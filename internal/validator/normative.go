package validator

import (
	"fmt"

	"github.com/shopspring/decimal"

	"einvoice-saas/internal/model"
)

var floatTolerance = decimal.NewFromFloat(0.02)

type NormativeValidator struct {
	StrictMode bool
}

func NewNormativeValidator(strictMode bool) *NormativeValidator {
	return &NormativeValidator{StrictMode: strictMode}
}

// Renommer pour Ã©viter le conflit avec d'Ã©ventuels types locaux
type NormativeValidationResult struct {
	Valid      bool
	RuleErrors []string
}

func (v *NormativeValidator) ValidateCanonical(inv *model.CanonicalInvoice) (*NormativeValidationResult, error) {
	if inv == nil {
		return nil, fmt.Errorf("normative: invoice is nil")
	}

	result := &NormativeValidationResult{Valid: true, RuleErrors: make([]string, 0)}

	// BR-LINE-NET-AMOUNT : QuantitÃ© * Prix unitaire â‰ˆ LineTotalAmount
	sumLines := decimal.Zero
	for _, line := range inv.Lines {
		expectedLineTotal := line.Quantity.Mul(line.NetPrice).Round(2)
		if expectedLineTotal.Sub(line.LineTotalAmount).Abs().GreaterThan(floatTolerance) {
			result.Valid = false
			result.RuleErrors = append(result.RuleErrors,
				fmt.Sprintf("BR-LINE-NET-AMOUNT: ligne %s calcul incohÃ©rent (quantitÃ©: %s, prix: %s, attendu: %s, reÃ§u: %s)",
					line.ID,
					line.Quantity.StringFixed(2),
					line.NetPrice.StringFixed(2),
					expectedLineTotal.StringFixed(2),
					line.LineTotalAmount.StringFixed(2)))
		}
		sumLines = sumLines.Add(line.LineTotalAmount)
	}

	// BR-CO-10
	sumLines = sumLines.Round(2)
	if sumLines.Sub(inv.Totals.LineTotalAmount).Abs().GreaterThan(floatTolerance) {
		result.Valid = false
		result.RuleErrors = append(result.RuleErrors,
			fmt.Sprintf("BR-CO-10: somme des lignes (%s) != LineTotalAmount (%s)",
				sumLines.StringFixed(2), inv.Totals.LineTotalAmount.StringFixed(2)))
	}

	// BR-TAX-CALCULATION
	sumTaxCalculated := decimal.Zero
	hundred := decimal.NewFromInt(100)
	for i, sub := range inv.Taxes {
		expectedTax := sub.TaxableAmount.Mul(sub.Percent).Div(hundred).Round(2)
		if expectedTax.Sub(sub.TaxAmount).Abs().GreaterThan(floatTolerance) {
			result.Valid = false
			result.RuleErrors = append(result.RuleErrors,
				fmt.Sprintf("BR-TAX-CALCULATION: sous-total TVA #%d incohÃ©rent (base: %s, taux: %s%%, attendu: %s, dÃ©clarÃ©: %s)",
					i+1,
					sub.TaxableAmount.StringFixed(2),
					sub.Percent.StringFixed(2),
					expectedTax.StringFixed(2),
					sub.TaxAmount.StringFixed(2)))
		}
		sumTaxCalculated = sumTaxCalculated.Add(sub.TaxAmount)
	}
	sumTaxCalculated = sumTaxCalculated.Round(2)

	// BR-CO-15
	expectedTTC := inv.Totals.TaxExclusiveAmount.Add(sumTaxCalculated).Round(2)
	if expectedTTC.Sub(inv.Totals.TaxInclusiveAmount).Abs().GreaterThan(floatTolerance) {
		result.Valid = false
		result.RuleErrors = append(result.RuleErrors,
			fmt.Sprintf("BR-CO-15: Ã©quilibre financier rompu (HT: %s + TVA: %s = %s != TTC: %s)",
				inv.Totals.TaxExclusiveAmount.StringFixed(2),
				sumTaxCalculated.StringFixed(2),
				expectedTTC.StringFixed(2),
				inv.Totals.TaxInclusiveAmount.StringFixed(2)))
	}

	// BR-CO-16
	if inv.Totals.PayableAmount.IsNegative() && v.StrictMode {
		result.Valid = false
		result.RuleErrors = append(result.RuleErrors, "BR-CO-16: PayableAmount ne peut pas Ãªtre strictement nÃ©gatif")
	}

	return result, nil
}
