package facturx

import (
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/syntax"
)

func TestFacturXCoherence(t *testing.T) {
	inv := &canonical.CanonicalInvoice{
		InvoiceNumber: "FX-2026-001",
		Currency:      "EUR",
		IssueDate:     time.Now().UTC(),
		MonetaryTotals: canonical.MonetaryTotals{
			NetHT: 10000, TaxAmount: 2000, GrossTTC: 12000, PayableDue: 12000,
		},
	}

	xmlBytes, err := syntax.GenerateCIIXML(inv)
	if err != nil {
		t.Fatalf("generation CII echec: %v", err)
	}

	pdfBytes, err := GenerateMinimalPDFA3(inv, xmlBytes)
	if err != nil || len(pdfBytes) == 0 {
		t.Fatalf("generation PDF echec: %v", err)
	}

	summary := ExtractedPDFSummary{
		InvoiceNumber: "FX-2026-001",
		GrossTTC:      12000,
		Currency:      "EUR",
	}
	if err := VerifyBilateralCoherence(summary, inv); err != nil {
		t.Fatalf("coherence stricte en echec: %v", err)
	}

	badSummary := ExtractedPDFSummary{
		InvoiceNumber: "FX-2026-001",
		GrossTTC:      99999,
		Currency:      "EUR",
	}
	if err := VerifyBilateralCoherence(badSummary, inv); err == nil {
		t.Fatal("incoherence non detectee")
	}
}
