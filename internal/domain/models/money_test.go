package models

import (
	"testing"
)

func TestMoney_Operations(t *testing.T) {
	t.Run("Addition sans dérive binaire de flottant", func(t *testing.T) {
		// 0.10 + 0.20 EUR = 0.30 EUR strict
		m1 := NewMoneyFromCents(10, CurrencyEUR)
		m2 := NewMoneyFromCents(20, CurrencyEUR)

		sum, err := m1.Add(m2)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if sum.Amount != 30 || sum.Scale != 2 {
			t.Errorf("attendu 30 centimes (Scale 2), obtenu: %d (Scale %d)", sum.Amount, sum.Scale)
		}
	})

	t.Run("Alignement des échelles (Prix unitaire à 4 décimales + Totaux)", func(t *testing.T) {
		// 1.0050 EUR (Scale 4) + 2.50 EUR (Scale 2 = 2.5000) = 3.5050 EUR (Scale 4)
		unit := NewMoney(10050, 4, CurrencyEUR)
		base := NewMoneyFromCents(250, CurrencyEUR)

		total, err := unit.Add(base)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if total.Amount != 35050 || total.Scale != 4 {
			t.Errorf("attendu 35050 (Scale 4), obtenu %d (Scale %d)", total.Amount, total.Scale)
		}

		// Arrondi final EN 16931 au centime le plus proche
		rounded := total.RoundTo2()
		if rounded.Amount != 351 || rounded.Scale != 2 {
			t.Errorf("attendu 351 centimes après arrondi (3.51 EUR), obtenu: %d", rounded.Amount)
		}
	})

	t.Run("Rejet addition devises différentes", func(t *testing.T) {
		eur := NewMoneyFromCents(100, CurrencyEUR)
		usd := NewMoneyFromCents(100, Currency("USD"))

		_, err := eur.Add(usd)
		if err == nil {
			t.Fatal("attendu une erreur d'incompatibilité de devises, obtenu nil")
		}
	})
}