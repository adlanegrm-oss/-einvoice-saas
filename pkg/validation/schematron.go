package validation

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

type SchematronSeverity string

const (
	SchematronError   SchematronSeverity = "ERROR"
	SchematronWarning SchematronSeverity = "WARNING"
)

type SchematronDiagnostic struct {
	RuleID   string             `json:"rule_id"`
	Severity SchematronSeverity `json:"severity"`
	Message  string             `json:"message"`
	Path     string             `json:"path"`
}

type SchematronReport struct {
	Valid       bool                   `json:"valid"`
	Diagnostics []SchematronDiagnostic `json:"diagnostics,omitempty"`
}

var (
	sirenRegex = regexp.MustCompile(`^[0-9]{9}$`)
	siretRegex = regexp.MustCompile(`^[0-9]{14}$`)
	vatFRRegex = regexp.MustCompile(`^FR[0-9A-Z]{2}[0-9]{9}$`)
)

type SchematronEngine struct {
	strictCIUSFR bool
}

func NewSchematronEngine(strictCIUSFR bool) *SchematronEngine {
	return &SchematronEngine{strictCIUSFR: strictCIUSFR}
}

type parsedLine struct {
	ID    string
	Total float64
}

type parsedTax struct {
	BasisAmount float64
	TaxAmount   float64
}

