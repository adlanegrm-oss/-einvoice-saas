package invoice

import (
	"database/sql/driver"
	"fmt"
)

// Value implémente driver.Valuer pour stocker un float64 dans SQLite
func (m Money) Value() (driver.Value, error) {
	return m.ToFloat(), nil
}

// Scan implémente sql.Scanner pour lire float64, int64, []byte ou string depuis SQLite
func (m *Money) Scan(value any) error {
	if value == nil {
		*m = NewMoney(0, 2, EUR)
		return nil
	}
	switch v := value.(type) {
	case float64:
		*m = NewMoneyFromFloat(v, 2, EUR)
		return nil
	case int64:
		*m = NewMoney(v*100, 2, EUR)
		return nil
	case []byte:
		var f float64
		if _, err := fmt.Sscanf(string(v), "%f", &f); err != nil {
			return err
		}
		*m = NewMoneyFromFloat(f, 2, EUR)
		return nil
	case string:
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
			return err
		}
		*m = NewMoneyFromFloat(f, 2, EUR)
		return nil
	default:
		return fmt.Errorf("type incompatible pour Money: %T", value)
	}
}
