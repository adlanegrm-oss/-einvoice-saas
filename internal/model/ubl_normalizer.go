package model

import (
	"encoding/xml"
	"strings"
)

type rawParty struct {
	PartyIdentification struct {
		ID struct {
			Value    string `xml:",chardata"`
			SchemeID string `xml:"schemeID,attr"`
		} `xml:"ID"`
	} `xml:"PartyIdentification"`
	PartyTaxScheme struct {
		CompanyID string `xml:"CompanyID"`
	} `xml:"PartyTaxScheme"`
	PartyLegalEntity struct {
		RegistrationName string `xml:"RegistrationName"`
	} `xml:"PartyLegalEntity"`
	PostalAddress struct {
		StreetName string `xml:"StreetName"`
		CityName   string `xml:"CityName"`
		PostalZone string `xml:"PostalZone"`
		Country    struct {
			IdentificationCode string `xml:"IdentificationCode"`
		} `xml:"Country"`
	} `xml:"PostalAddress"`
}

type rawTaxSubtotal struct {
	TaxableAmount struct {
		Value string `xml:",chardata"`
	} `xml:"TaxableAmount"`
	TaxAmount struct {
		Value string `xml:",chardata"`
	} `xml:"TaxAmount"`
	TaxCategory struct {
		Percent string `xml:"Percent"`
		ID      string `xml:"ID"`
	} `xml:"TaxCategory"`
}

type rawMonetaryTotal struct {
	LineExtensionAmount   string `xml:"LineExtensionAmount"`
	TaxExclusiveAmount    string `xml:"TaxExclusiveAmount"`
	TaxInclusiveAmount    string `xml:"TaxInclusiveAmount"`
	PayableAmount         string `xml:"PayableAmount"`
	AllowanceTotalAmount  string `xml:"AllowanceTotalAmount"`
	ChargeTotalAmount     string `xml:"ChargeTotalAmount"`
	PrepaidAmount         string `xml:"PrepaidAmount"`
	PayableRoundingAmount string `xml:"PayableRoundingAmount"`
}

type rawInvoiceLine struct {
	ID               string `xml:"ID"`
	InvoicedQuantity struct {
		Value    string `xml:",chardata"`
		UnitCode string `xml:"unitCode,attr"`
	} `xml:"InvoicedQuantity"`
	LineExtensionAmount string `xml:"LineExtensionAmount"`
	Item                struct {
		Name string `xml:"Name"`
	} `xml:"Item"`
	Price struct {
		PriceAmount string `xml:"PriceAmount"`
	} `xml:"Price"`
}

type rawUBLDocument struct {
	XMLName                 xml.Name `xml:"Invoice"`
	CustomizationID         string   `xml:"CustomizationID"`
	ProfileID               string   `xml:"ProfileID"`
	ID                      string   `xml:"ID"`
	InvoiceTypeCode         string   `xml:"InvoiceTypeCode"`
	IssueDate               string   `xml:"IssueDate"`
	DocumentCurrencyCode    string   `xml:"DocumentCurrencyCode"`
	TaxCurrencyCode         string   `xml:"TaxCurrencyCode"`
	BuyerReference          string   `xml:"BuyerReference"`
	AccountingSupplierParty struct {
		Party rawParty `xml:"Party"`
	} `xml:"AccountingSupplierParty"`
	AccountingCustomerParty struct {
		Party rawParty `xml:"Party"`
	} `xml:"AccountingCustomerParty"`
	TaxTotal struct {
		TaxSubtotal []rawTaxSubtotal `xml:"TaxSubtotal"`
	} `xml:"TaxTotal"`
	LegalMonetaryTotal rawMonetaryTotal `xml:"LegalMonetaryTotal"`
	InvoiceLines       []rawInvoiceLine `xml:"InvoiceLine"`
}

