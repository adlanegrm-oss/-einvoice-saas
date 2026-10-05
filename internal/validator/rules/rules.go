package rules

import (
"embed"
"fmt"
)

// RulesFS embarque les artefacts normatifs
//
//go:embed en16931-cii-d16b/*
var RulesFS embed.FS

const (
// Chemins vers les artefacts EN 16931 v1.3.16
PathEN16931XSD  = "en16931-cii-d16b/v1.3.16/xsd/CrossIndustryInvoice_100pD16B.xsd"
PathEN16931XSLT = "en16931-cii-d16b/v1.3.16/xslt/EN16931-CII-validation.xslt"
)

// LoadRuleAsset lit un artefact embarqué par son chemin relatif
func LoadRuleAsset(path string) ([]byte, error) {
data, err := RulesFS.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("impossible de charger l'artefact normatif [%s]: %w", path, err)
}
return data, nil
}
