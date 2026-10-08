package model

import (
	"encoding/xml"
	"strings"
)

// NormalizeCIIToCanonical convertit un flux CII (Factur-X / ZUGFeRD) en CanonicalInvoice
// Version minimale pour débloquer le build — à enrichir ensuite.
func NormalizeCIIToCanonical(xmlPayload []byte) (*CanonicalInvoice, error) {
	type ciiDoc struct {
		XMLName           xml.Name `xml:"CrossIndustryInvoice"`
		ExchangedDocument struct {
			ID            string `xml:"ID"`
			TypeCode      string `xml:"TypeCode"`
			IssueDateTime struct {
				DateTimeString string `xml:"DateTimeString"`
			} `xml:"IssueDateTime"`
		} `xml:"ExchangedDocument"`
		SupplyChainTradeTransaction struct {
			ApplicableHeaderTradeAgreement struct {
				BuyerReference   string `xml:"BuyerReference"`
				SellerTradeParty struct {
					Name                       string `xml:"Name"`
					SpecifiedLegalOrganization struct {
						ID string `xml:"ID"`
					} `xml:"SpecifiedLegalOrganization"`
					PostalTradeAddress struct {
						CountryID string `xml:"CountryID"`
					} `xml:"PostalTradeAddress"`
				} `xml:"SellerTradeParty"`
				BuyerTradeParty struct {
					Name                       string `xml:"Name"`
					SpecifiedLegalOrganization struct {
						ID string `xml:"ID"`
					} `xml:"SpecifiedLegalOrganization"`
					PostalTradeAddress struct {
						CountryID string `xml:"CountryID"`
					} `xml:"PostalTradeAddress"`
				} `xml:"BuyerTradeParty"`
			} `xml:"ApplicableHeaderTradeAgreement"`
			ApplicableHeaderTradeSettlement struct {
				InvoiceCurrencyCode                             string `xml:"InvoiceCurrencyCode"`
				SpecifiedTradeSettlementHeaderMonetarySummation struct {
					LineTotalAmount     string `xml:"LineTotalAmount"`
					TaxBasisTotalAmount string `xml:"TaxBasisTotalAmount"`
					TaxTotalAmount      string `xml:"TaxTotalAmount"`
					GrandTotalAmount    string `xml:"GrandTotalAmount"`
					DuePayableAmount    string `xml:"DuePayableAmount"`
				} `xml:"SpecifiedTradeSettlementHeaderMonetarySummation"`
			} `xml:"ApplicableHeaderTradeSettlement"`
		} `xml:"SupplyChainTradeTransaction"`
	}

	var doc ciiDoc
	if err := xml.Unmarshal(xmlPayload, &doc); err != nil {
		return nil, err
	}

	issueDate := strings.TrimSpace(doc.ExchangedDocument.IssueDateTime.DateTimeString)
	if len(issueDate) >= 8 && !strings.Contains(issueDate, "-") {
		// format CCYYMMDD → CCYY-MM-DD
		issueDate = issueDate[0:4] + "-" + issueDate[4:6] + "-" + issueDate[6:8]
	}

	currency := strings.TrimSpace(doc.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement.InvoiceCurrencyCode)
	sum := doc.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement.SpecifiedTradeSettlementHeaderMonetarySummation

	lineTotal, _ := ParseDecimal(sum.LineTotalAmount)
	taxExcl, _ := ParseDecimal(sum.TaxBasisTotalAmount)
	taxTotal, _ := ParseDecimal(sum.TaxTotalAmount)
	taxIncl, _ := ParseDecimal(sum.GrandTotalAmount)
	payable, _ := ParseDecimal(sum.DuePayableAmount)

	seller := doc.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.SellerTradeParty
	buyer := doc.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.BuyerTradeParty

	inv := &CanonicalInvoice{
		ID:               strings.TrimSpace(doc.ExchangedDocument.ID),
		IssueDate:        issueDate,
		TypeCode:         strings.TrimSpace(doc.ExchangedDocument.TypeCode),
		DocumentCurrency: currency,
		TaxCurrency:      currency,
		BuyerReference:   strings.TrimSpace(doc.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement.BuyerReference),
		Seller: Party{
			Name:        seller.Name,
			LegalID:     strings.TrimSpace(seller.SpecifiedLegalOrganization.ID),
			CountryCode: seller.PostalTradeAddress.CountryID,
		},
		Buyer: Party{
			Name:        buyer.Name,
			LegalID:     strings.TrimSpace(buyer.SpecifiedLegalOrganization.ID),
			CountryCode: buyer.PostalTradeAddress.CountryID,
		},
		Totals: Totals{
			LineTotalAmount:    lineTotal,
			TaxExclusiveAmount: taxExcl,
			TaxInclusiveAmount: taxIncl,
			PayableAmount:      payable,
			TotalTaxAmount:     taxTotal,
		},
	}

	return inv, nil
}
