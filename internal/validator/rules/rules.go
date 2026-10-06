package rules

import (
"embed"
"fmt"
)

// RulesFS embarque les artefacts normatifs européens et français
//
//go:embed en16931-cii-d16b/*
//go:embed cius-fr-v1.4/*
var RulesFS embed.FS

const (
// Chemins vers les artefacts EN 16931 v1.3.16
PathEN16931XSD  = "en16931-cii-d16b/v1.3.16/xsd/CrossIndustryInvoice_100pD16B.xsd"
PathEN16931XSLT = "en16931-cii-d16b/v1.3.16/xslt/EN16931-CII-validation.xslt"

// Chemins vers les artefacts CIUS-FR v1.4.0 (Mandat France B2B/B2G)
PathCIUSFRXSLT = "cius-fr-v1.4/CIUS-FR-CTC-validation.xslt"
)

// LoadRuleAsset lit un artefact embarqué par son chemin relatif
func LoadRuleAsset(path string) ([]byte, error) {
data, err := RulesFS.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("impossible de charger l'artefact normatif [%s]: %w", path, err)
}
return data, nil
}
