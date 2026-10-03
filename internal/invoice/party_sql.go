package invoice

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Value implémente driver.Valuer pour la persistance SQL
func (p Party) Value() (driver.Value, error) {
	return p.Name, nil
}

// Scan implémente sql.Scanner pour la lecture SQL
func (p *Party) Scan(value any) error {
	if value == nil {
		p.Name = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		p.Name = v
		return nil
	case []byte:
		p.Name = string(v)
		return nil
	default:
		return fmt.Errorf("type incompatible pour Party: %T", value)
	}
}

// UnmarshalJSON permet d'accepter soit une chaîne ("Client"), soit un objet ({"name": "Client"})
func (p *Party) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		p.Name = s
		return nil
	}
	type alias Party
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*p = Party(a)
	return nil
}
