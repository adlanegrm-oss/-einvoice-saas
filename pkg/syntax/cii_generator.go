package syntax

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

type Amount struct {
	CurrencyID string `xml:"currencyID,attr,omitempty"`
	Value      string `xml:",chardata"`
}

type Quantity struct {
	UnitCode string `xml:"unitCode,attr"`
	Value    string `xml:",chardata"`
}

type ID struct {
	SchemeID string `xml:"schemeID,attr,omitempty"`
	Value    string `xml:",chardata"`
}

type DateString struct {
	Format string `xml:"format,attr,omitempty"`
	Value  string `xml:",chardata"`
}

type TaxScheme struct {
	ID string `xml:"ram:ID"`
}

type TradeTax struct {
	CalculatedAmount      *Amount `xml:"ram:CalculatedAmount,omitempty"`
	TypeCode              string  `xml:"ram:TypeCode"`
	ExemptionReason       string  `xml:"ram:ExemptionReason,omitempty"`
	BasisAmount           *Amount `xml:"ram:BasisAmount,omitempty"`
	CategoryCode          string  `xml:"ram:CategoryCode"`
	RateApplicablePercent string  `xml:"ram:RateApplicablePercent,omitempty"`
}

type TradeAddress struct {
	PostcodeCode string `xml:"ram:PostcodeCode,omitempty"`
	LineOne      string `xml:"ram:LineOne,omitempty"`
	CityName     string `xml:"ram:CityName,omitempty"`
	CountryID    string `xml:"ram:CountryID"`
}

type TradeParty struct {
	ID                *ID          `xml:"ram:ID,omitempty"`
	Name              string       `xml:"ram:Name"`
	SpecifiedLegalOrg *LegalOrg    `xml:"ram:SpecifiedLegalOrganization,omitempty"`
	PostalAddress     TradeAddress `xml:"ram:PostalTradeAddress"`
	URIUniversalComm  *URIComm     `xml:"ram:URIUniversalCommunication,omitempty"`
	SpecifiedTaxReg   *TaxReg      `xml:"ram:SpecifiedTaxRegistration,omitempty"`
}

type LegalOrg struct {
	ID *ID `xml:"ram:ID,omitempty"`
}

type URIComm struct {
	URIID *ID `xml:"ram:URIID,omitempty"`
}

type TaxReg struct {
	ID *ID `xml:"ram:ID,omitempty"`
}

type TradeLineItem struct {
	AssociatedDocLineDoc struct {
		LineID string `xml:"ram:LineID"`
	} `xml:"ram:AssociatedDocumentLineDocument"`
	SpecifiedTradeProduct struct {
		Name        string `xml:"ram:Name"`
		Description string `xml:"ram:Description,omitempty"`
	} `xml:"ram:SpecifiedTradeProduct"`
	SpecifiedLineTradeAgreement struct {
		NetPriceProductTradePrice struct {
			ChargeAmount Amount `xml:"ram:ChargeAmount"`
		} `xml:"ram:NetPriceProductTradePrice"`
	} `xml:"ram:SpecifiedLineTradeAgreement"`
	SpecifiedLineTradeDelivery struct {
		BilledQuantity Quantity `xml:"ram:BilledQuantity"`
	} `xml:"ram:SpecifiedLineTradeDelivery"`
	SpecifiedLineTradeSettlement struct {
		ApplicableTradeTax TradeTax `xml:"ram:ApplicableTradeTax"`
		MonetarySummation  struct {
			LineTotalAmount Amount `xml:"ram:LineTotalAmount"`
		} `xml:"ram:SpecifiedTradeSettlementLineMonetarySummation"`
	} `xml:"ram:SpecifiedLineTradeSettlement"`
}

