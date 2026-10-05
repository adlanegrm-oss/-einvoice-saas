package ma

import (
"os"
"testing"
)

func TestValidateMoroccoUBL_File(t *testing.T) {
data, err := os.ReadFile("../../../../factures_test_lots/ma/test_facture_maroc_ok.xml")
if err != nil {
t.Fatalf("Impossible de lire le fichier de test: %v", err)
}

report := ValidateMoroccoUBL(data)

if !report.Valid {
t.Fatalf("La facture de test valide a été déclarée invalide. Erreurs: %d, Détails: %+v", report.ErrorsCount, report.Results)
}

if report.ErrorsCount != 0 {
t.Errorf("Attendu 0 erreur, obtenu: %d", report.ErrorsCount)
}
}
