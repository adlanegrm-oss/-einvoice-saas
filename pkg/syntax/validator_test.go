package syntax

import (
"testing"
"time"

"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

func TestValidateCoreBusinessRules(t *testing.T) {
inv := &canonical.CanonicalInvoice{
InvoiceNumber: "INV-2026-0001",
IssueDate:     time.Now().UTC(),
DueDate:       time.Now().UTC().Add(24 * time.Hour),
Currency:      "EUR",
Seller: canonical.Party{
ID:          "12345678900012",
TaxID:       "FR12345678901",
CountryCode: "FR",
},
TaxBreakdowns: []canonical.TaxBreakdown{
{VATCategory: "S", VATRate: 2000, BaseHT: 10000, TaxAmount: 2000},
},
MonetaryTotals: canonical.MonetaryTotals{
NetHT:      10000,
TaxAmount:  2000,
GrossTTC:   12000,
PayableDue: 12000,
},
}

rep := ValidateCoreBusinessRules(inv)
if !rep.Valid {
t.Fatalf("Facture valide rejetee : %+v", rep.Diagnostics)
}

inv.MonetaryTotals.GrossTTC = 99999
repBad := ValidateCoreBusinessRules(inv)
if repBad.Valid {
t.Fatal("Incoherence financiere TTC non detectee")
}
}