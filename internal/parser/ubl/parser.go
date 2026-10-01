package ubl

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain/models"
)

type Invoice struct {
	XMLName xml.Name `xml:"Invoice"`

	ID        string `xml:"ID"`
	IssueDate string `xml:"IssueDate"`
	Currency  string `xml:"DocumentCurrencyCode"`

	AccountingSupplierParty PartyXML `xml:"AccountingSupplierParty"`
	AccountingCustomerParty PartyXML `xml:"AccountingCustomerParty"`
}

// PartyXML : Peppol BIS / EN 16931 place la raison sociale dans PartyLegalEntity/RegistrationName
// (obligatoire) ; PartyName/Name est facultatif. Les deux sont lus.
type PartyXML struct {
	PartyName   PartyNameXML   `xml:"Party>PartyName"`
	LegalEntity LegalEntityXML `xml:"Party>PartyLegalEntity"`
}

type PartyNameXML struct {
	Name string `xml:"Name"`
}

type LegalEntityXML struct {
	RegistrationName string `xml:"RegistrationName"`
}

// displayName renvoie le nom commercial, à défaut la raison sociale.
func (p PartyXML) displayName() string {
	if n := strings.TrimSpace(p.PartyName.Name); n != "" {
		return n
	}
	return strings.TrimSpace(p.LegalEntity.RegistrationName)
}

func Parse(data []byte) (*models.Invoice, error) {
	var doc Invoice

	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid UBL XML: %w", err)
	}

	invoice := &models.Invoice{
		InvoiceNumber: strings.TrimSpace(doc.ID),
		Currency:      strings.TrimSpace(doc.Currency),
		Seller:        models.Party{Name: doc.AccountingSupplierParty.displayName()},
		Buyer:         models.Party{Name: doc.AccountingCustomerParty.displayName()},
	}

	return invoice, nil
}
