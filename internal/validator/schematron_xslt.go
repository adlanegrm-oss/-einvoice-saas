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

// Structure UBL pour les taxes
type ublInvoiceSummary struct {
XMLName  xml.Name `xml:"Invoice"`
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

// Structure CII pour les taxes
type ciiInvoiceSummary struct {
XMLName xml.Name `xml:"CrossIndustryInvoice"`
SupplyChainTradeTransaction struct {
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
var breakdown []taxBreakdownItem

// 1. Essai de détection UBL
var ubl invWrapper
if err := xml.Unmarshal(xmlInput, &ubl); err == nil && ubl.isUBL() {
var ublDoc ublInvoiceSummary
if errUBL := xml.Unmarshal(xmlInput, &ublDoc); errUBL == nil {
for i, tt := range ublDoc.TaxTotal {
for j, sub := range tt.TaxSubtotal {
basis, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxableAmount.Value), 64)
tax, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxAmount.Value), 64)
rate, _ := strconv.ParseFloat(strings.TrimSpace(sub.TaxCategory.Percent.Value), 64)
loc := fmt.Sprintf("/*[local-name()='Invoice'][1]/*[local-name()='TaxTotal'][%d]/*[local-name()='TaxSubtotal'][%d]", i+1, j+1)
breakdown = append(breakdown, taxBreakdownItem{Basis: basis, Rate: rate, Tax: tax, Xpath: loc})
}
}
}
} else {
// 2. Détection CII
var ciiDoc ciiInvoiceSummary
if errCII := xml.Unmarshal(xmlInput, &ciiDoc); errCII == nil {
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

var failedAsserts []string
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

type invWrapper struct {
XMLName xml.Name
}

func (w *invWrapper) isUBL() bool {
return w.XMLName.Local == "Invoice"
}
