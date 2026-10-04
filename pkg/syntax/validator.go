package syntax

import (
"encoding/xml"
"fmt"
"regexp"
"strings"

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

// ValidateCoreBusinessRules valide les règles de calcul financières et CIUS-FR de base.
// NOTE: Ne remplace pas la validation officielle par Schematron XSLT EN 16931.
func ValidateCoreBusinessRules(inv *canonical.CanonicalInvoice) Report {
var diags []ValidationDiagnostic

if strings.TrimSpace(inv.InvoiceNumber) == "" {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-01", Severity: "ERROR", Path: "/Invoice/ID",
Message: "Le numero de facture est obligatoire",
})
}

if inv.IssueDate.IsZero() {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-02", Severity: "ERROR", Path: "/Invoice/IssueDate",
Message: "La date d'emission est obligatoire",
})
}

if !inv.DueDate.IsZero() && inv.DueDate.Before(inv.IssueDate) {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-04", Severity: "ERROR", Path: "/Invoice/DueDate",
Message: "La date d'echeance ne peut pas etre anterieure a la date d'emission",
})
}

if inv.Seller.CountryCode == "FR" {
cleanedVat := strings.ReplaceAll(inv.Seller.TaxID, " ", "")
if cleanedVat != "" && !vatFRRegex.MatchString(cleanedVat) {
diags = append(diags, ValidationDiagnostic{
RuleID: "CIUS-FR-VAT", Severity: "ERROR", Path: "/Invoice/Seller/TaxID",
Message: "Format TVA FR invalide",
})
}
if inv.Seller.ID != "" && !siretRegex.MatchString(inv.Seller.ID) && !sirenRegex.MatchString(inv.Seller.ID) {
diags = append(diags, ValidationDiagnostic{
RuleID: "CIUS-FR-SIRET", Severity: "ERROR", Path: "/Invoice/Seller/ID",
Message: "Identifiant FR vendeur attendu SIREN/SIRET",
})
}
}

expectedTTC := inv.MonetaryTotals.NetHT + inv.MonetaryTotals.TaxAmount
if inv.MonetaryTotals.GrossTTC != expectedTTC {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-15", Severity: "ERROR", Path: "/Invoice/MonetaryTotals/GrossTTC",
Message: fmt.Sprintf("Incoherence total: NetHT (%d) + Tax (%d) != TTC (%d)",
inv.MonetaryTotals.NetHT, inv.MonetaryTotals.TaxAmount, inv.MonetaryTotals.GrossTTC),
})
}

var sumBaseHT, sumTax int64
for _, tb := range inv.TaxBreakdowns {
sumBaseHT += tb.BaseHT
sumTax += tb.TaxAmount
}
if sumBaseHT != inv.MonetaryTotals.NetHT {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-13", Severity: "ERROR", Path: "/Invoice/TaxBreakdowns",
Message: fmt.Sprintf("Somme des bases HT (%d) != NetHT (%d)", sumBaseHT, inv.MonetaryTotals.NetHT),
})
}
if sumTax != inv.MonetaryTotals.TaxAmount {
diags = append(diags, ValidationDiagnostic{
RuleID: "BR-CO-14", Severity: "ERROR", Path: "/Invoice/TaxBreakdowns",
Message: fmt.Sprintf("Somme des montants TVA (%d) != Total TVA (%d)", sumTax, inv.MonetaryTotals.TaxAmount),
})
}

return Report{
Valid:       len(diags) == 0,
Diagnostics: diags,
}
}

