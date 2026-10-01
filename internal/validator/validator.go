package validator

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain/models"
	"github.com/adlanegrm-oss/einvoice-saas/internal/parser/ubl"
)

// ==========================================
// 1. VALIDATION DU MODÈLE MÉTIER
// ==========================================

// Validate vérifie la conformité de la structure Invoice
func Validate(invoice *models.Invoice) []error {
	var errs []error

	if invoice == nil {
		return []error{fmt.Errorf("invoice is nil")}
	}

	if strings.TrimSpace(invoice.InvoiceNumber) == "" {
		errs = append(errs, fmt.Errorf("invoice number is required"))
	}
	if strings.TrimSpace(invoice.Seller.Name) == "" {
		errs = append(errs, fmt.Errorf("seller name is required"))
	}
	if strings.TrimSpace(invoice.Buyer.Name) == "" {
		errs = append(errs, fmt.Errorf("buyer name is required"))
	}
	if strings.TrimSpace(invoice.Currency) == "" {
		errs = append(errs, fmt.Errorf("currency is required"))
	}
	if invoice.Amounts.Total < 0 {
		errs = append(errs, fmt.Errorf("total amount cannot be negative"))
	}

	return errs
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
	FormatPDF       FormatType = "PDF"
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

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// ValidateFileContent analyse le contenu binaire du fichier.
//
// Périmètre actuel : détection du format, bonne formation du document et
// contrôle des champs essentiels. La validation complète contre les schémas
// XSD / règles Schematron (EN 16931, Factur-X, PEPPOL BIS) reste à brancher.
func (v *FileValidator) ValidateFileContent(content []byte) FileValidationResult {
	res := FileValidationResult{
		IsValid: false,
		Meta:    make(map[string]string),
	}

	trimmed := bytes.TrimSpace(bytes.TrimPrefix(content, utf8BOM))
	if len(trimmed) == 0 {
		res.Format = FormatUnknown
		res.Errors = append(res.Errors, "Le fichier est vide")
		return res
	}

	switch {
	case bytes.HasPrefix(trimmed, []byte("%PDF-")):
		return v.validatePDF(trimmed)
	case bytes.HasPrefix(trimmed, []byte("UNA")) || bytes.HasPrefix(trimmed, []byte("UNB")):
		return v.validateEDIFACT(trimmed)
	case bytes.HasPrefix(trimmed, []byte("<")):
		return v.validateXML(trimmed)
	}

	res.Format = FormatUnknown
	res.Errors = append(res.Errors, "Format de fichier inconnu")
	return res
}

// ---------- XML ----------

type ciiDoc struct {
	XMLName  xml.Name `xml:"CrossIndustryInvoice"`
	ID       string   `xml:"ExchangedDocument>ID"`
	Seller   string   `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeAgreement>SellerTradeParty>Name"`
	Buyer    string   `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeAgreement>BuyerTradeParty>Name"`
	Currency string   `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeSettlement>InvoiceCurrencyCode"`
}

func (v *FileValidator) validateXML(content []byte) FileValidationResult {
	res := FileValidationResult{IsValid: true, Meta: make(map[string]string)}

	// Lecture COMPLÈTE du document : une balise mal fermée ou un fichier tronqué
	// après l'élément racine doit être détecté.
	decoder := xml.NewDecoder(bytes.NewReader(content))
	var root xml.Name
	seenRoot := false
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			res.IsValid = false
			res.Format = FormatUnknown
			res.Errors = append(res.Errors, fmt.Sprintf("XML invalide : %v", err))
			return res
		}
		switch t := token.(type) {
		case xml.StartElement:
			if !seenRoot {
				seenRoot = true
				root = t.Name
			}
		case xml.Directive:
			// Les DTD sont proscrites dans les factures électroniques (et source d'attaques).
			if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(string(t))), "DOCTYPE") {
				res.IsValid = false
				res.Format = FormatUnknown
				res.Errors = append(res.Errors, "Déclaration DOCTYPE interdite")
				return res
			}
		}
	}
	if !seenRoot {
		res.IsValid = false
		res.Format = FormatUnknown
		res.Errors = append(res.Errors, "Aucun élément racine XML")
		return res
	}

	res.Meta["root_element"] = root.Local
	lowerRoot := strings.ToLower(root.Local)

	switch {
	case root.Local == "CrossIndustryInvoice" || strings.Contains(root.Space, "CrossIndustryInvoice"):
		res.Format = FormatFacturX
		v.checkCII(content, &res)

	case root.Local == "Invoice" || root.Local == "CreditNote" || strings.Contains(root.Space, "specification:ubl"):
		res.Format = FormatUBL
		v.checkUBL(content, root.Local, &res)

	case strings.Contains(lowerRoot, "iddoc") || strings.Contains(strings.ToLower(root.Space), "iddoc"):
		res.Format = FormatIDDoc

	default:
		res.Format = FormatUnknown
		res.IsValid = false
		res.Errors = append(res.Errors, fmt.Sprintf("Schéma XML non supporté: %s", root.Local))
	}

	return res
}

