package exporter

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

// CrossIndustryInvoice modélise la structure XML Factur-X / CII
type CrossIndustryInvoice struct {
	XMLName  xml.Name `xml:"rsm:CrossIndustryInvoice"`
	XmlnsRSM string   `xml:"xmlns:rsm,attr"`
	XmlnsRAM string   `xml:"xmlns:ram,attr"`
	XmlnsUDT string   `xml:"xmlns:udt,attr"`

	ExchangedDocumentContext struct {
		GuidelineSpecifiedDocumentContextParameter struct {
			ID string `xml:"ram:ID"`
		} `xml:"ram:GuidelineSpecifiedDocumentContextParameter"`
	} `xml:"rsm:ExchangedDocumentContext"`

	ExchangedDocument struct {
		ID            string `xml:"ram:ID"`
		TypeCode      string `xml:"ram:TypeCode"`
		IssueDateTime struct {
			DateTimeString struct {
				Format string `xml:"format,attr"`
				Value  string `xml:",chardata"`
			} `xml:"udt:DateTimeString"`
		} `xml:"ram:IssueDateTime"`
	} `xml:"rsm:ExchangedDocument"`

	SupplyChainTradeTransaction struct {
		ApplicableHeaderTradeAgreement struct {
			BuyerTradeParty struct {
				Name string `xml:"ram:Name"`
			} `xml:"ram:BuyerTradeParty"`
		} `xml:"ram:ApplicableHeaderTradeAgreement"`
		ApplicableHeaderTradeSettlement struct {
			SpecifiedTradeSettlementHeaderMonetarySummation struct {
				TaxBasisTotalAmount string `xml:"ram:TaxBasisTotalAmount"`
				TaxTotalAmount      string `xml:"ram:TaxTotalAmount"`
				GrandTotalAmount    string `xml:"ram:GrandTotalAmount"`
			} `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
		} `xml:"ram:ApplicableHeaderTradeSettlement"`
	} `xml:"rsm:SupplyChainTradeTransaction"`
}

// GenerateFacturXXML génère un document XML conforme au profil Factur-X MINIMUM/BASIC
func GenerateFacturXXML(inv invoice.Invoice) ([]byte, error) {
	cii := CrossIndustryInvoice{
		XmlnsRSM: "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100",
		XmlnsRAM: "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100",
		XmlnsUDT: "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100",
	}

	cii.ExchangedDocumentContext.GuidelineSpecifiedDocumentContextParameter.ID = "urn:factur-x.eu:1p0:minimum"
	cii.ExchangedDocument.ID = inv.Number
	cii.ExchangedDocument.TypeCode = "380" // Code standard pour facture commerciale

	issueDate := inv.IssueDate
	if issueDate.IsZero() {
		issueDate = time.Now()
	}
	cii.ExchangedDocument.IssueDateTime.DateTimeString.Format = "102"
	cii.ExchangedDocument.IssueDateTime.DateTimeString.Value = issueDate.Format("20060102")

	cii.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.BuyerTradeParty.Name = inv.Customer
	cii.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement.SpecifiedTradeSettlementHeaderMonetarySummation.TaxBasisTotalAmount = fmt.Sprintf("%.2f", inv.TotalHT)
	cii.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement.SpecifiedTradeSettlementHeaderMonetarySummation.TaxTotalAmount = fmt.Sprintf("%.2f", inv.TotalVAT)
	cii.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement.SpecifiedTradeSettlementHeaderMonetarySummation.GrandTotalAmount = fmt.Sprintf("%.2f", inv.TotalTTC)

	output, err := xml.MarshalIndent(cii, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), output...), nil
}
