package invoice

import (
	"strings"
	"testing"
	"time"
)

func TestValidateInvoice(t *testing.T) {
	tests := []struct {
		name    string
		invoice Invoice
		wantErr bool
	}{
		{
			name: "Facture valide avec calcul de totaux",
			invoice: Invoice{
				ID:        "1",
				Number:    "INV-001",
				Customer:  Party{Name: "Client A"},
				IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
				Items: []InvoiceItem{
					{Description: "Développement Go", Quantity: 2, UnitPrice: NewMoneyFromFloat(100.0, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(20.0, 2, CurrencyEUR)},
				},
			},
			wantErr: false,
		},
		{
			name: "Numéro de facture manquant",
			invoice: Invoice{
				Customer:  Party{Name: "Client B"},
				IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
				Items: []InvoiceItem{
					{Description: "Service", Quantity: 1, UnitPrice: NewMoneyFromFloat(50.0, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(20.0, 2, CurrencyEUR)},
				},
			},
			wantErr: true,
		},
		{
			name: "Aucun article dans la facture",
			invoice: Invoice{
				Number:    "INV-002",
				Customer:  Party{Name: "Client C"},
				IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
				Items:     []InvoiceItem{},
			},
			wantErr: true,
		},
		{
			name: "Quantité invalide (<= 0)",
			invoice: Invoice{
				Number:    "INV-003",
				Customer:  Party{Name: "Client D"},
				IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
				Items: []InvoiceItem{
					{Description: "Service", Quantity: 0, UnitPrice: NewMoneyFromFloat(50.0, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(20.0, 2, CurrencyEUR)},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.invoice.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateTotals(t *testing.T) {
	inv := Invoice{
		Number:    "INV-TVA",
		Customer:  Party{Name: "Test TVA"},
		IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
		Items: []InvoiceItem{
			{Description: "Produit A", Quantity: 2, UnitPrice: NewMoneyFromFloat(100.0, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(20.0, 2, CurrencyEUR)}, // HT: 200, TVA: 40
			{Description: "Produit B", Quantity: 1, UnitPrice: NewMoneyFromFloat(100.0, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(10.0, 2, CurrencyEUR)}, // HT: 100, TVA: 10
		},
	}

	err := inv.Validate()
	if err != nil {
		t.Fatalf("Validation échouée : %v", err)
	}

	if inv.TotalHT.ToFloat() != 300.0 {
		t.Errorf("TotalHT incorrect : reçu %v, attendu 300.0", inv.TotalHT)
	}
	if inv.TotalVAT.ToFloat() != 50.0 {
		t.Errorf("TotalVAT incorrect : reçu %v, attendu 50.0", inv.TotalVAT)
	}
	if inv.TotalTTC.ToFloat() != 350.0 {
		t.Errorf("TotalTTC incorrect : reçu %v, attendu 350.0", inv.TotalTTC)
	}
}

func TestCalculateTotalsRounding(t *testing.T) {
	inv := Invoice{
		Number:    "INV-ROUND",
		Customer:  Party{Name: "Client"},
		IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC),
		Items: []InvoiceItem{
			{Description: "A", Quantity: 3, UnitPrice: NewMoneyFromFloat(0.1, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(20.0, 2, CurrencyEUR)},
			{Description: "B", Quantity: 1, UnitPrice: NewMoneyFromFloat(0.2, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(20.0, 2, CurrencyEUR)},
		},
	}
	if err := inv.Validate(); err != nil {
		t.Fatal(err)
	}
	if inv.TotalHT.ToFloat() != 0.5 || inv.TotalVAT.ToFloat() != 0.1 || inv.TotalTTC.ToFloat() != 0.6 {
		t.Errorf("totaux arrondis attendus 0.5 / 0.1 / 0.6, reçus %v / %v / %v", inv.TotalHT, inv.TotalVAT, inv.TotalTTC)
	}
}

func TestValidateRejectsAbnormalValues(t *testing.T) {
	base := func() Invoice {
		return Invoice{Number: "N", Customer: Party{Name: "C"},
			IssueDate: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC), Items: []InvoiceItem{{Description: "x", Quantity: 1, UnitPrice: NewMoneyFromFloat(10, 2, CurrencyEUR), VATRate: NewMoneyFromFloat(20, 2, CurrencyEUR)}}}
	}
	cases := map[string]func(*Invoice){
		"TVA négative":       func(i *Invoice) { i.Items[0].VATRate = NewMoney(-500, 2, CurrencyEUR) },
		"TVA supérieure 100": func(i *Invoice) { i.Items[0].VATRate = NewMoney(25000, 2, CurrencyEUR) },
		"prix négatif":       func(i *Invoice) { i.Items[0].UnitPrice = NewMoney(-100, 2, CurrencyEUR) },
		// float NaN non applicable sur int64 Money
		// float Inf non applicable sur int64 Money
		"numéro trop long": func(i *Invoice) { i.Number = strings.Repeat("9", 65) },
		"numéro d'espaces": func(i *Invoice) { i.Number = "   " },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			inv := base()
			mutate(&inv)
			if err := inv.Validate(); err == nil {
				t.Error("une erreur de validation était attendue")
			}
		})
	}
}
