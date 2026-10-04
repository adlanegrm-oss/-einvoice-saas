package syntax

import (
"encoding/xml"
"errors"
"fmt"
"regexp"
"strings"
"time"

"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

var (
sirenRegex = regexp.MustCompile(`^[0-9]{9}$`)
siretRegex = regexp.MustCompile(`^[0-9]{14}$`)
vatFRRegex = regexp.MustCompile(`^FR[0-9A-Z]{2}[0-9]{9}$`)
)

type ValidationDiagnostic struct {
RuleID   string `json:"rule_id"`
Severity string `json:"severity"` // ERROR, WARNING
Message  string `json:"message"`
Path     string `json:"path"`
}

type Report struct {
Valid       bool                   `json:"valid"`
Diagnostics []ValidationDiagnostic `json:"diagnostics,omitempty"`
}

// ValidateEN16931Strict vérifie les règles sémantiques EN 16931 + CIUS-FR
func ValidateEN16931Strict(inv *canonical.CanonicalInvoice) Report {
var diags []ValidationDiagnostic

// BR-01: Numéro de facture obligatoire
if strings.TrimSpace(inv.InvoiceNumber) == "" {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-01", Severity: "ERROR", Path: "/Invoice/ID",
Message: "Le numero de facture est obligatoire",
})
}

// BR-02: Date d'émission obligatoire
if inv.IssueDate.IsZero() {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-02", Severity: "ERROR", Path: "/Invoice/IssueDate",
Message: "La date d'emission est obligatoire",
})
}

// BR-CO-04: DueDate >= IssueDate
if !inv.DueDate.IsZero() && inv.DueDate.Before(inv.IssueDate) {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-04", Severity: "ERROR", Path: "/Invoice/DueDate",
Message: "La date d'echeance ne peut pas etre anterieure a la date d'emission",
})
}

// BR-CO-09 & CIUS-FR-01: Identifiants Vendeur (SIREN/SIRET et TVA)
if inv.Seller.CountryCode == "FR" {
cleanedVat := strings.ReplaceAll(inv.Seller.TaxID, " ", "")
if cleanedVat != "" && !vatFRRegex.MatchString(cleanedVat) {
diags = append(diags, ValidationDiagnostic{
RuleID: "CIUS-FR-VAT", Severity: "ERROR", Path: "/Invoice/Seller/TaxID",
Message: "Format de TVA FR invalide (attendu FR + 2 car. + 9 chiffres)",
})
}
if inv.Seller.ID != "" && !siretRegex.MatchString(inv.Seller.ID) && !sirenRegex.MatchString(inv.Seller.ID) {
diags = append(diags, ValidationDiagnostic{
RuleID: "CIUS-FR-SIRET", Severity: "ERROR", Path: "/Invoice/Seller/ID",
Message: "L'identifiant national francais doit etre un SIREN (9) ou SIRET (14)",
})
}
}

// BR-CO-15: GrossTTC = NetHT + TaxAmount (en centimes)
expectedTTC := inv.MonetaryTotals.NetHT + inv.MonetaryTotals.TaxAmount
if inv.MonetaryTotals.GrossTTC != expectedTTC {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-15", Severity: "ERROR", Path: "/Invoice/MonetaryTotals/GrossTTC",
Message: fmt.Sprintf("Incoherence montant total: NetHT (%d) + Tax (%d) != TTC (%d)",
inv.MonetaryTotals.NetHT, inv.MonetaryTotals.TaxAmount, inv.MonetaryTotals.GrossTTC),
})
}