type CrossIndustryInvoiceXML struct {
	XMLName  xml.Name `xml:"rsm:CrossIndustryInvoice"`
	XmlnsRSM string   `xml:"xmlns:rsm,attr"`
	XmlnsRAM string   `xml:"xmlns:ram,attr"`
	XmlnsQDT string   `xml:"xmlns:qdt,attr"`
	XmlnsUDT string   `xml:"xmlns:udt,attr"`

	Context struct {
		GuidelineParameter struct {
			ID string `xml:"ram:ID"`
		} `xml:"ram:GuidelineSpecifiedDocumentContextParameter"`
	} `xml:"rsm:ExchangedDocumentContext"`

	ExchangedDocument struct {
		ID            string `xml:"ram:ID"`
		TypeCode      string `xml:"ram:TypeCode"`
		IssueDateTime struct {
			DateTimeString DateString `xml:"udt:DateTimeString"`
		} `xml:"ram:IssueDateTime"`
	} `xml:"rsm:ExchangedDocument"`

	SupplyChainTradeTransaction struct {
		IncludedSupplyChainTradeLineItem []TradeLineItem `xml:"ram:IncludedSupplyChainTradeLineItem"`
		ApplicableHeaderTradeAgreement   struct {
			BuyerReference          string     `xml:"ram:BuyerReference,omitempty"`
			SellerTradeParty        TradeParty `xml:"ram:SellerTradeParty"`
			BuyerTradeParty         TradeParty `xml:"ram:BuyerTradeParty"`
			BuyerOrderReferencedDoc *struct {
				IssuerAssignedID string `xml:"ram:IssuerAssignedID"`
			} `xml:"ram:BuyerOrderReferencedDocument,omitempty"`
		} `xml:"ram:ApplicableHeaderTradeAgreement"`
		ApplicableHeaderTradeDelivery   struct{} `xml:"ram:ApplicableHeaderTradeDelivery"`
		ApplicableHeaderTradeSettlement struct {
			PaymentReference                     string `xml:"ram:PaymentReference,omitempty"`
			InvoiceCurrencyCode                  string `xml:"ram:InvoiceCurrencyCode"`
			SpecifiedTradeSettlementPaymentMeans *struct {
				TypeCode                           string `xml:"ram:TypeCode"`
				PayeePartyCreditorFinancialAccount *struct {
					IBANID string `xml:"ram:IBANID"`
				} `xml:"ram:PayeePartyCreditorFinancialAccount,omitempty"`
			} `xml:"ram:SpecifiedTradeSettlementPaymentMeans,omitempty"`
			ApplicableTradeTax         []TradeTax `xml:"ram:ApplicableTradeTax"`
			SpecifiedTradePaymentTerms *struct {
				Description     string `xml:"ram:Description,omitempty"`
				DueDateDateTime *struct {
					DateTimeString DateString `xml:"udt:DateTimeString"`
				} `xml:"ram:DueDateDateTime,omitempty"`
			} `xml:"ram:SpecifiedTradePaymentTerms,omitempty"`
			SpecifiedTradeSettlementHeaderMonetarySummation struct {
				LineTotalAmount     Amount  `xml:"ram:LineTotalAmount"`
				TaxBasisTotalAmount Amount  `xml:"ram:TaxBasisTotalAmount"`
				TaxTotalAmount      Amount  `xml:"ram:TaxTotalAmount"`
				GrandTotalAmount    Amount  `xml:"ram:GrandTotalAmount"`
				PrepaidAmount       *Amount `xml:"ram:TotalPrepaidAmount,omitempty"`
				DuePayableAmount    Amount  `xml:"ram:DuePayableAmount"`
			} `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
		} `xml:"ram:ApplicableHeaderTradeSettlement"`
	} `xml:"rsm:SupplyChainTradeTransaction"`
}

func GenerateCIIXML(inv *canonical.CanonicalInvoice) ([]byte, error) {
	if inv == nil {
		return nil, fmt.Errorf("canonical invoice cannot be nil")
	}

	curr := inv.Currency
	if curr == "" {
		curr = "EUR"
	}

	profileURN := inv.ProfileURN
	if profileURN == "" {
		profileURN = "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:en16931"
	}

	docType := inv.TypeCode
	if docType == "" {
		docType = "380"
	}

	var root CrossIndustryInvoiceXML
	root.XmlnsRSM = "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
	root.XmlnsRAM = "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
	root.XmlnsQDT = "urn:un:unece:uncefact:data:standard:QualifiedDataType:100"
	root.XmlnsUDT = "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100"

	root.Context.GuidelineParameter.ID = profileURN
	root.ExchangedDocument.ID = inv.InvoiceNumber
	root.ExchangedDocument.TypeCode = docType
	root.ExchangedDocument.IssueDateTime.DateTimeString = DateString{
		Format: "102",
		Value:  inv.IssueDate.Format("20060102"),
	}

	agreement := &root.SupplyChainTradeTransaction.ApplicableHeaderTradeAgreement
	agreement.BuyerReference = inv.BuyerReference
	if inv.PurchaseOrderRef != "" {
		agreement.BuyerOrderReferencedDoc = &struct {
			IssuerAssignedID string `xml:"ram:IssuerAssignedID"`
		}{IssuerAssignedID: inv.PurchaseOrderRef}
	}

	agreement.SellerTradeParty = TradeParty{
		Name: inv.Seller.Name,
		PostalAddress: TradeAddress{
			LineOne:      inv.Seller.AddressLine1,
			PostcodeCode: inv.Seller.PostalCode,
			CityName:     inv.Seller.CityName,
			CountryID:    inv.Seller.CountryCode,
		},
	}
	if inv.Seller.LegalEntityID != "" {
		agreement.SellerTradeParty.SpecifiedLegalOrg = &LegalOrg{
			ID: &ID{SchemeID: inv.Seller.EndpointScheme, Value: inv.Seller.LegalEntityID},
		}
	}
	if inv.Seller.TaxID != "" {
		agreement.SellerTradeParty.SpecifiedTaxReg = &TaxReg{
			ID: &ID{SchemeID: "VA", Value: strings.ReplaceAll(inv.Seller.TaxID, " ", "")},
		}
	}
	if inv.Seller.ElectronicAddr != "" {
		agreement.SellerTradeParty.URIUniversalComm = &URIComm{
			URIID: &ID{SchemeID: inv.Seller.EndpointScheme, Value: inv.Seller.ElectronicAddr},
		}
	}

	agreement.BuyerTradeParty = TradeParty{
		Name: inv.Buyer.Name,
		PostalAddress: TradeAddress{
			LineOne:      inv.Buyer.AddressLine1,
			PostcodeCode: inv.Buyer.PostalCode,
			CityName:     inv.Buyer.CityName,
			CountryID:    inv.Buyer.CountryCode,
		},
	}
	if inv.Buyer.LegalEntityID != "" {
		agreement.BuyerTradeParty.SpecifiedLegalOrg = &LegalOrg{
			ID: &ID{SchemeID: inv.Buyer.EndpointScheme, Value: inv.Buyer.LegalEntityID},
		}
	}
	if inv.Buyer.TaxID != "" {
		agreement.BuyerTradeParty.SpecifiedTaxReg = &TaxReg{
			ID: &ID{SchemeID: "VA", Value: strings.ReplaceAll(inv.Buyer.TaxID, " ", "")},
		}
	}
	if inv.Buyer.ElectronicAddr != "" {
		agreement.BuyerTradeParty.URIUniversalComm = &URIComm{
			URIID: &ID{SchemeID: inv.Buyer.EndpointScheme, Value: inv.Buyer.ElectronicAddr},
		}
	}

	for _, l := range inv.Lines {
		unit := l.UnitCode
		if unit == "" {
			unit = "C62"
		}
		item := TradeLineItem{}
		item.AssociatedDocLineDoc.LineID = l.ID
		item.SpecifiedTradeProduct.Name = l.Name
		item.SpecifiedTradeProduct.Description = l.Description

		item.SpecifiedLineTradeAgreement.NetPriceProductTradePrice.ChargeAmount = Amount{
			Value: fmt.Sprintf("%.2f", float64(l.UnitPriceNet)/100.0),
		}

		item.SpecifiedLineTradeDelivery.BilledQuantity = Quantity{
			UnitCode: unit,
			Value:    fmt.Sprintf("%.4f", l.Quantity),
		}

		item.SpecifiedLineTradeSettlement.ApplicableTradeTax = TradeTax{
			TypeCode:              "VAT",
			CategoryCode:          l.TaxCategory,
			RateApplicablePercent: fmt.Sprintf("%.2f", l.TaxRatePercent),
		}
		item.SpecifiedLineTradeSettlement.MonetarySummation.LineTotalAmount = Amount{
			Value: fmt.Sprintf("%.2f", float64(l.LineExtensionNet)/100.0),
		}
		root.SupplyChainTradeTransaction.IncludedSupplyChainTradeLineItem = append(
			root.SupplyChainTradeTransaction.IncludedSupplyChainTradeLineItem, item,
		)
	}

	settle := &root.SupplyChainTradeTransaction.ApplicableHeaderTradeSettlement
	settle.InvoiceCurrencyCode = curr
	settle.PaymentReference = inv.InvoiceNumber

	if inv.PaymentMeans.TypeCode != "" {
		pm := struct {
			TypeCode                           string `xml:"ram:TypeCode"`
			PayeePartyCreditorFinancialAccount *struct {
				IBANID string `xml:"ram:IBANID"`
			} `xml:"ram:PayeePartyCreditorFinancialAccount,omitempty"`
		}{
			TypeCode: inv.PaymentMeans.TypeCode,
		}
		if inv.PaymentMeans.IBAN != "" {
			pm.PayeePartyCreditorFinancialAccount = &struct {
				IBANID string `xml:"ram:IBANID"`
			}{IBANID: strings.ReplaceAll(inv.PaymentMeans.IBAN, " ", "")}
		}
		settle.SpecifiedTradeSettlementPaymentMeans = &pm
	}

	if inv.PaymentTerms != "" || !inv.DueDate.IsZero() {
		pt := struct {
			Description     string `xml:"ram:Description,omitempty"`
			DueDateDateTime *struct {
				DateTimeString DateString `xml:"udt:DateTimeString"`
			} `xml:"ram:DueDateDateTime,omitempty"`
		}{
			Description: inv.PaymentTerms,
		}
		if !inv.DueDate.IsZero() {
			pt.DueDateDateTime = &struct {
				DateTimeString DateString `xml:"udt:DateTimeString"`
			}{
				DateTimeString: DateString{
					Format: "102",
					Value:  inv.DueDate.Format("20060102"),
				},
			}
		}
		settle.SpecifiedTradePaymentTerms = &pt
	}

	for _, tb := range inv.TaxBreakdowns {
		calcAmt := Amount{Value: fmt.Sprintf("%.2f", float64(tb.TaxAmount)/100.0)}
		basisAmt := Amount{Value: fmt.Sprintf("%.2f", float64(tb.BaseHT)/100.0)}
		tax := TradeTax{
			CalculatedAmount:      &calcAmt,
			TypeCode:              "VAT",
			BasisAmount:           &basisAmt,
			CategoryCode:          tb.TaxCategory,
			RateApplicablePercent: fmt.Sprintf("%.2f", tb.TaxRatePercent),
			ExemptionReason:       tb.ExemptionReason,
		}
		settle.ApplicableTradeTax = append(settle.ApplicableTradeTax, tax)
	}

	lineTotal := inv.MonetaryTotals.LineExtensionTotal
	if lineTotal == 0 {
		lineTotal = inv.MonetaryTotals.NetHT
	}
	settle.SpecifiedTradeSettlementHeaderMonetarySummation.LineTotalAmount = Amount{
		CurrencyID: curr, Value: fmt.Sprintf("%.2f", float64(lineTotal)/100.0),
	}
	settle.SpecifiedTradeSettlementHeaderMonetarySummation.TaxBasisTotalAmount = Amount{
		CurrencyID: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.NetHT)/100.0),
	}
	settle.SpecifiedTradeSettlementHeaderMonetarySummation.TaxTotalAmount = Amount{
		CurrencyID: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.TaxAmount)/100.0),
	}
	settle.SpecifiedTradeSettlementHeaderMonetarySummation.GrandTotalAmount = Amount{
		CurrencyID: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.GrossTTC)/100.0),
	}
	if inv.MonetaryTotals.PrepaidAmount > 0 {
		settle.SpecifiedTradeSettlementHeaderMonetarySummation.PrepaidAmount = &Amount{
			CurrencyID: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.PrepaidAmount)/100.0),
		}
	}
	settle.SpecifiedTradeSettlementHeaderMonetarySummation.DuePayableAmount = Amount{
		CurrencyID: curr, Value: fmt.Sprintf("%.2f", float64(inv.MonetaryTotals.PayableDue)/100.0),
	}

	raw, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serialisation XML CII : %w", err)
	}

	return append([]byte(xml.Header), raw...), nil
}
