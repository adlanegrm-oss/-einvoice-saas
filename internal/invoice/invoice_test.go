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
			name: "Facture valide",
			invoice: Invoice{
				ID:       "1",
				Number:   "INV-001",
				Customer: "Client A",
				Amount:   100.0,
				Currency: "EUR",
			},
			wantErr: false,
		},
		{
			name: "Numéro de facture manquant",
			invoice: Invoice{
				ID:       "2",
				Number:   "",
				Customer: "Client B",
				Amount:   100.0,
			},
			wantErr: true,
		},
		{
			name: "Nom de client manquant",
			invoice: Invoice{
				ID:       "3",
				Number:   "INV-003",
				Customer: "",
				Amount:   100.0,
			},
			wantErr: true,
		},
		{
			name: "Montant égal à zéro",
			invoice: Invoice{
				ID:       "4",
				Number:   "INV-004",
				Customer: "Client D",
				Amount:   0.0,
			},
			wantErr: true,
		},
		{
			name: "Montant négatif",
			invoice: Invoice{
				ID:       "5",
				Number:   "INV-005",
				Customer: "Client E",
				Amount:   -50.0,
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
