package ma

import (
"encoding/xml"
"strconv"
)

type ValidationReport struct {
Valid         bool               `json:"valid"`
Syntax        string             `json:"syntax"`
Profile       string             `json:"profile"`
ErrorsCount   int                `json:"errors_count"`
WarningsCount int                `json:"warnings_count"`
Results       []DiagnosticResult `json:"results"`
}

type PartyNode struct {
PartyIdentification struct {
ID struct {
Value    string `xml:",chardata"`
SchemeID string `xml:"schemeID,attr"`
} `xml:"ID"`
} `xml:"PartyIdentification"`
PartyTaxScheme struct {
CompanyID string `xml:"CompanyID"`
} `xml:"PartyTaxScheme"`
}

type TaxSubtotalNode struct {
TaxCategory struct {
Percent string `xml:"Percent"`
} `xml:"TaxCategory"`
}

type UBLInvoiceMA struct {
XMLName                  xml.Name `xml:"Invoice"`
CustomizationID          string   `xml:"CustomizationID"`
DocumentCurrencyCode     string   `xml:"DocumentCurrencyCode"`
AccountingSupplierParty struct {
Party PartyNode `xml:"Party"`
} `xml:"AccountingSupplierParty"`
AccountingCustomerParty struct {
Party PartyNode `xml:"Party"`
} `xml:"AccountingCustomerParty"`
TaxTotal struct {
TaxSubtotal []TaxSubtotalNode `xml:"TaxSubtotal"`
} `xml:"TaxTotal"`
}

// ValidateMoroccoUBL parse et applique les contrôles fiscaux marocains sur une facture UBL
func ValidateMoroccoUBL(xmlData []byte) ValidationReport {
report := ValidationReport{
Valid:   true,
Syntax:  "UBL-2.1",
Profile: "MAROC-DGI-1.0",
Results: make([]DiagnosticResult, 0),
}

var inv UBLInvoiceMA
if err := xml.Unmarshal(xmlData, &inv); err != nil {
report.Valid = false
report.ErrorsCount = 1
report.Results = append(report.Results, DiagnosticResult{
Code:         "MA-ERR-PARSE",
Severity:     "ERROR",
Path:         "/Invoice",
Message:      "Flux XML UBL invalide ou non décodable: " + err.Error(),
ExpectedRule: "XML UBL 2.1 syntaxiquement correct",
})
return report
}

// 1. Contrôle ICE Fournisseur
supplierICE := inv.AccountingSupplierParty.Party.PartyIdentification.ID.Value
if issue := ValidateICE(supplierICE, "AccountingSupplierParty.Party.PartyIdentification.ID"); issue != nil {
report.Results = append(report.Results, *issue)
}

// 2. Contrôle ICE Client
customerICE := inv.AccountingCustomerParty.Party.PartyIdentification.ID.Value
if issue := ValidateICE(customerICE, "AccountingCustomerParty.Party.PartyIdentification.ID"); issue != nil {
report.Results = append(report.Results, *issue)
}

// 3. Contrôle IF Fournisseur (PartyTaxScheme.CompanyID)
supplierIF := inv.AccountingSupplierParty.Party.PartyTaxScheme.CompanyID
if issue := ValidateIF(supplierIF, "AccountingSupplierParty.Party.PartyTaxScheme.CompanyID"); issue != nil {
report.Results = append(report.Results, *issue)
}

// 4. Contrôle des taux de TVA déclarés
for idx, sub := range inv.TaxTotal.TaxSubtotal {
if sub.TaxCategory.Percent != "" {
rate, err := strconv.ParseFloat(sub.TaxCategory.Percent, 64)
if err == nil {
if issue := ValidateVATRate(rate, strconv.Itoa(idx+1)); issue != nil {
report.Results = append(report.Results, *issue)
}
}
}
}

// Calcul des totaux et statut
for _, res := range report.Results {
if res.Severity == "ERROR" {
report.ErrorsCount++
report.Valid = false
} else if res.Severity == "WARNING" {
report.WarningsCount++
}
}

return report
}
