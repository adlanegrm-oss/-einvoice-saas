package invoice

import (
	"errors"
	"math"
	"time"
)

const (
	EUR Currency = "EUR"
	USD Currency = "USD"
)

type PostalAddress struct {
	StreetName  string `json:"street_name,omitempty"`
	CityName    string `json:"city_name,omitempty"`
	PostalZone  string `json:"postal_zone,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	Address     string `json:"address,omitempty"`
}

type Party struct {
	Name    string        `json:"name"`
	SIRET   string        `json:"siret,omitempty"`
	VATID   string        `json:"vat_id,omitempty"`
	Address PostalAddress `json:"address,omitempty"`
}

type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   Money   `json:"unit_price"`
	VATRate     Money   `json:"vat_rate"`
	TotalHT     Money   `json:"total_ht,omitempty"`
}

type Invoice struct {
	ID          string        `json:"id"`
	Number      string        `json:"number"`
	Currency    Currency      `json:"currency"`
	IssueDate   time.Time     `json:"issue_date"`
	Seller      Party         `json:"seller"`
	Customer    Party         `json:"customer"`
	Items       []InvoiceItem `json:"items"`
	TotalHT     Money         `json:"total_ht"`
	TotalVAT    Money         `json:"total_vat"`
	TotalTTC    Money         `json:"total_ttc"`
	IsValidated bool          `json:"is_validated"`
}

func (i *Invoice) CalculateTotals() {
	var totalHT, totalVAT float64
	cur := i.Currency
	if cur == "" {
		cur = EUR
		i.Currency = EUR
	}
	for idx := range i.Items {
		itemHT := math.Round(i.Items[idx].Quantity*i.Items[idx].UnitPrice.ToFloat()*100) / 100
		itemVAT := math.Round(itemHT*(i.Items[idx].VATRate.ToFloat()/100.0)*100) / 100
		i.Items[idx].TotalHT = NewMoneyFromFloat(itemHT, 2, cur)
		totalHT += itemHT
		totalVAT += itemVAT
	}
	i.TotalHT = NewMoneyFromFloat(totalHT, 2, cur)
	i.TotalVAT = NewMoneyFromFloat(totalVAT, 2, cur)
	i.TotalTTC = NewMoneyFromFloat(totalHT+totalVAT, 2, cur)
}


// Validate vérifie l'intégrité minimale et les champs obligatoires de la facture.
func (i *Invoice) Validate() error {
	if i.Number == "" {
		return errors.New("le numéro de facture est obligatoire")
	}
	if i.IssueDate.IsZero() {
		return errors.New("la date d'émission est obligatoire")
	}
	if len(i.Items) == 0 {
		return errors.New("la facture doit contenir au moins une ligne")
	}
	if i.TotalHT.ToFloat() < 0 || i.TotalTTC.ToFloat() < 0 {
		return errors.New("les montants ne peuvent pas être négatifs")
	}
	return nil
}
