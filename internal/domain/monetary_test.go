package domain_test

import (
"testing"

"github.com/adlanegrm-oss/einvoice-saas/internal/domain"
)

func TestMonetaryCalculations(t *testing.T) {
net := domain.NewAmount(100, 50) // 100.50 EUR
vatRate := int64(2000)           // 20.00%
tax := net.CalculateTax(vatRate)

if tax.Cents() != 2010 { // 20.10 EUR
t.Fatalf("Calcul TVA incorrect : attendu 2010 centimes, obtenu %d", tax.Cents())
}

totals := domain.MonetaryTotals{
NetTotal:   net,
TaxTotal:   tax,
GrossTotal: net.Add(tax),
Prepaid:    domain.FromCents(0),
Payable:    net.Add(tax),
}

if err := totals.Validate(); err != nil {
t.Fatalf("Validation des totaux en échec : %v", err)
}
}

func TestExternalStateProjection(t *testing.T) {
ext := domain.ComputeExternalState(domain.ComplianceValid, domain.TransmissionDelivered, domain.PaymentNotDue)
if ext != domain.ExtDelivered {
t.Fatalf("Attendu DELIVERED, obtenu %s", ext)
}

extPaid := domain.ComputeExternalState(domain.ComplianceValid, domain.TransmissionDelivered, domain.PaymentPaid)
if extPaid != domain.ExtPaid {
t.Fatalf("Attendu PAID, obtenu %s", extPaid)
}
}
