package validator

import (
	"testing"
)

func TestNormativeValidator_ValidInvoice(t *testing.T) {
	xmlSample := `<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100" xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">
  <rsm:ExchangedDocumentContext>
    <ram:GuidelineSpecifiedDocumentContextParameter>
      <ram:ID>urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:en16931</ram:ID>
    </ram:GuidelineSpecifiedDocumentContextParameter>
  </rsm:ExchangedDocumentContext>
  <rsm:ExchangedDocument>
    <ram:ID>INV-2026-001</ram:ID>
    <ram:TypeCode>380</ram:TypeCode>
    <ram:IssueDateTime>
      <udt:DateTimeString format="102" xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100">20261004</udt:DateTimeString>
    </ram:IssueDateTime>
  </rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:IncludedSupplyChainTradeLineItem>
      <ram:AssociatedDocumentLineDocument>
        <ram:LineID>1</ram:LineID>
      </ram:AssociatedDocumentLineDocument>
      <ram:SpecifiedLineTradeSettlement>
        <ram:SpecifiedTradeSettlementLineMonetarySummation>
          <ram:LineTotalAmount>1000.00</ram:LineTotalAmount>
        </ram:SpecifiedTradeSettlementLineMonetarySummation>
      </ram:SpecifiedLineTradeSettlement>
    </ram:IncludedSupplyChainTradeLineItem>
    <ram:ApplicableHeaderTradeAgreement>
      <ram:SellerTradeParty>
        <ram:Name>Fournisseur SAS</ram:Name>
        <ram:SpecifiedLegalOrganization>
          <ram:ID>73204903600045</ram:ID>
        </ram:SpecifiedLegalOrganization>
        <ram:PostalTradeAddress>
          <ram:CountryID>FR</ram:CountryID>
        </ram:PostalTradeAddress>
      </ram:SellerTradeParty>
      <ram:BuyerTradeParty>
        <ram:Name>Client SAS</ram:Name>
        <ram:PostalTradeAddress>
          <ram:CountryID>FR</ram:CountryID>
        </ram:PostalTradeAddress>
      </ram:BuyerTradeParty>
    </ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:ApplicableTradeTax>
        <ram:BasisAmount>1000.00</ram:BasisAmount>
        <ram:CalculatedAmount>200.00</ram:CalculatedAmount>
      </ram:ApplicableTradeTax>
      <ram:SpecifiedTradePaymentTerms>
        <ram:DueDateDateTime>
          <udt:DateTimeString format="102" xmlns:udt="urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100">20261104</udt:DateTimeString>
        </ram:DueDateDateTime>
      </ram:SpecifiedTradePaymentTerms>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>1000.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>1000.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>200.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>1200.00</ram:GrandTotalAmount>
        <ram:DuePayableAmount>1200.00</ram:DuePayableAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	validator := NewNormativeValidator(false)
	result, err := validator.ValidateEN16931AndPeppol([]byte(xmlSample))
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("Expected valid, got errors: %+v", result.RuleErrors)
	}
}

func TestNormativeValidator_RejectsMathInconsistency(t *testing.T) {
	xmlCorrupted := `<?xml version="1.0" encoding="UTF-8"?>
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100" xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">
  <rsm:ExchangedDocumentContext>
    <ram:GuidelineSpecifiedDocumentContextParameter><ram:ID>urn:cen.eu:en16931:2017</ram:ID></ram:GuidelineSpecifiedDocumentContextParameter>
  </rsm:ExchangedDocumentContext>
  <rsm:ExchangedDocument><ram:ID>INV-ERR-01</ram:ID><ram:TypeCode>380</ram:TypeCode><ram:IssueDateTime><udt:DateTimeString>20261004</udt:DateTimeString></ram:IssueDateTime></rsm:ExchangedDocument>
  <rsm:SupplyChainTradeTransaction>
    <ram:IncludedSupplyChainTradeLineItem><ram:AssociatedDocumentLineDocument><ram:LineID>1</ram:LineID></ram:AssociatedDocumentLineDocument><ram:SpecifiedLineTradeSettlement><ram:SpecifiedTradeSettlementLineMonetarySummation><ram:LineTotalAmount>100.00</ram:LineTotalAmount></ram:SpecifiedTradeSettlementLineMonetarySummation></ram:SpecifiedLineTradeSettlement></ram:IncludedSupplyChainTradeLineItem>
    <ram:ApplicableHeaderTradeAgreement><ram:SellerTradeParty><ram:Name>Vendeur</ram:Name><ram:PostalTradeAddress><ram:CountryID>FR</ram:CountryID></ram:PostalTradeAddress></ram:SellerTradeParty><ram:BuyerTradeParty><ram:Name>Acheteur</ram:Name><ram:PostalTradeAddress><ram:CountryID>FR</ram:CountryID></ram:PostalTradeAddress></ram:BuyerTradeParty></ram:ApplicableHeaderTradeAgreement>
    <ram:ApplicableHeaderTradeSettlement>
      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
      <ram:ApplicableTradeTax><ram:BasisAmount>100.00</ram:BasisAmount><ram:CalculatedAmount>20.00</ram:CalculatedAmount></ram:ApplicableTradeTax>
      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>
        <ram:LineTotalAmount>100.00</ram:LineTotalAmount>
        <ram:TaxBasisTotalAmount>100.00</ram:TaxBasisTotalAmount>
        <ram:TaxTotalAmount>20.00</ram:TaxTotalAmount>
        <ram:GrandTotalAmount>999.00</ram:GrandTotalAmount>
      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>
    </ram:ApplicableHeaderTradeSettlement>
  </rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`

	validator := NewNormativeValidator(false)
	result, err := validator.ValidateEN16931AndPeppol([]byte(xmlCorrupted))
	if err != nil {
		t.Fatalf("Unexpected execution error: %v", err)
	}
	if result.Valid {
		t.Fatal("Expected validation to fail on corrupt grand total (BR-CO-15)")
	}

	foundBRCO15 := false
	for _, r := range result.RuleErrors {
		if r.RuleID == "BR-CO-15" {
			foundBRCO15 = true
		}
	}
	if !foundBRCO15 {
		t.Errorf("Expected BR-CO-15 violation, got: %+v", result.RuleErrors)
	}
}
