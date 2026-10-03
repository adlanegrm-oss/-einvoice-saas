package facturae

import (
	"bytes"
	"fmt"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

func GenerateXML(inv invoice.Invoice) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.WriteString(`<fe:Facturae xmlns:fe="http://www.facturae.gob.es/formato/Versiones/Facturaev3_2_2.xml">` + "\n")
	buf.WriteString("  <FileHeader>\n    <SchemaVersion>3.2.2</SchemaVersion>\n    <Modality>I</Modality>\n    <InvoiceIssuerType>EM</InvoiceIssuerType>\n  </FileHeader>\n")
	buf.WriteString("  <Parties>\n")
	buf.WriteString(fmt.Sprintf("    <SellerParty><TaxIdentification><TaxIdentificationNumber>%s</TaxIdentificationNumber></TaxIdentification><LegalEntity><CorporateName>%s</CorporateName></LegalEntity></SellerParty>\n", inv.Seller.SIRET, inv.Seller.Name))
	buf.WriteString(fmt.Sprintf("    <BuyerParty><TaxIdentification><TaxIdentificationNumber>%s</TaxIdentificationNumber></TaxIdentification><LegalEntity><CorporateName>%s</CorporateName></LegalEntity></BuyerParty>\n", inv.Customer.SIRET, inv.Customer.Name))
	buf.WriteString("  </Parties>\n")
	buf.WriteString("  <Invoices>\n    <Invoice>\n      <InvoiceHeader>\n")
	buf.WriteString(fmt.Sprintf("        <InvoiceNumber>%s</InvoiceNumber><InvoiceDocumentType>FC</InvoiceDocumentType>\n", inv.Number))
	buf.WriteString("      </InvoiceHeader>\n")
	buf.WriteString(fmt.Sprintf("      <InvoiceIssueData><IssueDate>%s</IssueDate></InvoiceIssueData>\n", inv.IssueDate.Format("2006-01-02")))
	buf.WriteString("      <InvoiceTotals>\n")
	buf.WriteString(fmt.Sprintf("        <TotalGrossAmountBeforeTaxes>%.2f</TotalGrossAmountBeforeTaxes>\n", inv.TotalHT.ToFloat()))
	buf.WriteString(fmt.Sprintf("        <TotalTaxOutputs>%.2f</TotalTaxOutputs>\n", inv.TotalVAT.ToFloat()))
	buf.WriteString(fmt.Sprintf("        <InvoiceTotal>%.2f</InvoiceTotal>\n", inv.TotalTTC.ToFloat()))
	buf.WriteString("      </InvoiceTotals>\n      <Items>\n")
	for _, it := range inv.Items {
		buf.WriteString("        <InvoiceLine>\n")
		buf.WriteString(fmt.Sprintf("          <ItemDescription>%s</ItemDescription>\n", it.Description))
		buf.WriteString(fmt.Sprintf("          <Quantity>%.2f</Quantity>\n", it.Quantity))
		buf.WriteString(fmt.Sprintf("          <UnitPriceWithoutTax>%.2f</UnitPriceWithoutTax>\n", it.UnitPrice.ToFloat()))
		buf.WriteString("          <TaxesOutputs><Tax><TaxTypeCode>01</TaxTypeCode>\n")
		buf.WriteString(fmt.Sprintf("            <TaxRate>%.2f</TaxRate>\n", it.VATRate.ToFloat()))
		buf.WriteString("          </Tax></TaxesOutputs>\n        </InvoiceLine>\n")
	}
	buf.WriteString("      </Items>\n    </Invoice>\n  </Invoices>\n</fe:Facturae>")
	return buf.Bytes(), nil
}
