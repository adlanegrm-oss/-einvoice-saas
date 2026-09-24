package validator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain/models"
)

// ==========================================
// 1. VALIDATION DU MODÈLE MÉTIER (Ton code)
// ==========================================

// Validate vérifie la conformité de la structure Invoice
func Validate(invoice *models.Invoice) []error {
	var errors []error

	if invoice == nil {
		return []error{
			fmt.Errorf("invoice is nil"),
		}
	}

	if strings.TrimSpace(invoice.InvoiceNumber) == "" {
		errors = append(errors, fmt.Errorf("invoice number is required"))
	}

	if strings.TrimSpace(invoice.Seller.Name) == "" {
		errors = append(errors, fmt.Errorf("seller name is required"))
	}

	if strings.TrimSpace(invoice.Buyer.Name) == "" {
		errors = append(errors, fmt.Errorf("buyer name is required"))
	}

	if strings.TrimSpace(invoice.Currency) == "" {
		errors = append(errors, fmt.Errorf("currency is required"))
	}

	if invoice.Amounts.Total < 0 {
		errors = append(errors, fmt.Errorf("total amount cannot be negative"))
	}

	return errors
}

// ==========================================
// 2. VALIDATION DES FICHIERS (EDI, XML, PDF)
// ==========================================

type FormatType string

const (
	FormatFacturX   FormatType = "FACTUR-X"
	FormatUBL       FormatType = "UBL"
	FormatEDIFACT   FormatType = "EDIFACT"
	FormatIDDoc     FormatType = "IDDOC"
	FormatSignedPDF FormatType = "SIGNED_PDF"
	FormatUnknown   FormatType = "UNKNOWN"
)

type FileValidationResult struct {
	Format   FormatType        `json:"format"`
	IsValid  bool              `json:"is_valid"`
	Errors   []string          `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
	Meta     map[string]string `json:"meta,omitempty"`
}

type FileValidator struct{}

func NewFileValidator() *FileValidator {
	return &FileValidator{}
}

// ValidateFileContent analyse le contenu binaire du fichier
func (v *FileValidator) ValidateFileContent(content []byte) FileValidationResult {
	res := FileValidationResult{
		IsValid: false,
		Meta:    make(map[string]string),
	}

	if len(content) == 0 {
		res.Format = FormatUnknown
		res.Errors = append(res.Errors, "Le fichier est vide")
		return res
	}

	trimmed := bytes.TrimSpace(content)

	// Détection PDF / PDF Signé
	if bytes.HasPrefix(trimmed, []byte("%PDF-")) {
		return v.validatePDF(trimmed)
	}

	// Détection EDIFACT
	if bytes.HasPrefix(trimmed, []byte("UNA")) || bytes.HasPrefix(trimmed, []byte("UNB")) {
		return v.validateEDIFACT(trimmed)
	}

	// Détection XML (Factur-X, UBL, IDDoc)
	if bytes.HasPrefix(trimmed, []byte("<")) || strings.Contains(string(trimmed[:min(100, len(trimmed))]), "<?xml") {
		return v.validateXML(trimmed)
	}

	res.Format = FormatUnknown
	res.Errors = append(res.Errors, "Format de fichier inconnu")
	return res
}

func (v *FileValidator) validateXML(content []byte) FileValidationResult {
	res := FileValidationResult{IsValid: true, Meta: make(map[string]string)}
	decoder := xml.NewDecoder(bytes.NewReader(content))

	var rootName string
	var rootXMLNS string

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() != "EOF" {
				res.IsValid = false
				res.Errors = append(res.Errors, fmt.Sprintf("XML invalide : %v", err))
			}
			break
		}

		if se, ok := token.(xml.StartElement); ok {
			rootName = se.Name.Local
			for _, attr := range se.Attr {
				if attr.Name.Local == "xmlns" || strings.HasPrefix(attr.Name.Local, "xmlns:") {
					rootXMLNS += attr.Value + " "
				}
			}
			break
		}
	}

	if !res.IsValid {
		res.Format = FormatUnknown
		return res
	}

	res.Meta["root_element"] = rootName

	switch {
	case rootName == "CrossIndustryInvoice" || strings.Contains(rootXMLNS, "CrossIndustryInvoice"):
		res.Format = FormatFacturX

	case rootName == "Invoice" || rootName == "CreditNote" || strings.Contains(rootXMLNS, "specification:ubl"):
		res.Format = FormatUBL

	case rootName == "IDDoc" || strings.Contains(strings.ToLower(rootName), "iddoc") || strings.Contains(rootXMLNS, "iddoc"):
		res.Format = FormatIDDoc

	default:
		res.Format = FormatUnknown
		res.IsValid = false
		res.Errors = append(res.Errors, fmt.Sprintf("Schéma XML non supporté: %s", rootName))
	}

	return res
}

func (v *FileValidator) validateEDIFACT(content []byte) FileValidationResult {
	res := FileValidationResult{Format: FormatEDIFACT, IsValid: true, Meta: make(map[string]string)}
	str := string(content)

	if !strings.Contains(str, "UNH") || !strings.Contains(str, "UNT") {
		res.IsValid = false
		res.Errors = append(res.Errors, "Structure de segments EDIFACT manquante (UNH/UNT)")
	}

	return res
}

func (v *FileValidator) validatePDF(content []byte) FileValidationResult {
	res := FileValidationResult{Format: FormatSignedPDF, IsValid: true, Meta: make(map[string]string)}
	str := string(content)

	if strings.Contains(str, "/Type /Sig") || strings.Contains(str, "/Contents") {
		res.Meta["is_signed"] = "true"
	} else {
		res.Meta["is_signed"] = "false"
		res.Warnings = append(res.Warnings, "PDF non signé électroniquement")
	}

	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
