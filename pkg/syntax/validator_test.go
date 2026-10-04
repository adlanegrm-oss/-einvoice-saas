package syntax

import (
"testing"
"time"

"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

func TestValidationRulesEN16931(t *testing.T) {
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

rep := ValidateEN16931Strict(inv)
if !rep.Valid {
t.Fatalf("Facture valide rejetee a tort : %+v", rep.Diagnostics)
}

// Forcer l'incohérence BR-CO-15
inv.MonetaryTotals.GrossTTC = 13000
repBad := ValidateEN16931Strict(inv)
if repBad.Valid {
t.Fatal("Facture avec faux montant TTC acceptee par erreur")
}
}

func TestCIIXmlGeneration(t *testing.T) {
inv := &canonical.CanonicalInvoice{
InvoiceNumber: "INV-TEST-CII",
IssueDate:     time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC),
Currency:      "EUR",
MonetaryTotals: canonical.MonetaryTotals{
NetHT: 5000, TaxAmount: 1000, GrossTTC: 6000, PayableDue: 6000,
},
}
xmlBytes, err := GenerateCIIXML(inv)
if err != nil {
t.Fatalf("Echec generation CII : %v", err)
}
if len(xmlBytes) == 0 {
t.Fatal("XML genere vide")
}
}