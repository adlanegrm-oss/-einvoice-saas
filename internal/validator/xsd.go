package validator

import (
"bytes"
"fmt"
"os"
)

type XSDValidator struct {
SchemaPath string
}

func NewXSDValidator(schemaPath string) *XSDValidator {
return &XSDValidator{
SchemaPath: schemaPath,
}
}

// Validate vérifie la présence du schéma et s'assure que le document XML n'est pas vide.
// Cette méthode sert d'interface de pré-validation avant de brancher le parseur/moteur XSD cible.
func (v *XSDValidator) Validate(xmlData []byte) error {
if len(bytes.TrimSpace(xmlData)) == 0 {
return fmt.Errorf("xsd: empty XML")
}

if _, err := os.Stat(v.SchemaPath); err != nil {
return fmt.Errorf("xsd: schema %q: %w", v.SchemaPath, err)
}

return nil
}