func (v *FileValidator) checkUBL(content []byte, rootLocal string, res *FileValidationResult) {
	if rootLocal != "Invoice" {
		res.Warnings = append(res.Warnings, "Contrôle détaillé non disponible pour ce type de document UBL ("+rootLocal+")")
		return
	}
	inv, err := ubl.Parse(content)
	if err != nil {
		res.IsValid = false
		res.Errors = append(res.Errors, err.Error())
		return
	}
	res.Meta["invoice_number"] = inv.InvoiceNumber
	res.Meta["currency"] = inv.Currency

	missing := func(value, label string) {
		if strings.TrimSpace(value) == "" {
			res.IsValid = false
			res.Errors = append(res.Errors, label+" manquant")
		}
	}
	missing(inv.InvoiceNumber, "Numéro de facture (cbc:ID)")
	missing(inv.Currency, "Devise (cbc:DocumentCurrencyCode)")
	missing(inv.Seller.Name, "Nom du vendeur (AccountingSupplierParty)")
	missing(inv.Buyer.Name, "Nom de l'acheteur (AccountingCustomerParty)")
}

func (v *FileValidator) checkCII(content []byte, res *FileValidationResult) {
	var doc ciiDoc
	if err := xml.Unmarshal(content, &doc); err != nil {
		res.IsValid = false
		res.Errors = append(res.Errors, fmt.Sprintf("Structure CII illisible : %v", err))
		return
	}
	res.Meta["invoice_number"] = strings.TrimSpace(doc.ID)

	if strings.TrimSpace(doc.ID) == "" {
		res.IsValid = false
		res.Errors = append(res.Errors, "Numéro de facture (ExchangedDocument/ID) manquant")
	}
	if strings.TrimSpace(doc.Buyer) == "" {
		res.IsValid = false
		res.Errors = append(res.Errors, "Nom de l'acheteur (BuyerTradeParty/Name) manquant")
	}
	if strings.TrimSpace(doc.Seller) == "" {
		res.Warnings = append(res.Warnings, "Nom du vendeur (SellerTradeParty/Name) absent : requis par le profil Factur-X MINIMUM")
	}
	if strings.TrimSpace(doc.Currency) == "" {
		res.Warnings = append(res.Warnings, "Devise (InvoiceCurrencyCode) absente : requise par le profil Factur-X MINIMUM")
	}
}

// ---------- EDIFACT ----------

// splitEDIFACT découpe un échange en segments en respectant la chaîne de service
// UNA (séparateurs personnalisés) et le caractère d'échappement.
func splitEDIFACT(s string) (segments []string, elem string, unterminated bool) {
	elem = "+"
	release, term := byte('?'), byte('\'')
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "UNA") && len(s) >= 9 {
		elem = string(s[4])
		release = s[6]
		term = s[8]
		s = s[9:]
	}

	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == release && i+1 < len(s):
			cur.WriteByte(c)
			cur.WriteByte(s[i+1])
			i++
		case c == term:
			if seg := strings.TrimSpace(cur.String()); seg != "" {
				segments = append(segments, seg)
			}
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	if rest := strings.TrimSpace(cur.String()); rest != "" {
		segments = append(segments, rest)
		unterminated = true
	}
	return segments, elem, unterminated
}

