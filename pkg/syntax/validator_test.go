package syntax

import (
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

func TestValidateCoreBusinessRules(t *testing.T) {
	inv := &canonical.CanonicalInvoice{
		InvoiceNumber: "INV-001",
		IssueDate:     time.Now().UTC(),
		DueDate:       time.Now().UTC().Add(24 * time.Hour),
		Seller: canonical.Party{
			CountryCode:   "FR",
			LegalEntityID: "73204903600045",
			TaxID:         "FR12732049036",
		},
		MonetaryTotals: canonical.MonetaryTotals{
			NetHT:     10000,
			TaxAmount: 2000,
			GrossTTC:  12000,
		},
		TaxBreakdowns: []canonical.TaxBreakdown{
			{
				BaseHT:    10000,
				TaxAmount: 2000,
			},
		},
	}

	report := ValidateEN16931Strict(inv)
	if !report.Valid {
		t.Fatalf("Validation failed unexpectedly: %+v", report.Diagnostics)
	}
}
