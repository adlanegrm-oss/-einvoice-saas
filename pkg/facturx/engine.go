package facturx

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

var ErrCoherenceMismatch = errors.New("incoherence stricte entre metadonnees PDF et flux XML CII")

type ExtractedPDFSummary struct {
	InvoiceNumber string
	GrossTTC      int64
	Currency      string
}

func GenerateMinimalPDFA3(inv *canonical.CanonicalInvoice, ciiXML []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.7\n%\xE2\xE3\xCF\xD3\n")
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R /AF [4 0 R] >>\nendobj\n")
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] >>\nendobj\n")

	buf.WriteString(fmt.Sprintf("4 0 obj\n<< /Type /Filespec /F (factur-x.xml) /UF (factur-x.xml) /EF << /F 5 0 R >> /AFRelationship /Data >>\nendobj\n"))
	buf.WriteString(fmt.Sprintf("5 0 obj\n<< /Type /EmbeddedFile /Subtype /text#2Fxml /Length %d >>\nstream\n", len(ciiXML)))
	buf.Write(ciiXML)
	buf.WriteString("\nendstream\nendobj\n")

	buf.WriteString("xref\n0 6\n0000000000 65535 f \n0000000015 00000 n \n0000000078 00000 n \n0000000135 00000 n \n0000000206 00000 n \n0000000318 00000 n \n")
	buf.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n500\n%%EOF\n")

	return buf.Bytes(), nil
}

func VerifyBilateralCoherence(pdfData ExtractedPDFSummary, inv *canonical.CanonicalInvoice) error {
	if strings.TrimSpace(pdfData.InvoiceNumber) != strings.TrimSpace(inv.InvoiceNumber) {
		return fmt.Errorf("%w: numero facture PDF (%s) != XML (%s)", ErrCoherenceMismatch, pdfData.InvoiceNumber, inv.InvoiceNumber)
	}
	if pdfData.GrossTTC != inv.MonetaryTotals.GrossTTC {
		return fmt.Errorf("%w: total TTC PDF (%d) != XML (%d)", ErrCoherenceMismatch, pdfData.GrossTTC, inv.MonetaryTotals.GrossTTC)
	}
	if pdfData.Currency != "" && inv.Currency != "" && pdfData.Currency != inv.Currency {
		return fmt.Errorf("%w: devise PDF (%s) != XML (%s)", ErrCoherenceMismatch, pdfData.Currency, inv.Currency)
	}
	return nil
}