func (v *FileValidator) validateEDIFACT(content []byte) FileValidationResult {
	res := FileValidationResult{Format: FormatEDIFACT, IsValid: true, Meta: make(map[string]string)}
	fail := func(msg string) {
		res.IsValid = false
		res.Errors = append(res.Errors, msg)
	}

	segments, elem, unterminated := splitEDIFACT(string(content))
	if unterminated {
		fail("Dernier segment EDIFACT non terminé")
	}
	if len(segments) == 0 {
		fail("Aucun segment EDIFACT")
		return res
	}

	fields := func(seg string) []string { return strings.Split(seg, elem) }

	unh, unt := -1, -1
	hasBGM, hasUNZ := false, false
	for i, seg := range segments {
		f := fields(seg)
		switch f[0] {
		case "UNH":
			if unh == -1 {
				unh = i
				if len(f) > 2 {
					res.Meta["message_type"] = f[2]
					if !strings.HasPrefix(f[2], "INVOIC") {
						res.Warnings = append(res.Warnings, "Type de message inattendu : "+f[2]+" (INVOIC attendu)")
					}
				}
			}
		case "UNT":
			unt = i
		case "BGM":
			hasBGM = true
			if len(f) > 2 {
				res.Meta["invoice_number"] = f[2]
			}
		case "UNZ":
			hasUNZ = true
		}
	}

	if f := fields(segments[0]); f[0] != "UNB" {
		fail("L'échange doit commencer par un segment UNB")
	}
	if unh == -1 || unt == -1 {
		fail("Structure de segments EDIFACT manquante (UNH/UNT)")
	} else {
		f := fields(segments[unt])
		if len(f) > 1 {
			if n, err := strconv.Atoi(f[1]); err != nil || n != unt-unh+1 {
				fail(fmt.Sprintf("UNT : nombre de segments déclaré (%s) différent du nombre réel (%d)", f[1], unt-unh+1))
			}
		}
		if len(f) > 2 {
			if h := fields(segments[unh]); len(h) > 1 && h[1] != f[2] {
				fail("UNT : référence de message différente de celle du UNH")
			}
		}
	}
	if !hasBGM {
		fail("Segment BGM (début de message) manquant")
	}
	if !hasUNZ {
		fail("Segment UNZ (fin d'échange) manquant")
	}

	return res
}

// ---------- PDF ----------

var facturxAttachmentNames = []string{"factur-x.xml", "zugferd-invoice.xml", "xrechnung.xml"}

func (v *FileValidator) validatePDF(content []byte) FileValidationResult {
	res := FileValidationResult{Format: FormatPDF, IsValid: true, Meta: make(map[string]string)}

	if end := bytes.IndexByte(content, '\n'); end > 0 && end <= 16 {
		res.Meta["pdf_version"] = strings.TrimSpace(strings.TrimPrefix(string(content[:end]), "%PDF-"))
	}

	tail := content[max(0, len(content)-2048):]
	if !bytes.Contains(tail, []byte("%%EOF")) {
		res.IsValid = false
		res.Errors = append(res.Errors, "PDF tronqué ou corrompu (marqueur %%EOF absent)")
		return res
	}

	// Un dictionnaire de signature contient toujours /ByteRange. (Chercher "/Contents"
	// ou "/Sig" seul donnerait des faux positifs : ils apparaissent dans tout PDF.)
	signed := bytes.Contains(content, []byte("/ByteRange"))
	lower := bytes.ToLower(content)
	hybrid := false
	for _, name := range facturxAttachmentNames {
		if bytes.Contains(lower, []byte(name)) {
			hybrid = true
			res.Meta["embedded_xml"] = name
			break
		}
	}

	switch {
	case hybrid:
		res.Format = FormatFacturX
	case signed:
		res.Format = FormatSignedPDF
	}

	res.Meta["is_signed"] = strconv.FormatBool(signed)
	if !signed {
		res.Warnings = append(res.Warnings, "PDF non signé électroniquement")
	}
	if !hybrid {
		res.Warnings = append(res.Warnings, "Aucun XML Factur-X embarqué : ce PDF seul n'est pas une facture électronique structurée")
	}

	return res
}
