package model

import (
	"testing"
	"time"
)

func TestParseCIIDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"YYYYMMDD", "20261005", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), false},
		{"YYYY-MM-DD", "2026-10-05", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), false},
		{"Timestamp", "20261005143000", time.Date(2026, 10, 5, 14, 30, 0, 0, time.UTC), false},
		{"ISO8601", "2026-10-05T14:30:00Z", time.Date(2026, 10, 5, 14, 30, 0, 0, time.UTC), false},
		{"Chaine vide", "", time.Time{}, false},
		{"Date invalide", "NOT_A_DATE", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCIIDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCIIDate(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseCIIDate(%q) = %v, want = %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeCIIToCanonical_Validation(t *testing.T) {
	t.Run("Rejette date emission invalide", func(t *testing.T) {
		xmlDoc := []byte(`
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
                          xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">
<rsm:ExchangedDocument>
<ram:ID>INV-ERR-DATE</ram:ID>
<ram:IssueDateTime>
<ram:DateTimeString>2026-99-99</ram:DateTimeString>
</ram:IssueDateTime>
</rsm:ExchangedDocument>
</rsm:CrossIndustryInvoice>`)

		_, err := NormalizeCIIToCanonical(xmlDoc)
		if err == nil {
			t.Fatal("attendu : echec sur date invalide")
		}
	})

	t.Run("Rejette taux TVA non numerique", func(t *testing.T) {
		xmlDoc := []byte(`
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
                          xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">
<rsm:ExchangedDocument>
<ram:ID>INV-ERR-TAX</ram:ID>
<ram:IssueDateTime>
<ram:DateTimeString>20261005</ram:DateTimeString>
</ram:IssueDateTime>
</rsm:ExchangedDocument>
<rsm:SupplyChainTradeTransaction>
<ram:IncludedSupplyChainTradeLineItem>
<ram:AssociatedDocumentLineDocument>
<ram:LineID>1</ram:LineID>
</ram:AssociatedDocumentLineDocument>
<ram:SpecifiedLineTradeSettlement>
<ram:ApplicableTradeTax>
<ram:CategoryCode>S</ram:CategoryCode>
<ram:RateApplicablePercent>VINGT_POURCENT</ram:RateApplicablePercent>
</ram:ApplicableTradeTax>
</ram:SpecifiedLineTradeSettlement>
</ram:IncludedSupplyChainTradeLineItem>
</rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`)

		_, err := NormalizeCIIToCanonical(xmlDoc)
		if err == nil {
			t.Fatal("attendu : echec sur taux TVA invalide")
		}
	})

	t.Run("Normalisation complete avec DueDate et TVA Ligne", func(t *testing.T) {
		xmlDoc := []byte(`
<rsm:CrossIndustryInvoice xmlns:rsm="urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
                          xmlns:ram="urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100">
<rsm:ExchangedDocument>
<ram:ID>INV-2026-OK</ram:ID>
<ram:IssueDateTime>
<ram:DateTimeString>20261005</ram:DateTimeString>
</ram:IssueDateTime>
</rsm:ExchangedDocument>
<rsm:SupplyChainTradeTransaction>
<ram:IncludedSupplyChainTradeLineItem>
<ram:AssociatedDocumentLineDocument>
<ram:LineID>LINE-01</ram:LineID>
</ram:AssociatedDocumentLineDocument>
<ram:SpecifiedTradeProduct>
<ram:Name>Licence Logicielle</ram:Name>
</ram:SpecifiedTradeProduct>
<ram:SpecifiedLineTradeAgreement>
<ram:NetPriceProductTradePrice>
<ram:ChargeAmount>250.00</ram:ChargeAmount>
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
<ram:LineTotalAmount>500.00</ram:LineTotalAmount>
</ram:SpecifiedTradeSettlementLineMonetarySummation>
</ram:SpecifiedLineTradeSettlement>
</ram:IncludedSupplyChainTradeLineItem>
<ram:ApplicableHeaderTradeAgreement>
<ram:SellerTradeParty>
<ram:Name>Fournisseur SAS</ram:Name>
<ram:SpecifiedLegalOrganization>
<ram:ID>123456789</ram:ID>
</ram:SpecifiedLegalOrganization>
<ram:PostalTradeAddress>
<ram:CountryID>FR</ram:CountryID>
</ram:PostalTradeAddress>
</ram:SellerTradeParty>
<ram:BuyerTradeParty>
<ram:Name>Client SARL</ram:Name>
<ram:PostalTradeAddress>
<ram:CountryID>FR</ram:CountryID>
</ram:PostalTradeAddress>
</ram:BuyerTradeParty>
</ram:ApplicableHeaderTradeAgreement>
<ram:ApplicableHeaderTradeSettlement>
<ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>
<ram:SpecifiedTradePaymentTerms>
<ram:DueDateDateTime>
<ram:DateTimeString>20261115</ram:DateTimeString>
</ram:DueDateDateTime>
</ram:SpecifiedTradePaymentTerms>
<ram:ApplicableTradeTax>
<ram:BasisAmount>500.00</ram:BasisAmount>
<ram:RateApplicablePercent>20.00</ram:RateApplicablePercent>
<ram:CalculatedAmount>100.00</ram:CalculatedAmount>
<ram:CategoryCode>S</ram:CategoryCode>
</ram:ApplicableTradeTax>
<ram:SpecifiedTradeSettlementHeaderMonetarySummation>
<ram:LineTotalAmount>500.00</ram:LineTotalAmount>
<ram:TaxBasisTotalAmount>500.00</ram:TaxBasisTotalAmount>
<ram:TaxTotalAmount>100.00</ram:TaxTotalAmount>
<ram:GrandTotalAmount>600.00</ram:GrandTotalAmount>
<ram:DuePayableAmount>600.00</ram:DuePayableAmount>
</ram:SpecifiedTradeSettlementHeaderMonetarySummation>
</ram:ApplicableHeaderTradeSettlement>
</rsm:SupplyChainTradeTransaction>
</rsm:CrossIndustryInvoice>`)

		inv, err := NormalizeCIIToCanonical(xmlDoc)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}

		if inv.DueDate == nil {
			t.Fatal("DueDate est nil, attendu: 2026-11-15")
		}
		if inv.DueDate.Format("2006-01-02") != "2026-11-15" {
			t.Errorf("DueDate = %s, attendu: 2026-11-15", inv.DueDate.Format("2006-01-02"))
		}

		if len(inv.Lines) != 1 {
			t.Fatalf("lignes = %d, attendu: 1", len(inv.Lines))
		}
		line := inv.Lines[0]
		if line.VatPercent != 20.00 || line.VatCategory != "S" {
			t.Errorf("TVA Ligne incorrecte: %+v", line)
		}
		if line.LineTotal != 500.00 {
			t.Errorf("Total ligne = %.2f, attendu: 500.00", line.LineTotal)
		}

		if inv.Totals.TaxInclusiveAmount != 600.00 {
			t.Errorf("TTC = %.2f, attendu: 600.00", inv.Totals.TaxInclusiveAmount)
		}
	})
}