// BR-CO-13: La somme des bases HT doit matcher NetHT
var sumBaseHT int64
var sumTax int64
for _, tb := range inv.TaxBreakdowns {
sumBaseHT += tb.BaseHT
sumTax += tb.TaxAmount
}
if sumBaseHT != inv.MonetaryTotals.NetHT {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-13", Severity: "ERROR", Path: "/Invoice/TaxBreakdowns",
Message: fmt.Sprintf("Somme des bases HT de TVA (%d) != NetHT facture (%d)", sumBaseHT, inv.MonetaryTotals.NetHT),
})
}
if sumTax != inv.MonetaryTotals.TaxAmount {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-14", Severity: "ERROR", Path: "/Invoice/TaxBreakdowns",
Message: fmt.Sprintf("Somme des montants TVA (%d) != Total TVA facture (%d)", sumTax, inv.MonetaryTotals.TaxAmount),
})
}

return Report{
Valid:       len(diags) == 0,
Diagnostics: diags,
}
}

// GenerateCIIXML sérialise le modèle canonique en flux XML UN/CEFACT CII (D16B)
func GenerateCIIXML(inv *canonical.CanonicalInvoice) ([]byte, error) {
type Amount struct {
Currency string `xml:"currencyID,attr"`
Value    string `xml:",chardata"`
}

var root struct {
XMLName xml.Name `xml:"rsm:CrossIndustryInvoice"`
XmlnsRSM string   `xml:"xmlns:rsm,attr"`
XmlnsRAM string   `xml:"xmlns:ram,attr"`
XmlnsQDT string   `xml:"xmlns:qdt,attr"`
XmlnsUDT string   `xml:"xmlns:udt,attr"`
Context  struct {
GuidelineID string `xml:"ram:GuidelineSpecifiedDocumentContextParameter>ram:ID"`
} `xml:"rsm:ExchangedDocumentContext"`
Header struct {
ID       string `xml:"ram:ID"`
TypeCode string `xml:"ram:TypeCode"`
IssueDate struct {
Date string `xml:"udt:DateTimeString"`
} `xml:"ram:IssueDateTime"`
} `xml:"rsm:ExchangedDocument"`
SupplyChainTradeTransaction struct {
ApplicableHeaderTradeSettlement struct {
InvoiceCurrencyCode string `xml:"ram:InvoiceCurrencyCode"`
SpecifiedTradeSettlementHeaderMonetarySummation struct {
LineTotalAmount Amount `xml:"ram:LineTotalAmount"`
TaxTotalAmount  Amount `xml:"ram:TaxTotalAmount"`
GrandTotalAmount Amount `xml:"ram:GrandTotalAmount"`
DuePayableAmount Amount `xml:"ram:DuePayableAmount"`
} `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
} `xml:"ram:ApplicableHeaderTradeSettlement"`
} `xml:"rsm:SupplyChainTradeTransaction"`
}

root.XmlnsRSM = "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
root.XmlnsRAM = "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
root.XmlnsQDT = "urn:un:unece:uncefact:data:standard:QualifiedDataType:100"
root.XmlnsUDT = "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100"
root.Context.GuidelineID = "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:extended"
root.Header.ID = inv.InvoiceNumber
root.Header.TypeCode = "380"
root.Header.IssueDate.Date = inv.IssueDate.Format("20060102")

curr := inv.Currency
if curr == "" { curr = "EUR" }
settle := &root.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement
settle.InvoiceCurrencyCode = curr
settle.SpecifiedTradeSettlementHeaderMonetarySummation.LineTotalAmount = Amount{Currency: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.NetHT)/100.0)}
settle.SpecifiedTradeSettlementHeaderMonetarySummation.TaxTotalAmount = Amount{Currency: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.TaxAmount)/100.0)}
settle.SpecifiedTradeSettlementHeaderMonetarySummation.GrandTotalAmount = Amount{Currency: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.GrossTTC)/100.0)}
settle.SpecifiedTradeSettlementHeaderMonetarySummation.DuePayableAmount = Amount{Currency: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.PayableDue)/100.0)}

raw, err := xml.MarshalIndent(root, "", "  ")
if err != nil {
return nil, fmt.Errorf("serialisation CII : %w", err)
}
return append([]byte(xml.Header), raw...), nil
}