func (e *SchematronEngine) ValidateXML(xmlData []byte) (SchematronReport, error) {
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))

	var (
		guidelineID string
		invoiceID   string
		typeCode    string
		issueDate   string
		dueDate     string
		currency    string
		buyerRef    string

		sellerName    string
		sellerCountry string
		sellerSIRET   string
		sellerVAT     string

		buyerName    string
		buyerCountry string
		buyerSIRET   string
		buyerVAT     string

		lineTotalAmount float64
		taxBasisTotal   float64
		taxTotalAmount  float64
		grandTotal      float64
		payableAmount   float64

		lines []parsedLine
		taxes []parsedTax

		stack   []string
		textBuf strings.Builder

		curLine     parsedLine
		inLine      bool
		curTax      parsedTax
		inHeaderTax bool
	)

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return SchematronReport{
				Valid: false,
				Diagnostics: []SchematronDiagnostic{{
					RuleID:   "XML-SYNTAX-ERROR",
					Severity: SchematronError,
					Message:  fmt.Sprintf("Erreur syntaxe XML: %v", err),
					Path:     "/rsm:CrossIndustryInvoice",
				}},
			}, nil
		}

		switch el := tok.(type) {
		case xml.StartElement:
			stack = append(stack, el.Name.Local)
			textBuf.Reset()

			if el.Name.Local == "IncludedSupplyChainTradeLineItem" {
				inLine = true
				curLine = parsedLine{}
			} else if el.Name.Local == "ApplicableTradeTax" && len(stack) >= 2 && stack[len(stack)-2] == "ApplicableHeaderTradeSettlement" {
				inHeaderTax = true
				curTax = parsedTax{}
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
					guidelineID = val
				case strings.HasSuffix(path, "ExchangedDocument/ID"):
					invoiceID = val
				case strings.HasSuffix(path, "ExchangedDocument/TypeCode"):
					typeCode = val
				case strings.HasSuffix(path, "IssueDateTime/DateTimeString"):
					issueDate = val
				case strings.HasSuffix(path, "DueDateDateTime/DateTimeString"):
					dueDate = val
				case strings.HasSuffix(path, "ApplicableHeaderTradeAgreement/BuyerReference"):
					buyerRef = val
				case strings.HasSuffix(path, "ApplicableHeaderTradeSettlement/InvoiceCurrencyCode"):
					currency = val
				case strings.HasSuffix(path, "SellerTradeParty/Name"):
					sellerName = val
				case strings.HasSuffix(path, "SellerTradeParty/PostalTradeAddress/CountryID"):
					sellerCountry = val
				case strings.HasSuffix(path, "SellerTradeParty/SpecifiedLegalOrganization/ID"):
					sellerSIRET = val
				case strings.HasSuffix(path, "SellerTradeParty/SpecifiedTaxRegistration/ID"):
					sellerVAT = val
				case strings.HasSuffix(path, "BuyerTradeParty/Name"):
					buyerName = val
				case strings.HasSuffix(path, "BuyerTradeParty/PostalTradeAddress/CountryID"):
					buyerCountry = val
				case strings.HasSuffix(path, "BuyerTradeParty/SpecifiedLegalOrganization/ID"):
					buyerSIRET = val
				case strings.HasSuffix(path, "BuyerTradeParty/SpecifiedTaxRegistration/ID"):
					buyerVAT = val
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/LineTotalAmount"):
					lineTotalAmount = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/TaxBasisTotalAmount"):
					taxBasisTotal = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/TaxTotalAmount"):
					taxTotalAmount = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/GrandTotalAmount"):
					grandTotal = fVal
				case strings.HasSuffix(path, "SpecifiedTradeSettlementHeaderMonetarySummation/DuePayableAmount"):
					payableAmount = fVal

				// Lignes
				case inLine && strings.HasSuffix(path, "AssociatedDocumentLineDocument/LineID"):
					curLine.ID = val
				case inLine && strings.HasSuffix(path, "SpecifiedTradeSettlementLineMonetarySummation/LineTotalAmount"):
					curLine.Total = fVal

				// Taxes entête
				case inHeaderTax && strings.HasSuffix(path, "ApplicableTradeTax/BasisAmount"):
					curTax.BasisAmount = fVal
				case inHeaderTax && strings.HasSuffix(path, "ApplicableTradeTax/CalculatedAmount"):
					curTax.TaxAmount = fVal
				}
			}

			if el.Name.Local == "IncludedSupplyChainTradeLineItem" {
				lines = append(lines, curLine)
				inLine = false
			} else if el.Name.Local == "ApplicableTradeTax" && inHeaderTax {
				taxes = append(taxes, curTax)
				inHeaderTax = false
			}

			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			textBuf.Reset()
		}
	}

	var diags []SchematronDiagnostic

	if guidelineID == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-01", Severity: SchematronError, Message: "Missing Specification identifier", Path: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocumentContext/ram:GuidelineSpecifiedDocumentContextParameter/ram:ID"})
	}
	if invoiceID == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-02", Severity: SchematronError, Message: "Missing Invoice number", Path: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:ID"})
	}
	if issueDate == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-03", Severity: SchematronError, Message: "Missing Issue date", Path: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:IssueDateTime"})
	}
	if typeCode == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-04", Severity: SchematronError, Message: "Missing Type code", Path: "/rsm:CrossIndustryInvoice/rsm:ExchangedDocument/ram:TypeCode"})
	}
	if currency == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-05", Severity: SchematronError, Message: "Missing Currency code", Path: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:InvoiceCurrencyCode"})
	}
	if sellerName == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-06", Severity: SchematronError, Message: "Missing Seller name", Path: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:Name"})
	}
	if buyerName == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-07", Severity: SchematronError, Message: "Missing Buyer name", Path: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty/ram:Name"})
	}
	if sellerCountry == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-08", Severity: SchematronError, Message: "Missing Seller country", Path: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:PostalTradeAddress/ram:CountryID"})
	}
	if buyerCountry == "" {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-09", Severity: SchematronError, Message: "Missing Buyer country", Path: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty/ram:PostalTradeAddress/ram:CountryID"})
	}
	if len(lines) == 0 {
		diags = append(diags, SchematronDiagnostic{RuleID: "BR-16", Severity: SchematronError, Message: "At least one line is required", Path: "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:IncludedSupplyChainTradeLineItem"})
	}

	var sumLines float64
	for i, l := range lines {
		if strings.TrimSpace(l.ID) == "" {
			diags = append(diags, SchematronDiagnostic{
				RuleID:   "BR-21",
				Severity: SchematronError,
				Message:  fmt.Sprintf("Invoice line %d shall have a Line Identifier", i+1),
				Path:     fmt.Sprintf("/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:IncludedSupplyChainTradeLineItem[%d]/ram:AssociatedDocumentLineDocument/ram:LineID", i+1),
			})
		}
		sumLines += l.Total
	}

	if math.Abs(sumLines-lineTotalAmount) > 0.01 {
		diags = append(diags, SchematronDiagnostic{
			RuleID:   "BR-CO-10",
			Severity: SchematronError,
			Message:  fmt.Sprintf("Sum of Invoice line net amounts (%.2f) must equal LineTotalAmount (%.2f)", sumLines, lineTotalAmount),
			Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:LineTotalAmount",
		})
	}

	var sumTaxBase, sumTaxAmt float64
	for _, t := range taxes {
		sumTaxBase += t.BasisAmount
		sumTaxAmt += t.TaxAmount
	}

	if math.Abs(sumTaxBase-taxBasisTotal) > 0.01 {
		diags = append(diags, SchematronDiagnostic{
			RuleID:   "BR-CO-13",
			Severity: SchematronError,
			Message:  fmt.Sprintf("TaxBasisTotalAmount (%.2f) must equal sum of tax bases (%.2f)", taxBasisTotal, sumTaxBase),
			Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:TaxBasisTotalAmount",
		})
	}

	if math.Abs(sumTaxAmt-taxTotalAmount) > 0.01 {
		diags = append(diags, SchematronDiagnostic{
			RuleID:   "BR-CO-14",
			Severity: SchematronError,
			Message:  fmt.Sprintf("TaxTotalAmount (%.2f) must equal sum of tax amounts (%.2f)", taxTotalAmount, sumTaxAmt),
			Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:TaxTotalAmount",
		})
	}

	expectedGrand := taxBasisTotal + taxTotalAmount
	if math.Abs(grandTotal-expectedGrand) > 0.01 {
		diags = append(diags, SchematronDiagnostic{
			RuleID:   "BR-CO-15",
			Severity: SchematronError,
			Message:  fmt.Sprintf("GrandTotal (%.2f) must equal Net (%.2f) + Tax (%.2f)", grandTotal, taxBasisTotal, taxTotalAmount),
			Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:GrandTotalAmount",
		})
	}

	if math.Abs(payableAmount-grandTotal) > 0.01 && payableAmount > grandTotal {
		diags = append(diags, SchematronDiagnostic{
			RuleID:   "BR-CO-16",
			Severity: SchematronError,
			Message:  fmt.Sprintf("Payable (%.2f) cannot exceed GrandTotal (%.2f)", payableAmount, grandTotal),
			Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradeSettlementHeaderMonetarySummation/ram:DuePayableAmount",
		})
	}

	if issueDate != "" && dueDate != "" && dueDate < issueDate {
		diags = append(diags, SchematronDiagnostic{
			RuleID:   "BR-CO-25",
			Severity: SchematronError,
			Message:  fmt.Sprintf("Due date (%s) cannot be earlier than Issue date (%s)", dueDate, issueDate),
			Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeSettlement/ram:SpecifiedTradePaymentTerms/ram:DueDateDateTime",
		})
	}

	if e.strictCIUSFR || sellerCountry == "FR" {
		if sellerCountry == "FR" {
			cleanSellerSIRET := strings.ReplaceAll(sellerSIRET, " ", "")
			if cleanSellerSIRET == "" || (!sirenRegex.MatchString(cleanSellerSIRET) && !siretRegex.MatchString(cleanSellerSIRET)) {
				diags = append(diags, SchematronDiagnostic{
					RuleID:   "CIUS-FR-01",
					Severity: SchematronError,
					Message:  "L'identifiant du vendeur en France doit etre un SIREN (9) ou SIRET (14)",
					Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:SpecifiedLegalOrganization/ram:ID",
				})
			}
			cleanSellerVAT := strings.ReplaceAll(sellerVAT, " ", "")
			if cleanSellerVAT != "" && !vatFRRegex.MatchString(cleanSellerVAT) {
				diags = append(diags, SchematronDiagnostic{
					RuleID:   "CIUS-FR-02",
					Severity: SchematronError,
					Message:  "Format du numero de TVA intracommunautaire francais invalide",
					Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:SellerTradeParty/ram:SpecifiedTaxRegistration/ram:ID",
				})
			}
		}

		if buyerCountry == "FR" {
			if strings.TrimSpace(buyerRef) == "" {
				diags = append(diags, SchematronDiagnostic{
					RuleID:   "CIUS-FR-03",
					Severity: SchematronWarning,
					Message:  "La reference acheteur est recommandee pour Chorus Pro / PDP",
					Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerReference",
				})
			}
			cleanBuyerSIRET := strings.ReplaceAll(buyerSIRET, " ", "")
			if cleanBuyerSIRET != "" && !sirenRegex.MatchString(cleanBuyerSIRET) && !siretRegex.MatchString(cleanBuyerSIRET) {
				diags = append(diags, SchematronDiagnostic{
					RuleID:   "CIUS-FR-04",
					Severity: SchematronWarning,
					Message:  "L'identifiant acheteur francais doit correspondre a un SIREN (9) ou SIRET (14)",
					Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty/ram:SpecifiedLegalOrganization/ram:ID",
				})
			}
			cleanBuyerVAT := strings.ReplaceAll(buyerVAT, " ", "")
			if cleanBuyerVAT != "" && !vatFRRegex.MatchString(cleanBuyerVAT) {
				diags = append(diags, SchematronDiagnostic{
					RuleID:   "CIUS-FR-05",
					Severity: SchematronWarning,
					Message:  "Format du numero de TVA acheteur francais invalide",
					Path:     "/rsm:CrossIndustryInvoice/rsm:SupplyChainTradeTransaction/ram:ApplicableHeaderTradeAgreement/ram:BuyerTradeParty/ram:SpecifiedTaxRegistration/ram:ID",
				})
			}
		}
	}

	hasError := false
	for _, d := range diags {
		if d.Severity == SchematronError {
			hasError = true
			break
		}
	}

	return SchematronReport{
		Valid:       !hasError,
		Diagnostics: diags,
	}, nil
}
