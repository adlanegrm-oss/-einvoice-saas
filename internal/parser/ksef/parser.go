package ksef

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

type FakturaKSeF struct {
	XMLName  xml.Name `xml:"Faktura"`
	Naglowek struct {
		KodWaluty         string `xml:"KodWaluty"`
		DataWytworzeniaFa string `xml:"DataWytworzeniaFa"`
	} `xml:"Naglowek"`
	Podmiot1 struct {
		DaneIdentyfikacyjne struct {
			NIP        string `xml:"NIP"`
			PelnaNazwa string `xml:"PelnaNazwa"`
		} `xml:"DaneIdentyfikacyjne"`
	} `xml:"Podmiot1"`
	Podmiot2 struct {
		DaneIdentyfikacyjne struct {
			NIP        string `xml:"NIP"`
			PelnaNazwa string `xml:"PelnaNazwa"`
		} `xml:"DaneIdentyfikacyjne"`
	} `xml:"Podmiot2"`
	Fa struct {
		P_1       string     `xml:"P_1"`
		P_2       string     `xml:"P_2"`
		P_13_1    float64    `xml:"P_13_1"`
		P_14_1    float64    `xml:"P_14_1"`
		P_15      float64    `xml:"P_15"`
		FaWiersze []FaWiersz `xml:"FaWiersz"`
	} `xml:"Fa"`
}

type FaWiersz struct {
	P_7  string  `xml:"P_7"`
	P_8B float64 `xml:"P_8B"`
	P_9A float64 `xml:"P_9A"`
	P_12 float64 `xml:"P_12"`
}

func Parse(xmlBytes []byte) (*invoice.Invoice, error) {
	var k FakturaKSeF
	if err := xml.Unmarshal(xmlBytes, &k); err != nil {
		return nil, fmt.Errorf("échec décodage KSeF FA: %w", err)
	}

	currency := invoice.Currency(k.Naglowek.KodWaluty)
	if currency == "" {
		currency = invoice.CurrencyEUR
	}

	issueDate, _ := time.Parse("2006-01-02", k.Fa.P_1)
	if issueDate.IsZero() {
		issueDate, _ = time.Parse(time.RFC3339, k.Naglowek.DataWytworzeniaFa)
	}

	items := make([]invoice.InvoiceItem, 0, len(k.Fa.FaWiersze))
	for _, line := range k.Fa.FaWiersze {
		items = append(items, invoice.InvoiceItem{
			Description: line.P_7,
			Quantity:    float64(line.P_8B),
			UnitPrice:   invoice.NewMoneyFromFloat(line.P_9A, 2, currency),
			VATRate:     invoice.NewMoneyFromFloat(line.P_12, 2, currency),
		})
	}

	return &invoice.Invoice{
		ID:        k.Fa.P_2,
		Number:    k.Fa.P_2,
		IssueDate: issueDate,
		Currency:  currency,
		Seller: invoice.Party{
			Name:  k.Podmiot1.DaneIdentyfikacyjne.PelnaNazwa,
			SIRET: k.Podmiot1.DaneIdentyfikacyjne.NIP,
		},
		Customer: invoice.Party{
			Name:  k.Podmiot2.DaneIdentyfikacyjne.PelnaNazwa,
			SIRET: k.Podmiot2.DaneIdentyfikacyjne.NIP,
		},
		Items:    items,
		TotalHT:  invoice.NewMoneyFromFloat(k.Fa.P_13_1, 2, currency),
		TotalVAT: invoice.NewMoneyFromFloat(k.Fa.P_14_1, 2, currency),
		TotalTTC: invoice.NewMoneyFromFloat(k.Fa.P_15, 2, currency),
	}, nil
}
