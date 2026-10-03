package validation_test

import (
"testing"
"github.com/adlanegrm-oss/einvoice-saas/pkg/money"
"github.com/adlanegrm-oss/einvoice-saas/pkg/tax"
"github.com/adlanegrm-oss/einvoice-saas/pkg/validation"
)

func TestValidateStrictEN16931(t *testing.T) {
base := money.New(10000, money.EUR)
st, _ := tax.CalculateVATBreakdown(tax.StandardRate, tax.FromFloatPercent(20.0), base, "", "")

inv := validation.InvoiceTotals{
InvoiceNumber:      "INV-2026-001",
Currency:           money.EUR,
SumInvoiceLines:    base,
AllowanceTotal:     money.New(0, money.EUR),
ChargeTotal:        money.New(0, money.EUR),
TaxExclusiveAmount: base,
TaxTotalAmount:     st.TaxAmount,
TaxInclusiveAmount: money.New(12000, money.EUR),
VATBreakdowns:      []tax.Subtotal{st},
}

report := validation.ValidateStrictEN16931(inv)
if !report.Valid {
t.Fatalf("Facture valide rejetée: %+v", report.Diagnostics)
}
}

func TestToleranceStrictRejection(t *testing.T) {
// Ancien bug : un écart de 0.02€ était toléré. Il doit désormais être STRICTEMENT rejeté (BR-CO-15).
base := money.New(10000, money.EUR) // 100.00 EUR
vat := money.New(2000, money.EUR)   // 20.00 EUR

invWithDrift := validation.InvoiceTotals{
InvoiceNumber:      "INV-2026-DRIFT",
Currency:           money.EUR,
SumInvoiceLines:    base,
AllowanceTotal:     money.New(0, money.EUR),
ChargeTotal:        money.New(0, money.EUR),
TaxExclusiveAmount: base,
TaxTotalAmount:     vat,
TaxInclusiveAmount: money.New(12002, money.EUR), // 120.02 EUR au lieu de 120.00 EUR (+2 centimes)
VATBreakdowns: []tax.Subtotal{
{Category: tax.StandardRate, Rate: 2000, TaxableAmount: base, TaxAmount: vat},
},
}

report := validation.ValidateStrictEN16931(invWithDrift)
if report.Valid {
t.Fatalf("La règle BR-CO-15 aurait dû rejeter l'écart de 2 centimes")
}

foundBRCO15 := false
for _, diag := range report.Diagnostics {
if diag.RuleID == "BR-CO-15" {
foundBRCO15 = true
break
}
}
if !foundBRCO15 {
t.Fatalf("Diagnostic BR-CO-15 manquant dans le rapport: %+v", report.Diagnostics)
}
}
