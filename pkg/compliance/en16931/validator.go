package en16931

import (
"fmt"
"math"

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

// Validate applique les contrôles sémantiques et d'équilibres financiers EN 16931
func (v *Validator) Validate(inv *model.CanonicalInvoice) []RuleViolation {
var violations []RuleViolation
violations = append(violations, v.validateHeader(inv)...)
violations = append(violations, v.validateParties(inv)...)
violations = append(violations, v.validateMath(inv)...)
return violations
}

func (v *Validator) validateHeader(inv *model.CanonicalInvoice) []RuleViolation {
var violations []RuleViolation

// BR-01: An Invoice shall have an Invoice number (BT-1)
invoiceNum := inv.InvoiceNumber
if invoiceNum == "" {
invoiceNum = inv.ID
}
if invoiceNum == "" {
violations = append(violations, RuleViolation{
RuleID:   "BR-01",
Message:  "An Invoice shall have an Invoice number (BT-1).",
Path:     "Invoice.InvoiceNumber",
Severity: model.SeverityError,
})
}

// BR-02: An Invoice shall have an Invoice issue date (BT-2)
if inv.IssueDate.IsZero() {
violations = append(violations, RuleViolation{
RuleID:   "BR-02",
Message:  "An Invoice shall have an Invoice issue date (BT-2).",
Path:     "Invoice.IssueDate",
Severity: model.SeverityError,
})
}

// BR-05: An Invoice shall have an Invoice currency code (BT-5)
if inv.Currency == "" {
violations = append(violations, RuleViolation{
RuleID:   "BR-05",
Message:  "An Invoice shall have an Invoice currency code (BT-5).",
Path:     "Invoice.Currency",
Severity: model.SeverityError,
})
}

return violations
}

func (v *Validator) validateParties(inv *model.CanonicalInvoice) []RuleViolation {
var violations []RuleViolation

// BR-06: Seller name (BT-27)
if inv.Seller.Name == "" {
violations = append(violations, RuleViolation{
RuleID:   "BR-06",
Message:  "An Invoice shall contain the Seller name (BT-27).",
Path:     "Invoice.Seller.Name",
Severity: model.SeverityError,
})
}

// BR-07: Buyer name (BT-44)
if inv.Buyer.Name == "" {
violations = append(violations, RuleViolation{
RuleID:   "BR-07",
Message:  "An Invoice shall contain the Buyer name (BT-44).",
Path:     "Invoice.Buyer.Name",
Severity: model.SeverityError,
})
}

// BR-08: Seller postal address country code (BT-40)
if inv.Seller.Country == "" {
violations = append(violations, RuleViolation{
RuleID:   "BR-08",
Message:  "An Invoice shall contain the Seller postal address country code (BT-40).",
Path:     "Invoice.Seller.Country",
Severity: model.SeverityError,
})
}

// BR-09: Buyer postal address country code (BT-55)
if inv.Buyer.Country == "" {
violations = append(violations, RuleViolation{
RuleID:   "BR-09",
Message:  "An Invoice shall contain the Buyer postal address country code (BT-55).",
Path:     "Invoice.Buyer.Country",
Severity: model.SeverityError,
})
}

return violations
}

func (v *Validator) validateMath(inv *model.CanonicalInvoice) []RuleViolation {
var violations []RuleViolation

// BR-CO-10: Sum of Invoice line net amount = LineExtensionAmount (BT-106)
var expectedLineSum float64
for _, line := range inv.Lines {
expectedLineSum += line.LineTotal
}

if math.Abs(expectedLineSum-inv.Totals.LineExtensionAmount) > 0.01 {
violations = append(violations, RuleViolation{
RuleID: "BR-CO-10",
Message: fmt.Sprintf(
"Sum of Invoice line net amount (%.2f) must equal LineExtensionAmount (%.2f).",
expectedLineSum, inv.Totals.LineExtensionAmount,
),
Path:     "Invoice.Totals.LineExtensionAmount",
Severity: model.SeverityError,
})
}

// Somme totale des taxes déclarées dans TaxSubtotals
var sumTaxAmount float64
for _, subtotal := range inv.TaxSubtotals {
sumTaxAmount += subtotal.TaxAmount
}

// BR-CO-15: TaxInclusiveAmount (BT-112) = TaxExclusiveAmount (BT-109) + Total VAT
expectedGross := inv.Totals.TaxExclusiveAmount + sumTaxAmount
if math.Abs(expectedGross-inv.Totals.TaxInclusiveAmount) > 0.01 {
violations = append(violations, RuleViolation{
RuleID: "BR-CO-15",
Message: fmt.Sprintf(
"Invoice total amount with VAT (%.2f) must equal amount without VAT (%.2f) + VAT total amount (%.2f).",
inv.Totals.TaxInclusiveAmount, inv.Totals.TaxExclusiveAmount, sumTaxAmount,
),
Path:     "Invoice.Totals.TaxInclusiveAmount",
Severity: model.SeverityError,
})
}

return violations
}
