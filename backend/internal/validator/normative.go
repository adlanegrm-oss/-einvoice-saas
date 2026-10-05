package validator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// SVRLSeverity représente la sévérité officielle d'une assertion
type SVRLSeverity string

const (
	SeverityFatal   SVRLSeverity = "fatal"
	SeverityWarning SVRLSeverity = "warning"
)

// SVRLError structure officielle retournée en cas de violation d'une règle EN 16931 ou PEPPOL
type SVRLError struct {
	RuleID   string       `json:"rule_id"`
	Severity SVRLSeverity `json:"severity"`
	Context  string       `json:"context"`
	Message  string       `json:"message"`
	Location string       `json:"location"`
}

// ValidationResult représente le rapport de conformité complet
type ValidationResult struct {
	Valid      bool        `json:"valid"`
	RuleErrors []SVRLError `json:"errors,omitempty"`
	Warnings   []SVRLError `json:"warnings,omitempty"`
}

var (
	sirenRegex = regexp.MustCompile(`^[0-9]{9}$`)
	siretRegex = regexp.MustCompile(`^[0-9]{14}$`)
	vatFRRegex = regexp.MustCompile(`^FR[0-9A-Z]{2}[0-9]{9}$`)
	peppolIDRe = regexp.MustCompile(`^[0-9]{4}:.+`)
)

type NormativeValidator struct {
	strictPeppol bool
}

func NewNormativeValidator(strictPeppol bool) *NormativeValidator {
	return &NormativeValidator{strictPeppol: strictPeppol}
}

type parsedDoc struct {
	GuidelineID    string
	InvoiceID      string
	TypeCode       string
	IssueDate      string
	DueDate        string
	Currency       string
	BuyerReference string
	SellerName     string
	SellerCountry  string
	SellerLegalID  string
	SellerTaxID    string
	SellerEndpoint string
	BuyerName      string
	BuyerCountry   string
	BuyerLegalID   string
	BuyerTaxID     string
	BuyerEndpoint  string
	LineTotalAmt   float64
	TaxBasisTotal  float64
	TaxTotalAmt    float64
	GrandTotalAmt  float64
	PayableDueAmt  float64
	Lines          []parsedDocLine
	Taxes          []parsedDocTax
}

type parsedDocLine struct {
	ID        string
	Qty       float64
	UnitPrice float64
	TotalNet  float64
	VatCat    string
	VatRate   float64
}

type parsedDocTax struct {
	Category string
	Rate     float64
	Basis    float64
	Amount   float64
}

