package model

import (
"bytes"
"encoding/xml"
"fmt"
"strconv"
"strings"
"time"
)

type rawCIIAmount struct {
Value    float64 `xml:",chardata"`
Currency string  `xml:"currencyID,attr"`
}

type rawCIIQuantity struct {
Value float64 `xml:",chardata"`
Unit  string  `xml:"unitCode,attr"`
}

type rawCIITaxSubtotal struct {
CalculatedAmount      rawCIIAmount `xml:"CalculatedAmount"`
TypeCode              string       `xml:"TypeCode"`
BasisAmount           rawCIIAmount `xml:"BasisAmount"`
CategoryCode          string       `xml:"CategoryCode"`
RateApplicablePercent string       `xml:"RateApplicablePercent"`
}

type rawCIILineItem struct {
AssociatedDocumentLineDocument struct {
LineID string `xml:"LineID"`
} `xml:"AssociatedDocumentLineDocument"`

SpecifiedTradeProduct struct {
Name        string `xml:"Name"`
Description string `xml:"Description"`
} `xml:"SpecifiedTradeProduct"`

SpecifiedLineTradeAgreement struct {
NetPriceProductTradePrice struct {
ChargeAmount rawCIIAmount `xml:"ChargeAmount"`
} `xml:"NetPriceProductTradePrice"`
} `xml:"SpecifiedLineTradeAgreement"`

SpecifiedLineTradeDelivery struct {
BilledQuantity rawCIIQuantity `xml:"BilledQuantity"`
} `xml:"SpecifiedLineTradeDelivery"`

SpecifiedLineTradeSettlement struct {
ApplicableTradeTax []rawCIITaxSubtotal `xml:"ApplicableTradeTax"`

SpecifiedTradeSettlementLineMonetarySummation struct {
LineTotalAmount rawCIIAmount `xml:"LineTotalAmount"`
} `xml:"SpecifiedTradeSettlementLineMonetarySummation"`
} `xml:"SpecifiedLineTradeSettlement"`
}

type rawCIIDocument struct {
XMLName xml.Name `xml:"CrossIndustryInvoice"`

ExchangedDocumentContext struct {
GuidelineSpecifiedDocumentContextParameter struct {
ID string `xml:"ID"`
} `xml:"GuidelineSpecifiedDocumentContextParameter"`
} `xml:"ExchangedDocumentContext"`

ExchangedDocument struct {
ID            string `xml:"ID"`
TypeCode      string `xml:"TypeCode"`
IssueDateTime struct {
DateTimeString struct {
Value  string `xml:",chardata"`
Format string `xml:"format,attr"`
} `xml:"DateTimeString"`
} `xml:"IssueDateTime"`
} `xml:"ExchangedDocument"`

SupplyChainTradeTransaction struct {
IncludedSupplyChainTradeLineItem []rawCIILineItem `xml:"IncludedSupplyChainTradeLineItem"`

ApplicableHeaderTradeAgreement struct {
SellerTradeParty struct {
Name                     string `xml:"Name"`
SpecifiedLegalOrganization struct {
ID string `xml:"ID"`
} `xml:"SpecifiedLegalOrganization"`
SpecifiedTaxRegistration struct {
ID string `xml:"ID"`
} `xml:"SpecifiedTaxRegistration"`
PostalTradeAddress struct {
PostcodeCode string `xml:"PostcodeCode"`
LineOne      string `xml:"LineOne"`
CityName     string `xml:"CityName"`
CountryID    string `xml:"CountryID"`
} `xml:"PostalTradeAddress"`
} `xml:"SellerTradeParty"`

BuyerTradeParty struct {
Name                     string `xml:"Name"`
SpecifiedLegalOrganization struct {
ID string `xml:"ID"`
} `xml:"SpecifiedLegalOrganization"`
SpecifiedTaxRegistration struct {
ID string `xml:"ID"`
} `xml:"SpecifiedTaxRegistration"`
PostalTradeAddress struct {
PostcodeCode string `xml:"PostcodeCode"`
LineOne      string `xml:"LineOne"`
CityName     string `xml:"CityName"`
CountryID    string `xml:"CountryID"`
} `xml:"PostalTradeAddress"`
} `xml:"BuyerTradeParty"`
} `xml:"ApplicableHeaderTradeAgreement"`

ApplicableHeaderTradeSettlement struct {
InvoiceCurrencyCode string              `xml:"InvoiceCurrencyCode"`
PaymentReference    string              `xml:"PaymentReference"`
ApplicableTradeTax  []rawCIITaxSubtotal `xml:"ApplicableTradeTax"`

SpecifiedTradePaymentTerms struct {
DueDateDateTime struct {
DateTimeString struct {
Value  string `xml:",chardata"`
Format string `xml:"format,attr"`
} `xml:"DateTimeString"`
} `xml:"DueDateDateTime"`
} `xml:"SpecifiedTradePaymentTerms"`

SpecifiedTradeSettlementHeaderMonetarySummation struct {
LineTotalAmount     rawCIIAmount `xml:"LineTotalAmount"`
TaxBasisTotalAmount rawCIIAmount `xml:"TaxBasisTotalAmount"`
TaxTotalAmount      rawCIIAmount `xml:"TaxTotalAmount"`
GrandTotalAmount    rawCIIAmount `xml:"GrandTotalAmount"`
DuePayableAmount    rawCIIAmount `xml:"DuePayableAmount"`
} `xml:"SpecifiedTradeSettlementHeaderMonetarySummation"`
} `xml:"ApplicableHeaderTradeSettlement"`
} `xml:"SupplyChainTradeTransaction"`
}

