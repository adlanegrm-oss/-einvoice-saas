package validator

import (
"testing"

"github.com/shopspring/decimal"
)

func TestValidateEN16931CoreRules_ValidInvoice(t *testing.T) {
totals := InvoiceMonetaryTotals{
LineExtensionAmount: decimal.NewFromFloat(100.00),
TaxExclusiveAmount:  decimal.NewFromFloat(100.00),
TaxTotalAmount:      decimal.NewFromFloat(20.00),
TaxInclusiveAmount:  decimal.NewFromFloat(120.00),
PrepaidAmount:       decimal.NewFromFloat(0.00),
RoundingAmount:      decimal.NewFromFloat(0.00),
PayableAmount:       decimal.NewFromFloat(120.00),
AllowanceTotal:      decimal.Zero,
ChargeTotal:         decimal.Zero,
}

report := ValidateEN16931CoreRules(totals)
if !report.Valid {
t.Fatalf("expected invoice to be valid, got errors: %+v", report.Results)
}
if report.ErrorsCount != 0 {
t.Errorf("expected 0 errors, got: %d", report.ErrorsCount)
}
}

func TestValidateEN16931CoreRules_BR_CO_10_Failure(t *testing.T) {
// BT-115 (115.00) != BT-112 (120.00) - BT-113 (0.00) + BT-114 (0.00)
totals := InvoiceMonetaryTotals{
LineExtensionAmount: decimal.NewFromFloat(100.00),
TaxExclusiveAmount:  decimal.NewFromFloat(100.00),
TaxTotalAmount:      decimal.NewFromFloat(20.00),
TaxInclusiveAmount:  decimal.NewFromFloat(120.00),
PrepaidAmount:       decimal.Zero,
RoundingAmount:      decimal.Zero,
PayableAmount:       decimal.NewFromFloat(115.00),
AllowanceTotal:      decimal.Zero,
ChargeTotal:         decimal.Zero,
}

report := ValidateEN16931CoreRules(totals)
if report.Valid {
t.Fatal("expected report to be invalid due to BR-CO-10 violation")
}

found := false
for _, res := range report.Results {
if res.Code == "BR-CO-10" {
found = true
break
}
}
if !found {
t.Error("expected diagnostic error code BR-CO-10")
}
}

func TestValidateEN16931CoreRules_BR_CO_16_NegativePayable(t *testing.T) {
totals := InvoiceMonetaryTotals{
LineExtensionAmount: decimal.NewFromFloat(-10.00),
TaxExclusiveAmount:  decimal.NewFromFloat(-10.00),
TaxTotalAmount:      decimal.NewFromFloat(-2.00),
TaxInclusiveAmount:  decimal.NewFromFloat(-12.00),
PrepaidAmount:       decimal.Zero,
RoundingAmount:      decimal.Zero,
PayableAmount:       decimal.NewFromFloat(-12.00),
AllowanceTotal:      decimal.Zero,
ChargeTotal:         decimal.Zero,
}

report := ValidateEN16931CoreRules(totals)
if report.Valid {
t.Fatal("expected report to be invalid due to negative payable amount")
}

found := false
for _, res := range report.Results {
if res.Code == "BR-CO-16" {
found = true
break
}
}
if !found {
t.Error("expected diagnostic error code BR-CO-16")
}
}
