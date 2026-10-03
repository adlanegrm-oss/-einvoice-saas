package facturae

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

type FacturaeDocument struct {
	XMLName xml.Name `xml:"Facturae"`
	Parties struct {
		Seller struct {
			TaxIdentification struct {
				TaxIdentificationNumber string `xml:"TaxIdentificationNumber"`
			} `xml:"TaxIdentification"`
			LegalEntity struct {
				CorporateName  string `xml:"CorporateName"`
				AddressInSpain struct {
					Address     string `xml:"Address"`
					PostCode    string `xml:"PostCode"`
					Town        string `xml:"Town"`
					CountryCode string `xml:"CountryCode"`
				} `xml:"AddressInSpain"`
			} `xml:"LegalEntity"`
		} `xml:"SellerParty"`
		Buyer struct {
			TaxIdentification struct {
				TaxIdentificationNumber string `xml:"TaxIdentificationNumber"`
			} `xml:"TaxIdentification"`
			LegalEntity struct {
				CorporateName  string `xml:"CorporateName"`
				AddressInSpain struct {
					Address     string `xml:"Address"`
					PostCode    string `xml:"PostCode"`
					Town        string `xml:"Town"`
					CountryCode string `xml:"CountryCode"`
				} `xml:"AddressInSpain"`
			} `xml:"LegalEntity"`
		} `xml:"BuyerParty"`
	} `xml:"Parties"`
	Invoices struct {
		Invoice []struct {
			InvoiceHeader struct {
				InvoiceNumber string `xml:"InvoiceNumber"`
			} `xml:"InvoiceHeader"`
			InvoiceIssueData struct {
				IssueDate string `xml:"IssueDate"`
			} `xml:"InvoiceIssueData"`
			InvoiceTotals struct {
				TotalGrossAmountBeforeTaxes float64 `xml:"TotalGrossAmountBeforeTaxes"`
				TotalTaxOutputs             float64 `xml:"TotalTaxOutputs"`
				InvoiceTotal                float64 `xml:"InvoiceTotal"`
			} `xml:"InvoiceTotals"`
			Items struct {
				InvoiceLine []struct {
					ItemDescription     string  `xml:"ItemDescription"`
					Quantity            float64 `xml:"Quantity"`
					UnitPriceWithoutTax float64 `xml:"UnitPriceWithoutTax"`
					TaxesOutputs        struct {
						Tax []struct {
							TaxRate float64 `xml:"TaxRate"`
						} `xml:"Tax"`
					} `xml:"TaxesOutputs"`
				} `xml:"InvoiceLine"`
			} `xml:"Items"`
		} `xml:"Invoice"`
	} `xml:"Invoices"`
}

func Parse(xmlBytes []byte) (*invoice.Invoice, error) {
	var doc FacturaeDocument
	if err := xml.Unmarshal(xmlBytes, &doc); err != nil {
		return nil, fmt.Errorf("échec décodage Facturae: %w", err)
	}

	if len(doc.Invoices.Invoice) == 0 {
		return nil, fmt.Errorf("aucune facture trouvée dans Facturae")
	}

	rawInv := doc.Invoices.Invoice[0]
	parsedDate, _ := time.Parse("2006-01-02", rawInv.InvoiceIssueData.IssueDate)
	currency := invoice.CurrencyEUR

	items := make([]invoice.InvoiceItem, 0, len(rawInv.Items.InvoiceLine))
	for _, it := range rawInv.Items.InvoiceLine {
		var rate float64
		if len(it.TaxesOutputs.Tax) > 0 {
			rate = it.TaxesOutputs.Tax[0].TaxRate
		}
		items = append(items, invoice.InvoiceItem{
			Description: it.ItemDescription,
			Quantity:    float64(it.Quantity),
			UnitPrice:   invoice.NewMoneyFromFloat(it.UnitPriceWithoutTax, 2, currency),
			VATRate:     invoice.NewMoneyFromFloat(rate, 2, currency),
		})
	}

	seller := doc.Parties.Seller
	buyer := doc.Parties.Buyer

	return &invoice.Invoice{
		ID:        rawInv.InvoiceHeader.InvoiceNumber,
		Number:    rawInv.InvoiceHeader.InvoiceNumber,
		IssueDate: parsedDate,
		Currency:  currency,
		Seller: invoice.Party{
			Name:  seller.LegalEntity.CorporateName,
			SIRET: seller.TaxIdentification.TaxIdentificationNumber,
			Address: invoice.PostalAddress{
				StreetName:  seller.LegalEntity.AddressInSpain.Address,
				PostalZone:  seller.LegalEntity.AddressInSpain.PostCode,
				CityName:    seller.LegalEntity.AddressInSpain.Town,
				CountryCode: seller.LegalEntity.AddressInSpain.CountryCode,
			},
		},
		Customer: invoice.Party{
			Name:  buyer.LegalEntity.CorporateName,
			SIRET: buyer.TaxIdentification.TaxIdentificationNumber,
			Address: invoice.PostalAddress{
				StreetName:  buyer.LegalEntity.AddressInSpain.Address,
				PostalZone:  buyer.LegalEntity.AddressInSpain.PostCode,
				CityName:    buyer.LegalEntity.AddressInSpain.Town,
				CountryCode: buyer.LegalEntity.AddressInSpain.CountryCode,
			},
		},
		Items:    items,
		TotalHT:  invoice.NewMoneyFromFloat(rawInv.InvoiceTotals.TotalGrossAmountBeforeTaxes, 2, currency),
		TotalVAT: invoice.NewMoneyFromFloat(rawInv.InvoiceTotals.TotalTaxOutputs, 2, currency),
		TotalTTC: invoice.NewMoneyFromFloat(rawInv.InvoiceTotals.InvoiceTotal, 2, currency),
	}, nil
}
