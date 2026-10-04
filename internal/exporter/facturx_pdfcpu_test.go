package exporter

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"
)

// buildMinimalPDF genere un PDF 1 page valide, offsets xref corrects.
func buildMinimalPDF() []byte {
	var b bytes.Buffer
	b.WriteString("%PDF-1.7\n%\xE2\xE3\xCF\xD3\n")
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>",
	}
	offs := make([]int, 0, len(objs))
	for i, o := range objs {
		offs = append(offs, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, o := range offs {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)
	return b.Bytes()
}

// pdfpcpu encode les chaines en UTF-16BE (litterales ou hexadecimales) :
// une recherche ASCII brute de "factur-x.xml" echoue donc a tort.
func containsPDFString(pdf []byte, s string) bool {
	var u16 []byte
	for _, r := range s {
		u16 = append(u16, 0, byte(r))
	}
	bom := append([]byte{0xFE, 0xFF}, u16...)
	forms := [][]byte{
		[]byte(s),
		u16,
		bom,
		[]byte(strings.ToUpper(hex.EncodeToString(u16))),
		[]byte(strings.ToLower(hex.EncodeToString(u16))),
		[]byte(strings.ToUpper(hex.EncodeToString(bom))),
		[]byte(strings.ToLower(hex.EncodeToString(bom))),
	}
	for _, f := range forms {
		if bytes.Contains(pdf, f) {
			return true
		}
	}
	return false
}

func TestFacturXEnginePDFCPU_AttachmentAndXMP(t *testing.T) {
	engine := NewFacturXEnginePDFCPU("")
	xmlData := []byte(`<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"><rsm:ExchangedDocument><ram:ID>INV-999</ram:ID></rsm:ExchangedDocument></rsm:CrossIndustryInvoice>`)

	res, err := engine.BuildFacturXPDFA3(bytes.NewReader(buildMinimalPDF()), xmlData, "EN 16931")
	if err != nil {
		t.Fatalf("Failed to build Factur-X: %v", err)
	}
	if len(res) == 0 {
		t.Fatal("Resulting PDF is empty")
	}
	if !bytes.HasPrefix(res, []byte("%PDF-")) {
		t.Error("missing PDF header")
	}
	if !bytes.HasSuffix(bytes.TrimSpace(res), []byte("%%EOF")) {
		t.Errorf("missing %%%%EOF")
	}
	if !bytes.Contains(res, []byte("/EmbeddedFile")) {
		t.Error("no /EmbeddedFile stream")
	}
	if !containsPDFString(res, "factur-x.xml") {
		t.Error("Generated PDF does not declare factur-x.xml (ASCII/UTF-16/hex checked)")
	}
	if bytes.Contains(res, []byte("/Collection")) {
		t.Error("/Collection (portfolio) must not be present in PDF/A-3")
	}
	for _, want := range []string{"/AFRelationship/Alternative", "/AF[", "/Metadata ", "fx:ConformanceLevel>EN 16931<", "fx:DocumentFileName>factur-x.xml<", "/Prev ", "text#2Fxml"} {
		if !bytes.Contains(res, []byte(want)) {
			t.Errorf("missing %q in output", want)
		}
	}
}

func TestFacturXEnginePDFCPU_EmptyXML(t *testing.T) {
	engine := NewFacturXEnginePDFCPU("")
	if _, err := engine.BuildFacturXPDFA3(bytes.NewReader(buildMinimalPDF()), nil, ""); err == nil {
		t.Fatal("expected error for empty xml")
	}
}

func TestFacturXEnginePDFCPU_BadProfile(t *testing.T) {
	engine := NewFacturXEnginePDFCPU("")
	if _, err := engine.BuildFacturXPDFA3(bytes.NewReader(buildMinimalPDF()), []byte("<a/>"), "FOO"); err == nil {
		t.Fatal("expected error for unsupported profile")
	}
}

func TestNormalizeFacturXProfile(t *testing.T) {
	for in, want := range map[string]string{"": "EN 16931", "en16931": "EN 16931", " basic  wl ": "BASIC WL", "minimum": "MINIMUM", "Extended": "EXTENDED"} {
		got, err := normalizeFacturXProfile(in)
		if err != nil || got != want {
			t.Errorf("normalize(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

func TestFinalizeFacturX_RejectsGarbage(t *testing.T) {
	if _, err := finalizeFacturX([]byte("not a pdf"), "EN 16931", time.Now()); err == nil {
		t.Fatal("expected error on invalid input")
	}
}
