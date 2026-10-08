package en16931

import (
	"fmt"

	"github.com/shopspring/decimal"

	"einvoice-saas/internal/model"
)

type RuleViolation struct {
	RuleID   string              `json:"rule_id"`
	Message  string              `json:"message"`
	Path     string              `json:"path"`
	Severity model.IssueSeverity `json:"severity"`
}

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(inv *model.CanonicalInvoice) []RuleViolation {
	var violations []RuleViolation
	violations = append(violations, v.validateHeader(inv)...)
	violations = append(violations, v.validateParties(inv)...)
	violations = append(violations, v.validateMath(inv)...)
	return violations
}

func (v *Validator) validateHeader(inv *model.CanonicalInvoice) []RuleViolation {
	var violations []RuleViolation

	// BR-01: Invoice number (BT-1) → ID
	if inv.ID == "" {
		violations = append(violations, RuleViolation{
			RuleID:   "BR-01",
			Message:  "An Invoice shall have an Invoice number (BT-1).",
			Path:     "Invoice.ID",
			Severity: model.SeverityError,
		})
	}

	// BR-02: Issue date (BT-2) — string, pas time.Time
	if inv.IssueDate == "" {
		violations = append(violations, RuleViolation{
			RuleID:   "BR-02",
			Message:  "An Invoice shall have an Invoice issue date (BT-2).",
			Path:     "Invoice.IssueDate",
			Severity: model.SeverityError,
		})
	}

	// BR-05: Currency (BT-5) → DocumentCurrency
	if inv.DocumentCurrency == "" {
		violations = append(violations, RuleViolation{
			RuleID:   "BR-05",
			Message:  "An Invoice shall have an Invoice currency code (BT-5).",
			Path:     "Invoice.DocumentCurrency",
			Severity: model.SeverityError,
		})
	}

	return violations
}

func (v *Validator) validateParties(inv *model.CanonicalInvoice) []RuleViolation {
	var violations []RuleViolation

	if inv.Seller.Name == "" {
		violations = append(violations, RuleViolation{
			RuleID:   "BR-06",
			Message:  "An Invoice shall contain the Seller name (BT-27).",
			Path:     "Invoice.Seller.Name",
			Severity: model.SeverityError,
		})
	}

	if inv.Buyer.Name == "" {
		violations = append(violations, RuleViolation{
			RuleID:   "BR-07",
			Message:  "An Invoice shall contain the Buyer name (BT-44).",
			Path:     "Invoice.Buyer.Name",
			Severity: model.SeverityError,
		})
	}

	// Country → CountryCode
	if inv.Seller.CountryCode == "" {
		violations = append(violations, RuleViolation{
			RuleID:   "BR-08",
			Message:  "An Invoice shall contain the Seller postal address country code (BT-40).",
			Path:     "Invoice.Seller.CountryCode",
			Severity: model.SeverityError,
		})
	}

	if inv.Buyer.CountryCode == "" {
		violations = append(violations, RuleViolation{
			RuleID:   "BR-09",
			Message:  "An Invoice shall contain the Buyer postal address country code (BT-55).",
			Path:     "Invoice.Buyer.CountryCode",
			Severity: model.SeverityError,
		})
	}

	return violations
}

func (v *Validator) validateMath(inv *model.CanonicalInvoice) []RuleViolation {
	var violations []RuleViolation
	tol := decimal.NewFromFloat(0.01)

	// BR-CO-10 : somme des lignes = LineTotalAmount
	expectedLineSum := decimal.Zero
	for _, line := range inv.Lines {
		expectedLineSum = expectedLineSum.Add(line.LineTotalAmount)
	}

	if expectedLineSum.Sub(inv.Totals.LineTotalAmount).Abs().GreaterThan(tol) {
		violations = append(violations, RuleViolation{
			RuleID: "BR-CO-10",
			Message: fmt.Sprintf(
				"Sum of Invoice line net amount (%s) must equal LineTotalAmount (%s).",
				expectedLineSum.StringFixed(2), inv.Totals.LineTotalAmount.StringFixed(2),
			),
			Path:     "Invoice.Totals.LineTotalAmount",
			Severity: model.SeverityError,
		})
	}

	// Somme des taxes (Taxes, pas TaxSubtotals)
	sumTaxAmount := decimal.Zero
	for _, subtotal := range inv.Taxes {
		sumTaxAmount = sumTaxAmount.Add(subtotal.TaxAmount)
	}

	// BR-CO-15 : HT + TVA = TTC
	expectedGross := inv.Totals.TaxExclusiveAmount.Add(sumTaxAmount)
	if expectedGross.Sub(inv.Totals.TaxInclusiveAmount).Abs().GreaterThan(tol) {
		violations = append(violations, RuleViolation{
			RuleID: "BR-CO-15",
			Message: fmt.Sprintf(
				"Invoice total amount with VAT (%s) must equal amount without VAT (%s) + VAT total amount (%s).",
				inv.Totals.TaxInclusiveAmount.StringFixed(2),
				inv.Totals.TaxExclusiveAmount.StringFixed(2),
				sumTaxAmount.StringFixed(2),
			),
			Path:     "Invoice.Totals.TaxInclusiveAmount",
			Severity: model.SeverityError,
		})
	}

	return violations
}
