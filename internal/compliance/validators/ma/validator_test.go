package ma

import (
	"testing"
)

func TestValidateICE(t *testing.T) {
	tests := []struct {
		name      string
		ice       string
		expectErr bool
	}{
		{"ICE Valide", "001524368000045", false},
		{"ICE Trop court", "001524368", true},
		{"ICE Alphanumerique", "00152436800004A", true},
		{"ICE Vide", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ValidateICE(tt.ice, "AccountingSupplierParty.PartyIdentification.ID")
			if (res != nil) != tt.expectErr {
				t.Fatalf("ValidateICE(%s) erreur attendue=%v, obtenu=%v", tt.ice, tt.expectErr, res)
			}
		})
	}
}

func TestValidateVATRate(t *testing.T) {
	validRates := []float64{0.0, 7.0, 10.0, 14.0, 20.0}
	for _, rate := range validRates {
		if issue := ValidateVATRate(rate, "1"); issue != nil {
			t.Errorf("Le taux légal %v%% ne doit pas générer d'erreur", rate)
		}
	}

	invalidRates := []float64{5.5, 8.5, 19.0, 21.0}
	for _, rate := range invalidRates {
		if issue := ValidateVATRate(rate, "1"); issue == nil {
			t.Errorf("Le taux %v%% aurait dû être rejeté", rate)
		}
	}
}
