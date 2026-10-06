package validator_test

import (
"os"
"path/filepath"
"regexp"
"testing"

"einvoice-saas/internal/validator"
)

func TestGolden_EN16931_Schematron_Execution(t *testing.T) {
engine := validator.NewSchematronEngine(nil)

tests := []struct {
name          string
file          string
profile       validator.ValidationProfile
expectValid   bool
expectedRule  string
}{
{
name:        "FACT_2026_001_CONFORME_EN16931",
file:        "FACT_2026_001_CONFORME.xml",
profile:     validator.ProfileEN16931,
expectValid: true,
},
{
name:        "FACT_2026_001_CONFORME_CIUSFR",
file:        "FACT_2026_001_CONFORME.xml",
profile:     validator.ProfileCIUSFR,
expectValid: true,
},
{
name:         "FACT_2026_002_REJET_TVA_MANQUANTE",
file:         "FACT_2026_002_REJET_TVA_MANQUANTE.xml",
profile:      validator.ProfileEN16931,
expectValid:  false,
expectedRule: "BR-CO-09",
},
{
name:        "FACT_2026_003_CONFORME_EXO",
file:        "FACT_2026_003_CONFORME_EXO.xml",
profile:     validator.ProfileEN16931,
expectValid: true,
},
{
name:         "FACT_2026_004_REJET_SANS_CIUS",
file:         "FACT_2026_004_REJET_SANS_CIUS.xml",
profile:      validator.ProfileEN16931,
expectValid:  false,
expectedRule: "BR-01",
},
{
name:         "FACT_2026_005_REJET_CALCUL_TVA",
file:         "FACT_2026_005_REJET_CALCUL_TVA.xml",
profile:      validator.ProfileEN16931,
expectValid:  false,
expectedRule: "BR-CO-17",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
xmlPath := filepath.Join("..", "..", "factures_test_lots", tt.file)
xmlData, err := os.ReadFile(xmlPath)
if err != nil {
t.Skipf("fichier non trouvé (%s), skip", xmlPath)
}

report, err := engine.ValidateProfile(xmlData, tt.profile)
if err != nil {
t.Fatalf("erreur validation: %v", err)
}

if tt.expectValid {
if !report.Valid || len(report.Issues) > 0 {
t.Errorf("attendu: document valide sous %s, obtenu: rejet avec %d anomalies: %+v", tt.profile, len(report.Issues), report.Issues)
}
} else {
if report.Valid {
t.Errorf("attendu: rejet pour %s sous %s, obtenu: document valide", tt.expectedRule, tt.profile)
} else {
var found bool
for _, issue := range report.Issues {
if issue.RuleID == tt.expectedRule {
found = true
t.Logf("Succès rejet: [%s] (%s) %s", issue.RuleID, issue.Severity, issue.Message)
break
}
}
if !found {
t.Errorf("attendu: règle %s dans le rapport, obtenu: %+v", tt.expectedRule, report.Issues)
}
}
}
})
}

// Test spécifique CIUS-FR : Rejet SIRET manquant
t.Run("CIUS_FR_REJET_SIRET_VENDEUR_MANQUANT", func(t *testing.T) {
xmlPath := filepath.Join("..", "..", "factures_test_lots", "FACT_2026_001_CONFORME.xml")
xmlData, err := os.ReadFile(xmlPath)
if err != nil {
t.Skipf("fichier non trouvé (%s), skip", xmlPath)
}

// Suppression de la balise PartyIdentification du vendeur
re := regexp.MustCompile(`(?s)<cac:PartyIdentification>\s*<cbc:ID[^>]*>80245678900012</cbc:ID>\s*</cac:PartyIdentification>`)
corrupted := re.ReplaceAll(xmlData, []byte(""))

report, err := engine.ValidateProfile(corrupted, validator.ProfileCIUSFR)
if err != nil {
t.Fatalf("erreur validation: %v", err)
}

if report.Valid {
t.Errorf("attendu: rejet BR-FR-01 pour absence de SIRET sous CIUS-FR, obtenu: document valide")
} else {
var found bool
for _, issue := range report.Issues {
if issue.RuleID == "BR-FR-01" {
found = true
t.Logf("Succès rejet national: [%s] (%s) %s", issue.RuleID, issue.Severity, issue.Message)
break
}
}
if !found {
t.Errorf("attendu: règle BR-FR-01 dans le rapport, obtenu: %+v", report.Issues)
}
}
})
}
