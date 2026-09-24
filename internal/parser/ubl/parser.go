package ubl

import (
	"encoding/xml"
	"fmt"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain/models"
)

type Invoice struct {
	XMLName xml.Name `xml:"Invoice"`

	ID       string `xml:"ID"`
	IssueDate string `xml:"IssueDate"`
	Currency string `xml:"DocumentCurrencyCode"`

	AccountingSupplierParty PartyXML `xml:"AccountingSupplierParty"`
	AccountingCustomerParty PartyXML `xml:"AccountingCustomerParty"`
}

type PartyXML struct {
	PartyName PartyNameXML `xml:"Party>PartyName"`
}

type PartyNameXML struct {
	Name string `xml:"Name"`
}

func Parse(data []byte) (*models.Invoice, error) {
	var doc Invoice

	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid UBL XML: %w", err)
	}

	invoice := &models.Invoice{
		InvoiceNumber: doc.ID,
		Currency:      doc.Currency,
		Seller: models.Party{
			Name: doc.AccountingSupplierParty.PartyName.Name,
		},
		Buyer: models.Party{
			Name: doc.AccountingCustomerParty.PartyName.Name,
		},
	}

	return invoice, nil
}