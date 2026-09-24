package invoice

import (
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
