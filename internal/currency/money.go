package currency

import (
	"database/sql/driver"
	"fmt"
	"github.com/shopspring/decimal"
)

// Amount gère les montants monétaires avec précision arbitraire (EN 16931 compliant)
type Amount struct {
	decimal.Decimal
}

func NewAmount(val string) (Amount, error) {
	d, err := decimal.NewFromString(val)
	if err != nil {
		return Amount{}, fmt.Errorf("montant invalide: %w", err)
	}
	return Amount{Decimal: d}, nil
}

func FromFloat(val float64) Amount {
	return Amount{Decimal: decimal.NewFromFloat(val)}
}

// RoundTo2 applique l'arrondi bancaire officiel (Half-Up ou Half-Even standard) à 2 décimales
func (a Amount) RoundTo2() Amount {
	return Amount{Decimal: a.Decimal.RoundBank(2)}
}

// RoundTo4 pour les prix unitaires détaillés (BT-146)
func (a Amount) RoundTo4() Amount {
	return Amount{Decimal: a.Decimal.RoundBank(4)}
}

func (a Amount) Add(other Amount) Amount {
	return Amount{Decimal: a.Decimal.Add(other.Decimal)}
}

func (a Amount) Sub(other Amount) Amount {
	return Amount{Decimal: a.Decimal.Sub(other.Decimal)}
}

func (a Amount) Mul(other Amount) Amount {
	return Amount{Decimal: a.Decimal.Mul(other.Decimal)}
}

// Support driver SQL
func (a Amount) Value() (driver.Value, error) {
	return a.StringFixed(2), nil
}

func (a *Amount) Scan(value interface{}) error {
	if value == nil {
		a.Decimal = decimal.Zero
		return nil
	}
	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("type incompatible pour Amount: %T", value)
	}
	d, err := decimal.NewFromString(str)
	if err != nil {
		return err
	}
	a.Decimal = d
	return nil
}
