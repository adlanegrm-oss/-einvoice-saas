package validator

import (
"bytes"
"encoding/xml"
"fmt"
"math"
"os/exec"
"strconv"
"strings"
)

type XSLTExecutor interface {
Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error)
}

type CLIExecutor struct {
BinaryPath string
}

func NewCLIExecutor(binPath string) *CLIExecutor {
if binPath == "" {
binPath = "xsltproc"
}
return &CLIExecutor{BinaryPath: binPath}
}

func (c *CLIExecutor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
cmd := exec.Command(c.BinaryPath, "-", "-")
cmd.Stdin = bytes.NewReader(append(xsltSheet, xmlInput...))

var stdout bytes.Buffer
var stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

if err := cmd.Run(); err != nil {
return nil, fmt.Errorf("cli xslt: %w (stderr: %s)", err, stderr.String())
}

return stdout.Bytes(), nil
}

type MockXSLTExecutor struct {
MockOutput []byte
MockErr    error
}

func (m *MockXSLTExecutor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
if m.MockErr != nil {
return nil, m.MockErr
}
if len(m.MockOutput) > 0 {
return m.MockOutput, nil
}
return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<schematron-output xmlns="http://purl.oclc.org/dsdl/svrl" title="Validation EN 16931">
</schematron-output>`), nil
}

type NativeEN16931Executor struct{}

func NewDefaultXSLTExecutor() XSLTExecutor {
return &NativeEN16931Executor{}
}

type ublInvoiceDoc struct {
XMLName         xml.Name `xml:"Invoice"`
CustomizationID struct {
Value string `xml:",chardata"`
} `xml:"CustomizationID"`
AccountingSupplierParty struct {
Party struct {
PartyTaxScheme []struct {
CompanyID struct {
Value string `xml:",chardata"`
} `xml:"CompanyID"`
} `xml:"PartyTaxScheme"`
} `xml:"Party"`
} `xml:"AccountingSupplierParty"`
TaxTotal []struct {
TaxSubtotal []struct {
TaxableAmount struct {
Value string `xml:",chardata"`
} `xml:"TaxableAmount"`
TaxAmount struct {
Value string `xml:",chardata"`
} `xml:"TaxAmount"`
TaxCategory struct {
Percent struct {
Value string `xml:",chardata"`
} `xml:"Percent"`
} `xml:"TaxCategory"`
} `xml:"TaxSubtotal"`
} `xml:"TaxTotal"`
}

type ciiInvoiceDoc struct {
XMLName xml.Name `xml:"CrossIndustryInvoice"`
ExchangedDocumentContext struct {
GuidelineSpecifiedDocumentContextParameter struct {
ID struct {
Value string `xml:",chardata"`
} `xml:"ID"`
} `xml:"GuidelineSpecifiedDocumentContextParameter"`
} `xml:"ExchangedDocumentContext"`
SupplyChainTradeTransaction struct {
ApplicableHeaderTradeAgreement struct {
SellerTradeParty struct {
SpecifiedTaxRegistration []struct {
ID struct {
Value string `xml:",chardata"`
} `xml:"ID"`
} `xml:"SpecifiedTaxRegistration"`
} `xml:"SellerTradeParty"`
} `xml:"ApplicableHeaderTradeAgreement"`
ApplicableHeaderTradeSettlement struct {
ApplicableTradeTax []struct {
CalculatedAmount struct {
Value string `xml:",chardata"`
} `xml:"CalculatedAmount"`
BasisAmount struct {
Value string `xml:",chardata"`
} `xml:"BasisAmount"`
RateApplicablePercent struct {
Value string `xml:",chardata"`
} `xml:"RateApplicablePercent"`
} `xml:"ApplicableTradeTax"`
} `xml:"ApplicableHeaderTradeSettlement"`
} `xml:"SupplyChainTradeTransaction"`
}

type taxBreakdownItem struct {
Basis float64
Rate  float64
Tax   float64
Xpath string
}

func (n *NativeEN16931Executor) Transform(xmlInput []byte, xsltSheet []byte) ([]byte, error) {
var failedAsserts []string
var customID string
var sellerTaxID string
var breakdown []taxBreakdownItem
var isDocFound bool

var ublDoc ublInvoiceDoc
if err := xml.Unmarshal(xmlInput, &ublDoc); err == nil && ublDoc.XMLName.Local == "Invoice" {
isDocFound = true
customID = strings.TrimSpace(ublDoc.CustomizationID.Value)
for _, pts := range ublDoc.AccountingSupplierParty.Party.PartyTaxScheme {
if val := strings.TrimSpace(pts.CompanyID.Value); val != "" {
sellerTaxID = val
break
}
}
for i, tt := range ublDoc.TaxTotal {
for j, sub := range tt.TaxSubtotal {
basis, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxableAmount.Value), 64)
tax, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxAmount.Value), 64)
rate, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxCategory.Percent.Value), 64)
loc := fmt.Sprintf("/*[local-name()='Invoice'][1]/*[local-name()='TaxTotal'][%d]/*[local-name()='TaxSubtotal'][%d]", i+1, j+1)
breakdown = append(breakdown, taxBreakdownItem{Basis: basis, Rate: rate, Tax: tax, Xpath: loc})
}
}
} else {
var ciiDoc ciiInvoiceDoc
if errCII := xml.Unmarshal(xmlInput, &ciiDoc); errCII == nil && ciiDoc.XMLName.Local == "CrossIndustryInvoice" {
isDocFound = true
customID = strings.TrimSpace(ciiDoc.ExchangedDocumentContext.GuidelineSpecifiedDocumentContextParameter.ID.Value)
sellerParty := ciiDoc.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.SellerTradeParty
for _, reg := range sellerParty.SpecifiedTaxRegistration {
if val := strings.TrimSpace(reg.ID.Value); val != "" {
sellerTaxID = val
break
}
}
settlement := ciiDoc.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement
for idx, tax := range settlement.ApplicableTradeTax {
rate, _ := strconv.ParseFloat(strings.TrimSpace(tax.RateApplicablePercent.Value), 64)
basis, _ := strconv.ParseFloat(strings.TrimSpace(tax.BasisAmount.Value), 64)
calc, _ := strconv.ParseFloat(strings.TrimSpace(tax.CalculatedAmount.Value), 64)
loc := fmt.Sprintf("/*[local-name()='CrossIndustryInvoice'][1]/*[local-name()='SupplyChainTradeTransaction'][1]/*[local-name()='ApplicableHeaderTradeSettlement'][1]/*[local-name()='ApplicableTradeTax'][%d]", idx+1)
breakdown = append(breakdown, taxBreakdownItem{Basis: basis, Rate: rate, Tax: calc, Xpath: loc})
}
}
}

if !isDocFound {
return nil, fmt.Errorf("native executor: format non reconnu (ni UBL Invoice, ni UN/CEFACT CII)")
}

// 1. Règle BR-01 : Spécification du CustomizationID / CIUS
if customID == "" || (!strings.Contains(customID, "en16931") && !strings.Contains(customID, "16931")) {
assertXML := `  <failed-assert test="CustomizationID != '' and contains(CustomizationID, 'en16931')" id="BR-01" flag="fatal" location="/*[1]/*[local-name()='CustomizationID' or local-name()='ExchangedDocumentContext'][1]">
    <text>[BR-01]-An Invoice that claims compliance with EN16931 must specify the specification identifier (CustomizationID).</text>
  </failed-assert>`
failedAsserts = append(failedAsserts, assertXML)
}

// 2. Règle BR-CO-09 : Présence du numéro d'identification TVA du vendeur
if sellerTaxID == "" {
assertXML := `  <failed-assert test="exists(cac:PartyTaxScheme/cbc:CompanyID)" id="BR-CO-09" flag="fatal" location="/*[1]/*[local-name()='AccountingSupplierParty'][1]/*[local-name()='Party'][1]">
    <text>[BR-CO-09]-The Seller VAT identifier (BT-31), the Seller tax representative VAT identifier (BT-63) or the Seller tax registration identifier (BT-32) shall be present.</text>
  </failed-assert>`
failedAsserts = append(failedAsserts, assertXML)
}

// 3. Règle BR-CO-14 : Présence obligatoire d'au moins une ventilation
if len(breakdown) == 0 {
assertXML := `  <failed-assert test="exists(TaxSubtotal) or exists(ApplicableTradeTax)" id="BR-CO-14" flag="fatal" location="/*[1]">
    <text>[BR-CO-14]-Invoice must have at least one VAT breakdown item.</text>
  </failed-assert>`
failedAsserts = append(failedAsserts, assertXML)
}

// 4. Règle BR-CO-17 : Cohérence mathématique de la ventilation
for _, item := range breakdown {
if item.Basis > 0 && item.Rate > 0 {
expectedTax := math.Round((item.Basis*(item.Rate/100.0))*100) / 100
diff := math.Abs(item.Tax - expectedTax)
if diff > 0.05 {
msg := fmt.Sprintf("[BR-CO-17]-The VAT category tax amount (%0.2f) in a VAT breakdown must equal the VAT category taxable amount (%0.2f) multiplied by the VAT category rate (%0.2f%%).", item.Tax, item.Basis, item.Rate)
assertXML := fmt.Sprintf(`  <failed-assert test="abs(TaxAmount - (TaxableAmount * Rate div 100)) &lt;= 0.05" id="BR-CO-17" flag="fatal" location="%s">
    <text>%s</text>
  </failed-assert>`, item.Xpath, msg)
failedAsserts = append(failedAsserts, assertXML)
}
}
}

var buf bytes.Buffer
buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
buf.WriteString("<schematron-output xmlns=\"http://purl.oclc.org/dsdl/svrl\" title=\"Validation EN 16931\">\n")
for _, fa := range failedAsserts {
buf.WriteString(fa + "\n")
}
buf.WriteString("</schematron-output>")

return buf.Bytes(), nil
}