// GenerateCIIXML génère un document CII D16B avec vendeur, acheteur, lignes et ventilation
func GenerateCIIXML(inv *canonical.CanonicalInvoice) ([]byte, error) {
type Amount struct {
Currency string `xml:"currencyID,attr,omitempty"`
Value    string `xml:",chardata"`
}

curr := inv.Currency
if curr == "" {
curr = "EUR"
}

var root struct {
XMLName  xml.Name `xml:"rsm:CrossIndustryInvoice"`
XmlnsRSM string   `xml:"xmlns:rsm,attr"`
XmlnsRAM string   `xml:"xmlns:ram,attr"`
XmlnsUDT string   `xml:"xmlns:udt,attr"`
Context  struct {
GuidelineID string `xml:"ram:GuidelineSpecifiedDocumentContextParameter>ram:ID"`
} `xml:"rsm:ExchangedDocumentContext"`
Header struct {
ID        string `xml:"ram:ID"`
TypeCode  string `xml:"ram:TypeCode"`
IssueDate struct {
Date string `xml:"udt:DateTimeString"`
} `xml:"ram:IssueDateTime"`
} `xml:"rsm:ExchangedDocument"`
SupplyChainTradeTransaction struct {
IncludedSupplyChainTradeLineItem []struct {
LineID string `xml:"ram:AssociatedDocumentLineDocument>ram:LineID"`
Item   struct {
Name string `xml:"ram:Name"`
} `xml:"ram:SpecifiedTradeProduct"`
Price struct {
ChargeAmount Amount `xml:"ram:ChargeAmount"`
} `xml:"ram:SpecifiedLineTradeAgreement>ram:NetPriceProductTradePrice"`
Settlement struct {
TradeTax struct {
TypeCode     string `xml:"ram:TypeCode"`
CategoryCode string `xml:"ram:CategoryCode"`
RateApplicablePercent string `xml:"ram:RateApplicablePercent"`
} `xml:"ram:ApplicableTradeTax"`
LineTotal Amount `xml:"ram:SpecifiedTradeSettlementLineMonetarySummation>ram:LineTotalAmount"`
} `xml:"ram:SpecifiedLineTradeSettlement"`
} `xml:"ram:IncludedSupplyChainTradeLineItem"`
ApplicableHeaderTradeAgreement struct {
SellerTradeParty struct {
Name           string `xml:"ram:Name"`
ID             string `xml:"ram:SpecifiedLegalOrganization>ram:ID"`
TaxRegistration struct {
ID string `xml:"ram:ID"`
} `xml:"ram:SpecifiedTaxRegistration>ram:ID"`
} `xml:"ram:SellerTradeParty"`
BuyerTradeParty struct {
Name           string `xml:"ram:Name"`
ID             string `xml:"ram:SpecifiedLegalOrganization>ram:ID"`
} `xml:"ram:BuyerTradeParty"`
} `xml:"ram:ApplicableHeaderTradeAgreement"`
ApplicableHeaderTradeSettlement struct {
InvoiceCurrencyCode string `xml:"ram:InvoiceCurrencyCode"`
ApplicableTradeTax  []struct {
CalculatedAmount Amount `xml:"ram:CalculatedAmount"`
TypeCode         string `xml:"ram:TypeCode"`
BasisAmount      Amount `xml:"ram:BasisAmount"`
CategoryCode     string `xml:"ram:CategoryCode"`
RateApplicablePercent string `xml:"ram:RateApplicablePercent"`
} `xml:"ram:ApplicableTradeTax"`
SpecifiedTradeSettlementHeaderMonetarySummation struct {
LineTotalAmount  Amount `xml:"ram:LineTotalAmount"`
TaxTotalAmount   Amount `xml:"ram:TaxTotalAmount"`
GrandTotalAmount Amount `xml:"ram:GrandTotalAmount"`
DuePayableAmount Amount `xml:"ram:DuePayableAmount"`
} `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
} `xml:"ram:ApplicableHeaderTradeSettlement"`
} `xml:"rsm:SupplyChainTradeTransaction"`
}

root.XmlnsRSM = "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
root.XmlnsRAM = "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
root.XmlnsUDT = "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100"
// Profil EN 16931 de base (sans prétention au profil extended avant certification)
root.Context.GuidelineID = "urn:cen.eu:en16931:2017"
root.Header.ID = inv.InvoiceNumber
root.Header.TypeCode = "380"
root.Header.IssueDate.Date = inv.IssueDate.Format("20060102")

tradeAggr := &root.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement
tradeAggr.SellerTradeParty.Name = inv.Seller.Name
tradeAggr.SellerTradeParty.ID = inv.Seller.ID
tradeAggr.SellerTradeParty.TaxRegistration.ID = inv.Seller.TaxID
tradeAggr.BuyerTradeParty.Name = inv.Buyer.Name
tradeAggr.BuyerTradeParty.ID = inv.Buyer.ID

settle := &root.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement
settle.InvoiceCurrencyCode = curr

for _, tb := range inv.TaxBreakdowns {
settle.ApplicableTradeTax = append(settle.ApplicableTradeTax, struct {
CalculatedAmount Amount `xml:"ram:CalculatedAmount"`
TypeCode         string `xml:"ram:TypeCode"`
BasisAmount      Amount `xml:"ram:BasisAmount"`
CategoryCode     string `xml:"ram:CategoryCode"`
RateApplicablePercent string `xml:"ram:RateApplicablePercent"`
}{
CalculatedAmount:      Amount{Currency: curr, Value: fmt.Sprintf("%.2f", float64(tb.TaxAmount)/100.0)},
TypeCode:              "VAT",
BasisAmount:           Amount{Currency: curr, Value: fmt.Sprintf("%.2f", float64(tb.BaseHT)/100.0)},
CategoryCode:          tb.VATCategory,
RateApplicablePercent: fmt.Sprintf("%.2f", float64(tb.VATRate)/100.0),
})
}

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