package model

import (
"encoding/xml"
"strconv"
"strings"
"time"
)

type rawCIIAmount struct {
Value float64 `xml:",chardata"`
}

type rawCIIQuantity struct {
Value float64 `xml:",chardata"`
Unit  string  `xml:"unitCode,attr"`
}

type rawCIIParty struct {
Name string `xml:"Name"`
SpecifiedLegalOrganization struct {
ID struct {
Value    string `xml:",chardata"`
SchemeID string `xml:"schemeID,attr"`
} `xml:"ID"`
} `xml:"SpecifiedLegalOrganization"`
PostalTradeAddress struct {
CountryID string `xml:"CountryID"`
} `xml:"PostalTradeAddress"`
SpecifiedTaxRegistration struct {
ID struct {
Value    string `xml:",chardata"`
SchemeID string `xml:"schemeID,attr"`
} `xml:"ID"`
} `xml:"SpecifiedTaxRegistration"`
}

type rawCIITaxSubtotal struct {
CalculatedAmount rawCIIAmount `xml:"CalculatedAmount"`
BasisAmount      rawCIIAmount `xml:"BasisAmount"`
CategoryCode     string       `xml:"CategoryCode"`
RateApplicablePercent string  `xml:"RateApplicablePercent"`
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
ID       string `xml:"ID"`
TypeCode string `xml:"TypeCode"`
IssueDateTime struct {
DateTimeString struct {
Value  string `xml:",chardata"`
Format string `xml:"format,attr"`
} `xml:"DateTimeString"`
} `xml:"IssueDateTime"`
} `xml:"ExchangedDocument"`
SupplyChainTradeTransaction struct {
IncludedSupplyChainTradeLineItem []rawCIILineItem `xml:"IncludedSupplyChainTradeLineItem"`
ApplicableHeaderTradeAgreement  struct {
SellerTradeParty rawCIIParty `xml:"SellerTradeParty"`
BuyerTradeParty  rawCIIParty `xml:"BuyerTradeParty"`
} `xml:"ApplicableHeaderTradeAgreement"`
ApplicableHeaderTradeSettlement struct {
InvoiceCurrencyCode             string              `xml:"InvoiceCurrencyCode"`
ApplicableTradeTax              []rawCIITaxSubtotal `xml:"ApplicableTradeTax"`
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

// NormalizeCIIToCanonical convertit un flux XML UN/CEFACT CII (Factur-X) en CanonicalInvoice
func NormalizeCIIToCanonical(xmlPayload []byte) (*CanonicalInvoice, error) {
var doc rawCIIDocument
if err := xml.Unmarshal(xmlPayload, &doc); err != nil {
return nil, err
}

dateRaw := strings.TrimSpace(doc.ExchangedDocument.IssueDateTime.DateTimeString.Value)
date, _ := time.Parse("20060102", dateRaw)
if date.IsZero() {
date, _ = time.Parse("2006-01-02", dateRaw)
}

agreement := doc.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement
settlement := doc.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement
totals := settlement.SpecifiedTradeSettlementHeaderMonetarySummation

sellerCountry := strings.ToUpper(strings.TrimSpace(agreement.SellerTradeParty.PostalTradeAddress.CountryID))
buyerCountry := strings.ToUpper(strings.TrimSpace(agreement.BuyerTradeParty.PostalTradeAddress.CountryID))

targetJurisdiction := "FR"
if sellerCountry == "MA" || buyerCountry == "MA" || settlement.InvoiceCurrencyCode == "MAD" {
targetJurisdiction = "MA"
} else if sellerCountry != "" {
targetJurisdiction = sellerCountry
}

sellerTaxID := agreement.SellerTradeParty.SpecifiedTaxRegistration.ID.Value
sellerOrgID := agreement.SellerTradeParty.SpecifiedLegalOrganization.ID.Value
buyerTaxID := agreement.BuyerTradeParty.SpecifiedTaxRegistration.ID.Value
buyerOrgID := agreement.BuyerTradeParty.SpecifiedLegalOrganization.ID.Value

invoice := &CanonicalInvoice{
ID:                 doc.ExchangedDocument.ID,
InvoiceNumber:      doc.ExchangedDocument.ID,
IssueDate:          date,
Currency:           strings.TrimSpace(settlement.InvoiceCurrencyCode),
SourceSyntax:       "CII-D16B",
TargetJurisdiction: targetJurisdiction,
Seller: Party{
Name:        agreement.SellerTradeParty.Name,
TaxID:       sellerTaxID,
Country:     sellerCountry,
Identifiers: make(map[string]string),
},
Buyer: Party{
Name:        agreement.BuyerTradeParty.Name,
TaxID:       buyerTaxID,
Country:     buyerCountry,
Identifiers: make(map[string]string),
},
Totals: MonetaryTotals{
LineExtensionAmount: totals.LineTotalAmount.Value,
TaxExclusiveAmount:  totals.TaxBasisTotalAmount.Value,
TaxInclusiveAmount:  totals.GrandTotalAmount.Value,
PayableAmount:       totals.DuePayableAmount.Value,
},
Lines: make([]InvoiceLine, 0, len(doc.SupplyChainTradeTransaction.IncludedSupplyChainTradeLineItem)),
}

if sellerOrgID != "" {
invoice.Seller.Identifiers["SIRET"] = sellerOrgID
}
if buyerOrgID != "" {
invoice.Buyer.Identifiers["SIRET"] = buyerOrgID
}

if targetJurisdiction == "MA" {
invoice.Seller.NationalID = sellerOrgID
invoice.Buyer.NationalID = buyerOrgID
}

// Normalisation des lignes
for _, item := range doc.SupplyChainTradeTransaction.IncludedSupplyChainTradeLineItem {
desc := item.SpecifiedTradeProduct.Description
if desc == "" {
desc = item.SpecifiedTradeProduct.Name
}
invoice.Lines = append(invoice.Lines, InvoiceLine{
ID:          item.AssociatedDocumentLineDocument.LineID,
Description: desc,
Quantity:    item.SpecifiedLineTradeDelivery.BilledQuantity.Value,
UnitPrice:   item.SpecifiedLineTradeAgreement.NetPriceProductTradePrice.ChargeAmount.Value,
LineTotal:   item.SpecifiedLineTradeSettlement.SpecifiedTradeSettlementLineMonetarySummation.LineTotalAmount.Value,
})
}

// Normalisation des taxes
for _, tax := range settlement.ApplicableTradeTax {
pct, _ := strconv.ParseFloat(tax.RateApplicablePercent, 64)
invoice.TaxSubtotals = append(invoice.TaxSubtotals, TaxSubtotal{
TaxableAmount: tax.BasisAmount.Value,
TaxAmount:     tax.CalculatedAmount.Value,
Percent:       pct,
CategoryCode:  tax.CategoryCode,
})
}

return invoice, nil
}