// ValidateEN16931AndPeppol valide le document XML contre les règles normatives
func (v *NormativeValidator) ValidateEN16931AndPeppol(xmlContent []byte) (*ValidationResult, error) {
	doc, err := v.parseXML(xmlContent)
	if err != nil {
		return &ValidationResult{
			Valid: false,
			RuleErrors: []SVRLError{{
				RuleID:   "XML-NOT-WELL-FORMED",
				Severity: SeverityFatal,
				Location: "/Invoice",
				Message:  fmt.Sprintf("Le document XML est mal formé : %v", err),
			}},
		}, nil
	}

	var errors []SVRLError
	var warnings []SVRLError

	// BR-01: Profile / Specification Identifier
	if doc.GuidelineID == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-01",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocumentContext/ram:GuidelineSpecifiedDocumentContextParameter/ram:ID",
			Message:  "Une facture doit spécifier un identifiant de spécification conforme (Profile URN).",
		})
	}

	// BR-02: Numéro de facture
	if doc.InvoiceID == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-02",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:ID",
			Message:  "Une facture doit obligatoirement comporter un numéro de facture (Invoice Identifier).",
		})
	}

	// BR-03: Date d'émission
	if doc.IssueDate == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-03",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:IssueDateTime",
			Message:  "Une facture doit comporter une date d'émission.",
		})
	}

	// BR-04: Type de facture
	if doc.TypeCode == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-04",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:TypeCode",
			Message:  "Une facture doit comporter un code de type de document (ex: 380, 381).",
		})
	}

	// BR-05: Devise
	if doc.Currency == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-05",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:InvoiceCurrencyCode",
			Message:  "Le code devise de facturation est obligatoire (ISO 4217).",
		})
	}

	// BR-06 & BR-08: Vendeur (Nom & Pays)
	if doc.SellerName == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-06",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:Name",
			Message:  "Le nom du vendeur est obligatoire.",
		})
	}
	if doc.SellerCountry == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-08",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:PostalTradeAddress/ram:CountryID",
			Message:  "L'adresse du vendeur doit contenir un code pays ISO 3166-1 alpha-2.",
		})
	}

	// BR-07 & BR-09: Acheteur (Nom & Pays)
	if doc.BuyerName == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-07",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty/ram:Name",
			Message:  "Le nom de l'acheteur est obligatoire.",
		})
	}
	if doc.BuyerCountry == "" {
		errors = append(errors, SVRLError{
			RuleID:   "BR-09",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty/ram:PostalTradeAddress/ram:CountryID",
			Message:  "L'adresse de l'acheteur doit contenir un code pays.",
		})
	}

	// BR-16: Au moins une ligne
	if len(doc.Lines) == 0 {
		errors = append(errors, SVRLError{
			RuleID:   "BR-16",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:IncludedSupplyChainTradeLineItem",
			Message:  "Une facture doit obligatoirement comporter au moins une ligne d'article.",
		})
	}

	// BR-CO-10: Total des lignes == LineTotalAmount
	var calcSumLines float64
	for i, l := range doc.Lines {
		if l.ID == "" {
			errors = append(errors, SVRLError{
				RuleID:   "BR-21",
				Severity: SeverityFatal,
				Location: fmt.Sprintf("/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:IncludedSupplyChainTradeLineItem[%d]/ram:AssociatedDocumentLineDocument/ram:LineID", i+1),
				Message:  fmt.Sprintf("La ligne de facture %d doit comporter un identifiant.", i+1),
			})
		}
		calcSumLines += l.TotalNet
	}
	if math.Abs(calcSumLines-doc.LineTotalAmt) > 0.02 {
		errors = append(errors, SVRLError{
			RuleID:   "BR-CO-10",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:LineTotalAmount",
			Message:  fmt.Sprintf("La somme des montants nets des lignes (%.2f) doit être égale à LineTotalAmount (%.2f).", calcSumLines, doc.LineTotalAmt),
		})
	}

	// BR-CO-13 & BR-CO-14: Assiette TVA et Montant TVA
	var sumBases, sumTaxes float64
	for _, t := range doc.Taxes {
		sumBases += t.Basis
		sumTaxes += t.Amount
	}
	if math.Abs(sumBases-doc.TaxBasisTotal) > 0.02 {
		errors = append(errors, SVRLError{
			RuleID:   "BR-CO-13",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:TaxBasisTotalAmount",
			Message:  fmt.Sprintf("La somme des assiettes fiscales (%.2f) doit correspondre au total HT (%.2f).", sumBases, doc.TaxBasisTotal),
		})
	}
	if math.Abs(sumTaxes-doc.TaxTotalAmt) > 0.02 {
		errors = append(errors, SVRLError{
			RuleID:   "BR-CO-14",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:TaxTotalAmount",
			Message:  fmt.Sprintf("La somme des taxes par catégorie (%.2f) doit correspondre au montant total de la taxe (%.2f).", sumTaxes, doc.TaxTotalAmt),
		})
	}

	// BR-CO-15: GrandTotalAmount = TaxBasisTotal + TaxTotal
	expectedGrand := doc.TaxBasisTotal + doc.TaxTotalAmt
	if math.Abs(doc.GrandTotalAmt-expectedGrand) > 0.02 {
		errors = append(errors, SVRLError{
			RuleID:   "BR-CO-15",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:GrandTotalAmount",
			Message:  fmt.Sprintf("Le montant total TTC (%.2f) doit égaler Net HT (%.2f) + Taxe (%.2f).", doc.GrandTotalAmt, doc.TaxBasisTotal, doc.TaxTotalAmt),
		})
	}

	// BR-CO-25: DueDate >= IssueDate
	if doc.IssueDate != "" && doc.DueDate != "" && doc.DueDate < doc.IssueDate {
		errors = append(errors, SVRLError{
			RuleID:   "BR-CO-25",
			Severity: SeverityFatal,
			Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradePaymentTerms/ram:DueDateDateTime",
			Message:  fmt.Sprintf("La date d'échéance (%s) ne peut pas être antérieure à la date d'émission (%s).", doc.DueDate, doc.IssueDate),
		})
	}

	// Contrôles spécifiques France (CIUS-FR)
	if doc.SellerCountry == "FR" {
		cleanSellerID := strings.ReplaceAll(doc.SellerLegalID, " ", "")
		if cleanSellerID != "" && !sirenRegex.MatchString(cleanSellerID) && !siretRegex.MatchString(cleanSellerID) {
			errors = append(errors, SVRLError{
				RuleID:   "CIUS-FR-01",
				Severity: SeverityFatal,
				Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:SpecifiedLegalOrganization/ram:ID",
				Message:  "L'identifiant légal du vendeur en France doit être un numéro SIREN (9 chiffres) ou SIRET (14 chiffres).",
			})
		}
		cleanVAT := strings.ReplaceAll(doc.SellerTaxID, " ", "")
		if cleanVAT != "" && !vatFRRegex.MatchString(cleanVAT) {
			errors = append(errors, SVRLError{
				RuleID:   "CIUS-FR-02",
				Severity: SeverityFatal,
				Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:SpecifiedTaxRegistration/ram:ID",
				Message:  "Le numéro de TVA intracommunautaire français doit respecter le format FR + 2 caractères + 9 chiffres.",
			})
		}
	}

	// Contrôles PEPPOL BIS v3
	if v.strictPeppol {
		if doc.SellerEndpoint == "" {
			errors = append(errors, SVRLError{
				RuleID:   "PEPPOL-EN16931-R001",
				Severity: SeverityFatal,
				Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:URIUniversalCommunication",
				Message:  "L'adresse électronique Peppol (EndpointID) du vendeur est obligatoire dans le profil Peppol BIS.",
			})
		}
		if doc.BuyerEndpoint == "" {
			errors = append(errors, SVRLError{
				RuleID:   "PEPPOL-EN16931-R002",
				Severity: SeverityFatal,
				Location: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty/ram:URIUniversalCommunication",
				Message:  "L'adresse électronique Peppol (EndpointID) de l'acheteur est obligatoire dans le profil Peppol BIS.",
			})
		}
	}

	return &ValidationResult{
		Valid:      len(errors) == 0,
		RuleErrors: errors,
		Warnings:   warnings,
	}, nil
}

