package validator_test

import (
"os"
"path/filepath"
"testing"

"einvoice-saas/internal/validator"
)

func TestGolden_EN16931_Schematron_Execution(t *testing.T) {
engine := validator.NewSchematronEngine(nil)

tests := []struct {
name          string
file          string
expectValid   bool
expectedRule  string
}{
{
name:        "FACT_2026_001_CONFORME",
file:        "FACT_2026_001_CONFORME.xml",
expectValid: true,
},
{
name:         "FACT_2026_002_REJET_TVA_MANQUANTE",
file:         "FACT_2026_002_REJET_TVA_MANQUANTE.xml",
expectValid:  false,
expectedRule: "BR-CO-09",
},
{
name:        "FACT_2026_003_CONFORME_EXO",
file:        "FACT_2026_003_CONFORME_EXO.xml",
expectValid: true,
},
{
name:         "FACT_2026_004_REJET_SANS_CIUS",
file:         "FACT_2026_004_REJET_SANS_CIUS.xml",
expectValid:  false,
expectedRule: "BR-01",
},
{
name:         "FACT_2026_005_REJET_CALCUL_TVA",
file:         "FACT_2026_005_REJET_CALCUL_TVA.xml",
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

report, err := engine.ValidateProfile(xmlData, validator.ProfileEN16931)
if err != nil {
t.Fatalf("erreur validation: %v", err)
}

if tt.expectValid {
if !report.Valid || len(report.Issues) > 0 {
t.Errorf("attendu: document valide, obtenu: rejet avec %d anomalies: %+v", len(report.Issues), report.Issues)
}
} else {
if report.Valid {
t.Errorf("attendu: rejet pour %s, obtenu: document valide", tt.expectedRule)
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
}
