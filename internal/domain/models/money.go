package models

import (
	"encoding/json"
	"fmt"
	"math"
)

// Currency identifie la devise (ISO 4217, ex: EUR).
type Currency string

const (
	CurrencyEUR Currency = "EUR"
)

// Money représente une valeur monétaire exacte sans virgule flottante.
// Amount est en sous-unités définies par Scale (ex: 12050 avec Scale=2 vaut 120.50 EUR).
type Money struct {
	Amount   int64    `json:"amount"`   // Sous-unités entières
	Scale    int32    `json:"scale"`    // Précision (ex: 2 pour totaux, 4 pour prix unitaires)
	Currency Currency `json:"currency"` // Code devise ISO 4217
}

// NewMoney instancie une valeur monétaire typée.
func NewMoney(amount int64, scale int32, cur Currency) Money {
	if cur == "" {
		cur = CurrencyEUR
	}
	return Money{
		Amount:   amount,
		Scale:    scale,
		Currency: cur,
	}
}

// NewMoneyFromCents crée un montant à 2 décimales.
func NewMoneyFromCents(cents int64, cur Currency) Money {
	return NewMoney(cents, 2, cur)
}

// NewMoneyFromFloat convertit un float64 vers Money en éliminant les imprécisions binaires.
func NewMoneyFromFloat(val float64, scale int32, cur Currency) Money {
	mult := math.Pow10(int(scale))
	cents := int64(math.Round(val * mult))
	return NewMoney(cents, scale, cur)
}

// ToFloat renvoie la valeur décimale approchée (pour sérialisation/affichage).
func (m Money) ToFloat() float64 {
	div := math.Pow10(int(m.Scale))
	return float64(m.Amount) / div
}

// Add additionne deux montants de même devise en alignant les échelles.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != "" && other.Currency != "" && m.Currency != other.Currency {
		return Money{}, fmt.Errorf("devises incompatibles: %s != %s", m.Currency, other.Currency)
	}

	maxScale := m.Scale
	if other.Scale > maxScale {
		maxScale = other.Scale
	}

	cur := m.Currency
	if cur == "" {
		cur = other.Currency
	}

	return Money{
		Amount:   m.rescale(maxScale) + other.rescale(maxScale),
		Scale:    maxScale,
		Currency: cur,
	}, nil
}

// Sub soustrait deux montants de même devise.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != "" && other.Currency != "" && m.Currency != other.Currency {
		return Money{}, fmt.Errorf("devises incompatibles: %s != %s", m.Currency, other.Currency)
	}

	maxScale := m.Scale
	if other.Scale > maxScale {
		maxScale = other.Scale
	}

	cur := m.Currency
	if cur == "" {
		cur = other.Currency
	}

	return Money{
		Amount:   m.rescale(maxScale) - other.rescale(maxScale),
		Scale:    maxScale,
		Currency: cur,
	}, nil
}

// RoundTo2 applique un arrondi demi-supérieur conforme EN 16931 vers 2 décimales.
func (m Money) RoundTo2() Money {
	if m.Scale <= 2 {
		return m.ToScale(2)
	}

	diff := m.Scale - 2
	factor := int64(math.Pow10(int(diff)))
	half := factor / 2

	var rounded int64
	if m.Amount >= 0 {
		rounded = (m.Amount + half) / factor
	} else {
		rounded = (m.Amount - half) / factor
	}

	return Money{
		Amount:   rounded,
		Scale:    2,
		Currency: m.Currency,
	}
}

// ToScale convertit le montant à une échelle cible sans arrondir.
func (m Money) ToScale(targetScale int32) Money {
	return Money{
		Amount:   m.rescale(targetScale),
		Scale:    targetScale,
		Currency: m.Currency,
	}
}

func (m Money) rescale(targetScale int32) int64 {
	if m.Scale == targetScale {
		return m.Amount
	}
	if targetScale > m.Scale {
		return m.Amount * int64(math.Pow10(int(targetScale-m.Scale)))
	}
	return m.Amount / int64(math.Pow10(int(m.Scale-targetScale)))
}

// UnmarshalJSON permet d'accepter soit un nombre direct (ex: 120.50), soit l'objet Money structuré.
func (m *Money) UnmarshalJSON(data []byte) error {
	var val float64
	if err := json.Unmarshal(data, &val); err == nil {
		*m = NewMoneyFromFloat(val, 2, CurrencyEUR)
		return nil
	}

	type alias Money
	var obj alias
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	*m = Money(obj)
	return nil
}

// MarshalJSON sérialise le montant sous forme décimale standardisée.
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.ToFloat())
}