package integration

import (
	"bytes"
	"testing"

	"einvoice-saas/internal/exporter"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/validator"
)

func TestFacturXFullPipeline(t *testing.T) {
	ciiXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
                          xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">
<rsm:ExchangedDocumentContext>
<ram:GuidelineSpecifiedDocumentContextParameter>
<ram:ID>urn:cen.eu:en16931:2017</ram:ID>
</ram:GuidelineSpecifiedDocumentContextParameter>
</rsm:ExchangedDocumentContext>
<rsm:ExchangedDocument>
<ram:ID>INV-2026-PIPE-01</ram:ID>
<ram:TypeCode>380</ram:TypeCode>
<ram:IssueDateTime>
<ram:DateTimeString>20261005</ram:DateTimeString>
</ram:IssueDateTime>
</rsm:ExchangedDocument>
<rsm:SupplyChainTradeTransaction>
<ram:IncludedSupplyChainTradeLineItem>
<ram:AssociatedDocumentLineDocument>
<ram:LineID>1</ram:LineID>
</ram:AssociatedDocumentLineDocument>
<ram:SpecifiedTradeProduct>
<ram:Name>Développement Système</ram:Name>
</ram:SpecifiedTradeProduct>
<ram:SpecifiedLineTradeAgreement>
<ram:NetPriceProductTradePrice>
<ram:ChargeAmount>500.00</ram:ChargeAmount>
</ram:NetPriceProductTradePrice>
</ram:SpecifiedLineTradeAgreement>
<ram:SpecifiedLineTradeDelivery>
<ram:BilledQuantity unitCode="C62">2.0</ram:BilledQuantity>
</ram:SpecifiedLineTradeDelivery>
<ram:SpecifiedLineTradeSettlement>
<ram:ApplicableTradeTax>
<ram:CategoryCode>S</ram:CategoryCode>
<ram:RateApplicablePercent>20.00</ram:RateApplicablePercent>
</ram:ApplicableTradeTax>
<ram:SpecifiedTradeSettlementLineMonetarySummation>
<ram:LineTotalAmount>1000.00</ram:LineTotalAmount>
</ram:SpecifiedTradeSettlementLineMonetarySummation>
</ram:SpecifiedLineTradeSettlement>
</ram:IncludedSupplyChainTradeLineItem>
<ram:ApplicableHeaderTradeAgreement>
<ram:SellerTradeParty>
<ram:Name>Vendeur Solutions</ram:Name>
<ram:PostalTradeAddress><ram:CountryID>FR</ram:CountryID></ram:PostalTradeAddress>
</ram:SellerTradeParty>
<ram:BuyerTradeParty>
<ram:Name>Acheteur Global</ram:Name>
<ram:PostalTradeAddress><ram:CountryID>FR</ram:CountryID></ram:PostalTradeAddress>
</ram:BuyerTradeParty>
</ram:ApplicableHeaderTradeAgreement>
<ram:ApplicableHeaderTradeSettlement>
<ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
<ram:SpecifiedTradePaymentTerms>
<ram:DueDateDateTime><ram:DateTimeString>20261105</ram:DateTimeString></ram:DueDateDateTime>
</ram:SpecifiedTradePaymentTerms>
<ram:ApplicableTradeTax>
<ram:BasisAmount>1000.00</ram:BasisAmount>
<ram:RateApplicablePercent>20.00</ram:RateApplicablePercent>
<ram:CalculatedAmount>200.00</ram:CalculatedAmount>
<ram:CategoryCode>S</ram:CategoryCode>
</ram:ApplicableTradeTax>
<ram:SpecifiedTradeSettlementHeaderMonetarySummation>
<ram:LineTotalAmount>1000.00</ram:LineTotalAmount>
<ram:TaxBasisTotalAmount>1000.00</ram:TaxBasisTotalAmount>
<ram:TaxTotalAmount>200.00</ram:TaxTotalAmount>
<ram:GrandTotalAmount>1200.00</ram:GrandTotalAmount>
<ram:DuePayableAmount>1200.00</ram:DuePayableAmount>
</ram:SpecifiedTradeSettlementHeaderMonetarySummation>
</ram:ApplicableHeaderTradeSettlement>
</rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`)

	canonical, err := model.NormalizeCIIToCanonical(ciiXML)
	if err != nil {
		t.Fatalf("échec normalisation : %v", err)
	}

	val := validator.NewNormativeValidator(false)
	result, err := val.ValidateEN16931AndPeppol(ciiXML)
	if err != nil || !result.Valid {
		t.Fatalf("échec validation : %v, errors: %+v", err, result.RuleErrors)
	}

	meta := exporter.InvoiceMetadata{
		InvoiceNumber: canonical.InvoiceNumber,
		SellerName:    canonical.Seller.Name,
		BuyerName:     canonical.Buyer.Name,
		IssueDate:     canonical.IssueDate,
		Currency:      canonical.Currency,
		TotalHT:       canonical.Totals.TaxExclusiveAmount,
		TotalTTC:      canonical.Totals.TaxInclusiveAmount,
		Profile:       exporter.ProfileEN16931,
	}

	pdfData, err := exporter.GenerateFacturXPDFA3(meta, ciiXML)
	if err != nil {
		t.Fatalf("échec génération Factur-X : %v", err)
	}

	if err := exporter.VerifyFacturXContainer(pdfData); err != nil {
		t.Fatalf("échec vérification Factur-X : %v", err)
	}

	if !bytes.Contains(pdfData, ciiXML) {
		t.Fatal("le XML embarqué ne correspond pas au XML source")
	}
}
