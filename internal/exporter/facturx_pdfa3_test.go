package exporter

import (
	"testing"
	"time"
)

func TestGenerateFacturXPDFA3_Conformity(t *testing.T) {
	meta := InvoiceMetadata{
		InvoiceNumber: "FX-2026-999",
		SellerName:    "Cegedim Services",
		SellerSIREN:   "732049036",
		SellerVAT:     "FR12732049036",
		BuyerName:     "Odigo France",
		IssueDate:     time.Now().UTC(),
		Currency:      "EUR",
		TotalHT:       1500.00,
		TotalTax:      300.00,
		TotalTTC:      1800.00,
		Profile:       ProfileEN16931,
	}

	ciiPayload := []byte(`<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"><rsm:ExchangedDocument><ram:ID>FX-2026-999</ram:ID></rsm:ExchangedDocument></rsm:CrossIndustryInvoice>`)

	pdfBytes, err := GenerateFacturXPDFA3(meta, ciiPayload)
	if err != nil {
		t.Fatalf("Failed to generate PDF/A-3: %v", err)
	}

	if len(pdfBytes) < 500 {
		t.Fatalf("Generated PDF is suspiciously small: %d bytes", len(pdfBytes))
	}

	if err := VerifyPDFA3Conformance(pdfBytes); err != nil {
		t.Fatalf("PDF/A-3 Factur-X verification failed: %v", err)
	}
}
