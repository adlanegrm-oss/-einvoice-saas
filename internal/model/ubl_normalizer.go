package model

import (
"encoding/xml"
"strconv"
"strings"
"time"
)

type rawParty struct {
PartyIdentification struct {
ID struct {
Value    string `xml:",chardata"`
SchemeID string `xml:"schemeID,attr"`
} `xml:"ID"`
} `xml:"PartyIdentification"`
PartyTaxScheme struct {
CompanyID string `xml:"CompanyID"`
} `xml:"PartyTaxScheme"`
PartyLegalEntity struct {
RegistrationName string `xml:"RegistrationName"`
} `xml:"PartyLegalEntity"`
}

type rawTaxSubtotal struct {
TaxableAmount struct {
Value float64 `xml:",chardata"`
} `xml:"TaxableAmount"`
TaxAmount struct {
Value float64 `xml:",chardata"`
} `xml:"TaxAmount"`
TaxCategory struct {
Percent string `xml:"Percent"`
ID      string `xml:"ID"`
} `xml:"TaxCategory"`
}

type rawMonetaryTotal struct {
LineExtensionAmount float64 `xml:"LineExtensionAmount"`
TaxExclusiveAmount  float64 `xml:"TaxExclusiveAmount"`
TaxInclusiveAmount  float64 `xml:"TaxInclusiveAmount"`
PayableAmount       float64 `xml:"PayableAmount"`
}

type rawInvoiceLine struct {
ID                  string  `xml:"ID"`
InvoicedQuantity    float64 `xml:"InvoicedQuantity"`
LineExtensionAmount float64 `xml:"LineExtensionAmount"`
Item                struct {
Name        string `xml:"Name"`
Description string `xml:"Description"`
} `xml:"Item"`
Price struct {
PriceAmount float64 `xml:"PriceAmount"`
} `xml:"Price"`
}

type rawUBLDocument struct {
XMLName                 xml.Name         `xml:"Invoice"`
CustomizationID         string           `xml:"CustomizationID"`
ID                      string           `xml:"ID"`
InvoiceTypeCode         string           `xml:"InvoiceTypeCode"`
IssueDate               string           `xml:"IssueDate"`
DocumentCurrencyCode    string           `xml:"DocumentCurrencyCode"`
AccountingSupplierParty struct {
Party rawParty `xml:"Party"`
} `xml:"AccountingSupplierParty"`
AccountingCustomerParty struct {
Party rawParty `xml:"Party"`
} `xml:"AccountingCustomerParty"`
TaxTotal struct {
TaxSubtotal []rawTaxSubtotal `xml:"TaxSubtotal"`
} `xml:"TaxTotal"`
LegalMonetaryTotal rawMonetaryTotal `xml:"LegalMonetaryTotal"`
InvoiceLines       []rawInvoiceLine `xml:"InvoiceLine"`
}

// NormalizeUBLToCanonical convertit un flux UBL en instance canonique agnostique
func NormalizeUBLToCanonical(xmlPayload []byte) (*CanonicalInvoice, error) {
var doc rawUBLDocument
if err := xml.Unmarshal(xmlPayload, &doc); err != nil {
return nil, err
}

date, _ := time.Parse("2006-01-02", strings.TrimSpace(doc.IssueDate))

// Détection du pays cible selon l'en-tête ou la monnaie
jurisdiction := "FR"
if strings.Contains(doc.CustomizationID, "urn:dgi.gov.ma") || doc.DocumentCurrencyCode == "MAD" {
jurisdiction = "MA"
}

sellerScheme := doc.AccountingSupplierParty.Party.PartyIdentification.ID.SchemeID
sellerIDVal := strings.TrimSpace(doc.AccountingSupplierParty.Party.PartyIdentification.ID.Value)
buyerScheme := doc.AccountingCustomerParty.Party.PartyIdentification.ID.SchemeID
buyerIDVal := strings.TrimSpace(doc.AccountingCustomerParty.Party.PartyIdentification.ID.Value)

invoice := &CanonicalInvoice{
ID:                 doc.ID,
InvoiceNumber:      doc.ID,
InvoiceTypeCode:    strings.TrimSpace(doc.InvoiceTypeCode),
IssueDate:          date,
Currency:           strings.TrimSpace(doc.DocumentCurrencyCode),
SourceSyntax:       "UBL-2.1",
TargetJurisdiction: jurisdiction,
Seller: Party{
Name:        doc.AccountingSupplierParty.Party.PartyLegalEntity.RegistrationName,
TaxID:       doc.AccountingSupplierParty.Party.PartyTaxScheme.CompanyID,
Country:     jurisdiction,
Identifiers: map[string]string{sellerScheme: sellerIDVal},
},
Buyer: Party{
Name:        doc.AccountingCustomerParty.Party.PartyLegalEntity.RegistrationName,
Country:     jurisdiction,
Identifiers: map[string]string{buyerScheme: buyerIDVal},
},
Totals: MonetaryTotals{
LineExtensionAmount: doc.LegalMonetaryTotal.LineExtensionAmount,
TaxExclusiveAmount:  doc.LegalMonetaryTotal.TaxExclusiveAmount,
TaxInclusiveAmount:  doc.LegalMonetaryTotal.TaxInclusiveAmount,
PayableAmount:       doc.LegalMonetaryTotal.PayableAmount,
},
Lines: make([]InvoiceLine, 0, len(doc.InvoiceLines)),
}

// Affectation des identifiants nationaux
if jurisdiction == "MA" || jurisdiction == "FR" {
invoice.Seller.NationalID = sellerIDVal
invoice.Buyer.NationalID = buyerIDVal
}

// Normalisation des lignes
for _, rawLine := range doc.InvoiceLines {
desc := rawLine.Item.Description
if desc == "" {
desc = rawLine.Item.Name
}
invoice.Lines = append(invoice.Lines, InvoiceLine{
ID:          rawLine.ID,
Description: desc,
Quantity:    rawLine.InvoicedQuantity,
UnitPrice:   rawLine.Price.PriceAmount,
LineTotal:   rawLine.LineExtensionAmount,
})
}

// Normalisation des sous-totaux fiscaux
for _, sub := range doc.TaxTotal.TaxSubtotal {
pct, _ := strconv.ParseFloat(sub.TaxCategory.Percent, 64)
invoice.TaxSubtotals = append(invoice.TaxSubtotals, TaxSubtotal{
TaxableAmount: sub.TaxableAmount.Value,
TaxAmount:     sub.TaxAmount.Value,
Percent:       pct,
CategoryCode:  sub.TaxCategory.ID,
})
}

return invoice, nil
}
