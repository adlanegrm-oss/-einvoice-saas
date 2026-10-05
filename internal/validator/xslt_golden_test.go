package validator_test

import (
"os"
"path/filepath"
"testing"

"einvoice-saas/internal/validator"
)

func TestGolden_EN16931_Schematron_Execution(t *testing.T) {
engine := validator.NewSchematronEngine(nil)

// 1. FACT_2026_001_CONFORME.xml -> Doit être VALIDE
t.Run("FACT_2026_001_CONFORME", func(t *testing.T) {
xmlPath := filepath.Join("..", "..", "factures_test_lots", "FACT_2026_001_CONFORME.xml")
xmlData, err := os.ReadFile(xmlPath)
if err != nil {
t.Skipf("fichier non trouvé (%s), skip", xmlPath)
}

report, err := engine.ValidateProfile(xmlData, validator.ProfileEN16931)
if err != nil {
t.Fatalf("erreur validation: %v", err)
}

if !report.Valid || len(report.Issues) > 0 {
t.Errorf("facture conforme rejetée avec %d anomalies: %+v", len(report.Issues), report.Issues)
}
})

// 2. FACT_2026_005_REJET_CALCUL_TVA.xml -> Doit être INVALIDE
t.Run("FACT_2026_005_REJET_CALCUL_TVA", func(t *testing.T) {
xmlPath := filepath.Join("..", "..", "factures_test_lots", "FACT_2026_005_REJET_CALCUL_TVA.xml")
xmlData, err := os.ReadFile(xmlPath)
if err != nil {
t.Skipf("fichier non trouvé (%s), skip", xmlPath)
}

report, err := engine.ValidateProfile(xmlData, validator.ProfileEN16931)
if err != nil {
t.Fatalf("erreur validation: %v", err)
}

if report.Valid {
t.Errorf("attendu: document invalide (rejet calcul TVA), obtenu: document valide")
} else {
t.Logf("Succès rejet: %d anomalie(s) détectée(s)", len(report.Issues))
for _, issue := range report.Issues {
t.Logf(" -> [%s] (%s) %s at %s", issue.RuleID, issue.Severity, issue.Message, issue.XPath)
}
}
})
}
