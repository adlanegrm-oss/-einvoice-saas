package exporter

import (
"strings"
"testing"

"github.com/shopspring/decimal"

"einvoice-saas/internal/model"
)

func TestFromCanonical_Mapping(t *testing.T) {
inv := &model.CanonicalInvoice{
ID:               "FAC-2026-001",
IssueDate:        "2026-10-08",
DocumentCurrency: "EUR",
Seller: model.Party{
Name:    "ACME SAS",
LegalID: "12345678901234",
VATID:   "FR12345678901",
},
Buyer: model.Party{Name: "Client SA"},
Totals: model.Totals{
TaxExclusiveAmount: decimal.NewFromFloat(1000),
TaxInclusiveAmount: decimal.NewFromFloat(1200),
TotalTaxAmount:     decimal.NewFromFloat(200),
},
}

meta := FromCanonical(inv, ProfileEN16931)
if meta.InvoiceNumber != "FAC-2026-001" {
t.Errorf("InvoiceNumber=%s", meta.InvoiceNumber)
}
if meta.SellerSIREN != "123456789" {
t.Errorf("SellerSIREN=%s want 123456789", meta.SellerSIREN)
}
if meta.SellerVAT != "FR12345678901" {
t.Errorf("SellerVAT=%s", meta.SellerVAT)
}
if meta.Currency != "EUR" {
t.Errorf("Currency=%s", meta.Currency)
}
if meta.TotalHT != 1000 || meta.TotalTTC != 1200 || meta.TotalTax != 200 {
t.Errorf("totals HT=%.2f TTC=%.2f Tax=%.2f", meta.TotalHT, meta.TotalTTC, meta.TotalTax)
}
if meta.IssueDate.Format("2006-01-02") != "2026-10-08" {
t.Errorf("IssueDate=%v", meta.IssueDate)
}
}

func TestGenerateFromCanonical_PDFA3Markers(t *testing.T) {
inv := &model.CanonicalInvoice{
ID:               "FAC-99",
IssueDate:        "2026-10-08",
DocumentCurrency: "EUR",
Seller:           model.Party{Name: "Seller", LegalID: "123456789", VATID: "FR12345678901"},
Buyer:            model.Party{Name: "Buyer"},
Totals: model.Totals{
TaxExclusiveAmount: decimal.NewFromFloat(100),
TaxInclusiveAmount: decimal.NewFromFloat(120),
TotalTaxAmount:     decimal.NewFromFloat(20),
},
}
cii := []byte(`<?xml version="1.0"?><CrossIndustryInvoice xmlns="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"><ExchangedDocument><ID>FAC-99</ID></ExchangedDocument></CrossIndustryInvoice>`)

pdf, err := GenerateFromCanonical(inv, cii, ProfileEN16931)
if err != nil {
t.Fatalf("GenerateFromCanonical: %v", err)
}
if err := VerifyPDFA3Conformance(pdf); err != nil {
t.Fatalf("VerifyPDFA3Conformance: %v", err)
}
raw := string(pdf)
if !strings.Contains(raw, "FAC-99") {
t.Error("PDF should contain invoice number")
}
}
