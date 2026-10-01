package invoice

import (
	"math"
	"strings"
	"testing"
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
				ID:       "1",
				Number:   "INV-001",
				Customer: "Client A",
				Items: []InvoiceItem{
					{Description: "Développement Go", Quantity: 2, UnitPrice: 100.0, VATRate: 20.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Numéro de facture manquant",
			invoice: Invoice{
				Customer: "Client B",
				Items: []InvoiceItem{
					{Description: "Service", Quantity: 1, UnitPrice: 50.0, VATRate: 20.0},
				},
			},
			wantErr: true,
		},
		{
			name: "Aucun article dans la facture",
			invoice: Invoice{
				Number:   "INV-002",
				Customer: "Client C",
				Items:    []InvoiceItem{},
			},
			wantErr: true,
		},
		{
			name: "Quantité invalide (<= 0)",
			invoice: Invoice{
				Number:   "INV-003",
				Customer: "Client D",
				Items: []InvoiceItem{
					{Description: "Service", Quantity: 0, UnitPrice: 50.0, VATRate: 20.0},
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
		Number:   "INV-TVA",
		Customer: "Test TVA",
		Items: []InvoiceItem{
			{Description: "Produit A", Quantity: 2, UnitPrice: 100.0, VATRate: 20.0}, // HT: 200, TVA: 40
			{Description: "Produit B", Quantity: 1, UnitPrice: 100.0, VATRate: 10.0}, // HT: 100, TVA: 10
		},
	}

	err := inv.Validate()
	if err != nil {
		t.Fatalf("Validation échouée : %v", err)
	}

	if inv.TotalHT != 300.0 {
		t.Errorf("TotalHT incorrect : reçu %v, attendu 300.0", inv.TotalHT)
	}
	if inv.TotalVAT != 50.0 {
		t.Errorf("TotalVAT incorrect : reçu %v, attendu 50.0", inv.TotalVAT)
	}
	if inv.TotalTTC != 350.0 {
		t.Errorf("TotalTTC incorrect : reçu %v, attendu 350.0", inv.TotalTTC)
	}
}

func TestCalculateTotalsRounding(t *testing.T) {
	inv := Invoice{
		Number:   "INV-ROUND",
		Customer: "Client",
		Items: []InvoiceItem{
			{Description: "A", Quantity: 3, UnitPrice: 0.1, VATRate: 20.0},
			{Description: "B", Quantity: 1, UnitPrice: 0.2, VATRate: 20.0},
		},
	}
	if err := inv.Validate(); err != nil {
		t.Fatal(err)
	}
	if inv.TotalHT != 0.5 || inv.TotalVAT != 0.1 || inv.TotalTTC != 0.6 {
		t.Errorf("totaux arrondis attendus 0.5 / 0.1 / 0.6, reçus %v / %v / %v", inv.TotalHT, inv.TotalVAT, inv.TotalTTC)
	}
}

func TestValidateRejectsAbnormalValues(t *testing.T) {
	base := func() Invoice {
		return Invoice{Number: "N", Customer: "C", Items: []InvoiceItem{{Description: "x", Quantity: 1, UnitPrice: 10, VATRate: 20}}}
	}
	cases := map[string]func(*Invoice){
		"TVA négative":       func(i *Invoice) { i.Items[0].VATRate = -5 },
		"TVA supérieure 100": func(i *Invoice) { i.Items[0].VATRate = 250 },
		"prix négatif":       func(i *Invoice) { i.Items[0].UnitPrice = -1 },
		"prix NaN":           func(i *Invoice) { i.Items[0].UnitPrice = math.NaN() },
		"prix infini":        func(i *Invoice) { i.Items[0].UnitPrice = math.Inf(1) },
		"numéro trop long":   func(i *Invoice) { i.Number = strings.Repeat("9", 65) },
		"numéro d'espaces":   func(i *Invoice) { i.Number = "   " },
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
