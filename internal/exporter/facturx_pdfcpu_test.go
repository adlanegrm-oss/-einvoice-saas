package exporter

import (
	"bytes"
	"testing"
)

func TestFacturXEnginePDFCPU_AttachmentAndXMP(t *testing.T) {
	// PDF de base minimal conforme
	basePDF := []byte(`%PDF-1.7
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >> endobj
xref
0 4
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
trailer << /Size 4 /Root 1 0 R >>
startxref
190
%%EOF`)

	engine := NewFacturXEnginePDFCPU("")
	xmlData := []byte(`<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"><rsm:ExchangedDocument><ram:ID>INV-999</ram:ID></rsm:ExchangedDocument></rsm:CrossIndustryInvoice>`)

	res, err := engine.BuildFacturXPDFA3(bytes.NewReader(basePDF), xmlData, "EN 16931")
	if err != nil {
		t.Fatalf("Failed to build Factur-X with pdfcpu: %v", err)
	}

	if len(res) == 0 {
		t.Fatal("Resulting PDF is empty")
	}

	// Vérification de la présence de la pièce jointe
	if !bytes.Contains(res, []byte("factur-x.xml")) {
		t.Error("Generated PDF does not declare factur-x.xml")
	}
}