func mapParty(p rawParty) Party {
	return Party{
		Name:             p.PartyLegalEntity.RegistrationName,
		VATID:            strings.TrimSpace(p.PartyTaxScheme.CompanyID),
		LegalID:          strings.TrimSpace(p.PartyIdentification.ID.Value),
		LegalIDScheme:    p.PartyIdentification.ID.SchemeID,
		RegistrationName: p.PartyLegalEntity.RegistrationName,
		CountryCode:      p.PostalAddress.Country.IdentificationCode,
		City:             p.PostalAddress.CityName,
		PostalCode:       p.PostalAddress.PostalZone,
		Street:           p.PostalAddress.StreetName,
	}
}

// NormalizeUBLToCanonical convertit un flux UBL en CanonicalInvoice
func NormalizeUBLToCanonical(xmlPayload []byte) (*CanonicalInvoice, error) {
	var doc rawUBLDocument
	if err := xml.Unmarshal(xmlPayload, &doc); err != nil {
		return nil, err
	}

	taxCurrency := strings.TrimSpace(doc.TaxCurrencyCode)
	if taxCurrency == "" {
		taxCurrency = strings.TrimSpace(doc.DocumentCurrencyCode)
	}

	inv := &CanonicalInvoice{
		ID:               strings.TrimSpace(doc.ID),
		CustomizationID:  strings.TrimSpace(doc.CustomizationID),
		ProfileID:        strings.TrimSpace(doc.ProfileID),
		IssueDate:        strings.TrimSpace(doc.IssueDate),
		TypeCode:         strings.TrimSpace(doc.InvoiceTypeCode),
		DocumentCurrency: strings.TrimSpace(doc.DocumentCurrencyCode),
		TaxCurrency:      taxCurrency,
		BuyerReference:   strings.TrimSpace(doc.BuyerReference),
		Seller:           mapParty(doc.AccountingSupplierParty.Party),
		Buyer:            mapParty(doc.AccountingCustomerParty.Party),
	}

	lineTotal, _ := ParseDecimal(doc.LegalMonetaryTotal.LineExtensionAmount)
	taxExcl, _ := ParseDecimal(doc.LegalMonetaryTotal.TaxExclusiveAmount)
	taxIncl, _ := ParseDecimal(doc.LegalMonetaryTotal.TaxInclusiveAmount)
	payable, _ := ParseDecimal(doc.LegalMonetaryTotal.PayableAmount)
	allowance, _ := ParseDecimal(doc.LegalMonetaryTotal.AllowanceTotalAmount)
	charge, _ := ParseDecimal(doc.LegalMonetaryTotal.ChargeTotalAmount)
	prepaid, _ := ParseDecimal(doc.LegalMonetaryTotal.PrepaidAmount)
	rounding, _ := ParseDecimal(doc.LegalMonetaryTotal.PayableRoundingAmount)

	inv.Totals = Totals{
		LineTotalAmount:      lineTotal,
		AllowanceTotalAmount: allowance,
		ChargeTotalAmount:    charge,
		TaxExclusiveAmount:   taxExcl,
		TaxInclusiveAmount:   taxIncl,
		PrepaidAmount:        prepaid,
		RoundingAmount:       rounding,
		PayableAmount:        payable,
	}

	for _, rawLine := range doc.InvoiceLines {
		qty, _ := ParseDecimal(rawLine.InvoicedQuantity.Value)
		net, _ := ParseDecimal(rawLine.Price.PriceAmount)
		lt, _ := ParseDecimal(rawLine.LineExtensionAmount)
		inv.Lines = append(inv.Lines, Line{
			ID:              rawLine.ID,
			Quantity:        qty,
			UnitCode:        rawLine.InvoicedQuantity.UnitCode,
			LineTotalAmount: lt,
			NetPrice:        net,
			ItemName:        rawLine.Item.Name,
		})
	}

	for _, sub := range doc.TaxTotal.TaxSubtotal {
		taxable, _ := ParseDecimal(sub.TaxableAmount.Value)
		taxAmt, _ := ParseDecimal(sub.TaxAmount.Value)
		pct, _ := ParseDecimal(sub.TaxCategory.Percent)
		inv.Taxes = append(inv.Taxes, TaxSubtotal{
			TaxableAmount:   taxable,
			TaxAmount:       taxAmt,
			TaxCategoryCode: sub.TaxCategory.ID,
			Percent:         pct,
		})
	}

	return inv, nil
}