func (v *NormativeValidator) parseXML(xmlData []byte) (*parsedDoc, error) {
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	var doc parsedDoc
	var stack []string
	var textBuf strings.Builder
	var curLine parsedDocLine
	var inLine bool
	var curTax parsedDocTax
	var inHeaderTax bool

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch el := tok.(type) {
		case xml.StartElement:
			stack = append(stack, el.Name.Local)
			textBuf.Reset()

			if el.Name.Local == "IncludedSupplyChainTradeLineItem" {
				inLine = true
				curLine = parsedDocLine{}
			} else if el.Name.Local == "ApplicableTradeTax" && len(stack) >= 2 && stack[len(stack)-2] == "ApplicableHeaderTradeSettlement" {
				inHeaderTax = true
				curTax = parsedDocTax{}
			}

		case xml.CharData:
			textBuf.Write(el)

		case xml.EndElement:
			val := strings.TrimSpace(textBuf.String())
			path := strings.Join(stack, "/")
			fVal, _ := strconv.ParseFloat(val, 64)

			if val != "" {
				switch {
				case strings.HasSuffix(path, "GuidelineSpecifiedDocumentContextParameter/ID"):
					doc.GuidelineID = val
				case strings.HasSuffix(path, "ExchangedDocument/ID"):
					doc.InvoiceID = val
				case strings.HasSuffix(path, "ExchangedDocument/TypeCode"):
					doc.TypeCode = val
				case strings.HasSuffix(path, "IssueDateTime/DateTimeString"):
					doc.IssueDate = val
				case strings.HasSuffix(path, "DueDateDateTime/DateTimeString"):
					doc.DueDate = val
				case strings.HasSuffix(path, "ApplicableHeaderTradeAgreement/BuyerReference"):
					doc.BuyerReference = val
				case strings.HasSuffix(path, "ApplicableHeaderTradeSettlement/InvoiceCurrencyCode"):
					doc.Currency = val
				case strings.HasSuffix(path, "SellerTradeParty/Name"):
					doc.SellerName = val
				case strings.HasSuffix(path, "SellerTradeParty/PostalTradeAddress/CountryID"):
					doc.SellerCountry = val
				case strings.HasSuffix(path, "SellerTradeParty/SpecifiedLegalOrganization/ID"):
					doc.SellerLegalID = val
				case strings.HasSuffix(path, "SellerTradeParty/SpecifiedTaxRegistration/ID"):
					doc.SellerTaxID = val
				case strings.HasSuffix(path, "SellerTradeParty/URIUniversalCommunication/URIID"):
					doc.SellerEndpoint = val
				case strings.HasSuffix(path, "BuyerTradeParty/Name"):
					doc.BuyerName = val
				case strings.HasSuffix(path, "BuyerTradeParty/PostalTradeAddress/CountryID"):
					doc.BuyerCountry = val
				case strings.HasSuffix(path, "BuyerTradeParty/SpecifiedLegalOrganization/ID"):
					doc.BuyerLegalID = val
				case strings.HasSuffix(path, "BuyerTradeParty/SpecifiedTaxRegistration/ID"):
					doc.BuyerTaxID = val
				case strings.HasSuffix(path, "BuyerTradeParty/URIUniversalCommunication/URIID"):
					doc.BuyerEndpoint = val
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/LineTotalAmount"):
					doc.LineTotalAmt = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/TaxBasisTotalAmount"):
					doc.TaxBasisTotal = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/TaxTotalAmount"):
					doc.TaxTotalAmt = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/GrandTotalAmount"):
					doc.GrandTotalAmt = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/DuePayableAmount"):
					doc.PayableDueAmt = fVal
				case inLine && strings.HasSuffix(path, "AssociatedDocumentLineDocument/LineID"):
					curLine.ID = val
				case inLine && strings.HasSuffix(path, "SpecifiedTradeSettlementLineMonetarySummation/LineTotalAmount"):
					curLine.TotalNet = fVal
				case inHeaderTax && strings.HasSuffix(path, "ApplicableTradeTax/BasisAmount"):
					curTax.Basis = fVal
				case inHeaderTax && strings.HasSuffix(path, "ApplicableTradeTax/CalculatedAmount"):
					curTax.Amount = fVal
				}
			}

			if el.Name.Local == "IncludedSupplyChainTradeLineItem" {
				doc.Lines = append(doc.Lines, curLine)
				inLine = false
			} else if el.Name.Local == "ApplicableTradeTax" && inHeaderTax {
				doc.Taxes = append(doc.Taxes, curTax)
				inHeaderTax = false
			}

			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			textBuf.Reset()
		}
	}

	return &doc, nil
}