func parseCIIDate(value string) (time.Time, error) {
value = strings.TrimSpace(value)
if value == "" {
return time.Time{}, nil
}

formats := []string{
"20060102",
"2006-01-02",
"20060102150405",
"2006-01-02T15:04:05",
"2006-01-02T15:04:05Z07:00",
}

for _, format := range formats {
if t, err := time.Parse(format, value); err == nil {
return t, nil
}
}

return time.Time{}, fmt.Errorf("format de date CII invalide: %q", value)
}

func NormalizeCIIToCanonical(xmlData []byte) (*CanonicalInvoice, error) {
if len(bytes.TrimSpace(xmlData)) == 0 {
return nil, fmt.Errorf("cii: payload XML vide")
}

var doc rawCIIDocument
if err := xml.Unmarshal(xmlData, &doc); err != nil {
return nil, fmt.Errorf("cii: echec unmarshal XML: %w", err)
}

issueDateRaw := strings.TrimSpace(doc.ExchangedDocument.IssueDateTime.DateTimeString.Value)
if issueDateRaw == "" {
return nil, fmt.Errorf("cii: la date d'emission (IssueDateTime) est obligatoire")
}
issueDate, err := parseCIIDate(issueDateRaw)
if err != nil {
return nil, fmt.Errorf("cii: date d'emission invalide: %w", err)
}

settlement := doc.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement
agreement := doc.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement

var dueDate *time.Time
dueDateRaw := strings.TrimSpace(settlement.SpecifiedTradePaymentTerms.DueDateDateTime.DateTimeString.Value)
if dueDateRaw != "" {
parsedDue, err := parseCIIDate(dueDateRaw)
if err != nil {
return nil, fmt.Errorf("cii: date d'echeance invalide: %w", err)
}
dueDate = &parsedDue
}

canonical := &CanonicalInvoice{
		ID:                        strings.TrimSpace(doc.ExchangedDocument.ID),
		InvoiceNumber:             strings.TrimSpace(doc.ExchangedDocument.ID),
		InvoiceTypeCode:           strings.TrimSpace(doc.ExchangedDocument.TypeCode),
		IssueDate:                 issueDate, // <--- LIGNE EN DOUBLE À SUPPRIMER
		DueDate:                   dueDate,
		Currency:                  strings.TrimSpace(settlement.InvoiceCurrencyCode),
		SourceSyntax:              "CII-D16B",
		TargetJurisdiction:        "FR",
		PaymentReference:          strings.TrimSpace(settlement.PaymentReference),

Seller: Party{
Name:        strings.TrimSpace(agreement.SellerTradeParty.Name),
TaxID:       strings.TrimSpace(agreement.SellerTradeParty.SpecifiedTaxRegistration.ID),
NationalID:  strings.TrimSpace(agreement.SellerTradeParty.SpecifiedLegalOrganization.ID),
AddressLine: strings.TrimSpace(agreement.SellerTradeParty.PostalTradeAddress.LineOne),
City:        strings.TrimSpace(agreement.SellerTradeParty.PostalTradeAddress.CityName),
PostalZone:  strings.TrimSpace(agreement.SellerTradeParty.PostalTradeAddress.PostcodeCode),
Country:     strings.TrimSpace(agreement.SellerTradeParty.PostalTradeAddress.CountryID),
},
Buyer: Party{
Name:        strings.TrimSpace(agreement.BuyerTradeParty.Name),
TaxID:       strings.TrimSpace(agreement.BuyerTradeParty.SpecifiedTaxRegistration.ID),
NationalID:  strings.TrimSpace(agreement.BuyerTradeParty.SpecifiedLegalOrganization.ID),
AddressLine: strings.TrimSpace(agreement.BuyerTradeParty.PostalTradeAddress.LineOne),
City:        strings.TrimSpace(agreement.BuyerTradeParty.PostalTradeAddress.CityName),
PostalZone:  strings.TrimSpace(agreement.BuyerTradeParty.PostalTradeAddress.PostcodeCode),
Country:     strings.TrimSpace(agreement.BuyerTradeParty.PostalTradeAddress.CountryID),
},
Totals: MonetaryTotals{
LineExtensionAmount: settlement.SpecifiedTradeSettlementHeaderMonetarySummation.LineTotalAmount.Value,
TaxExclusiveAmount:  settlement.SpecifiedTradeSettlementHeaderMonetarySummation.TaxBasisTotalAmount.Value,
TaxInclusiveAmount:  settlement.SpecifiedTradeSettlementHeaderMonetarySummation.GrandTotalAmount.Value,
PayableAmount:       settlement.SpecifiedTradeSettlementHeaderMonetarySummation.DuePayableAmount.Value,
},
}

for _, item := range doc.SupplyChainTradeTransaction.IncludedSupplyChainTradeLineItem {
desc := strings.TrimSpace(item.SpecifiedTradeProduct.Description)
if desc == "" {
desc = strings.TrimSpace(item.SpecifiedTradeProduct.Name)
}

var vatPercent float64
var vatCategory string

if len(item.SpecifiedLineTradeSettlement.ApplicableTradeTax) > 0 {
tax := item.SpecifiedLineTradeSettlement.ApplicableTradeTax[0]
vatCategory = strings.TrimSpace(tax.CategoryCode)

pctRaw := strings.TrimSpace(tax.RateApplicablePercent)
if pctRaw != "" {
var parseErr error
vatPercent, parseErr = strconv.ParseFloat(pctRaw, 64)
if parseErr != nil {
return nil, fmt.Errorf("cii: taux TVA invalide %q sur la ligne %q: %w", pctRaw, item.AssociatedDocumentLineDocument.LineID, parseErr)
}
}
}

canonical.Lines = append(canonical.Lines, InvoiceLine{
ID:          strings.TrimSpace(item.AssociatedDocumentLineDocument.LineID),
Description: desc,
Quantity:    item.SpecifiedLineTradeDelivery.BilledQuantity.Value,
UnitPrice:   item.SpecifiedLineTradeAgreement.NetPriceProductTradePrice.ChargeAmount.Value,
LineTotal:   item.SpecifiedLineTradeSettlement.SpecifiedTradeSettlementLineMonetarySummation.LineTotalAmount.Value,
VatPercent:  vatPercent,
VatCategory: vatCategory,
})
}

for _, tax := range settlement.ApplicableTradeTax {
pctRaw := strings.TrimSpace(tax.RateApplicablePercent)
var pct float64
if pctRaw != "" {
var parseErr error
pct, parseErr = strconv.ParseFloat(pctRaw, 64)
if parseErr != nil {
return nil, fmt.Errorf("cii: taux TVA en-tete invalide %q: %w", pctRaw, parseErr)
}
}

canonical.TaxSubtotals = append(canonical.TaxSubtotals, TaxSubtotal{
TaxableAmount: tax.BasisAmount.Value,
TaxAmount:     tax.CalculatedAmount.Value,
Percent:       pct,
CategoryCode:  strings.TrimSpace(tax.CategoryCode),
})
}

return canonical, nil
}
