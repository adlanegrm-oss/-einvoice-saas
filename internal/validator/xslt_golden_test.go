package validator_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"einvoice-saas/internal/validator"
)

func TestGolden_EN16931_Schematron_Execution(t *testing.T) {
	engine := validator.NewSchematronEngine(validator.NewDefaultXSLTExecutor())

	cases := []struct {
		name        string
		file        string
		profile     validator.ValidationProfile
		expectValid bool
		expectRule  string
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
			name:        "FACT_2026_002_REJET_TVA_MANQUANTE",
			file:        "FACT_2026_002_REJET_TVA_MANQUANTE.xml",
			profile:     validator.ProfileEN16931,
			expectValid: false,
			expectRule:  "BR-CO-09",
		},
		{
			name:        "FACT_2026_003_CONFORME_EXO",
			file:        "FACT_2026_003_CONFORME_EXO.xml",
			profile:     validator.ProfileEN16931,
			expectValid: true,
		},
		{
			name:        "FACT_2026_004_REJET_SANS_CIUS",
			file:        "FACT_2026_004_REJET_SANS_CIUS.xml",
			profile:     validator.ProfileEN16931,
			expectValid: false,
			expectRule:  "BR-01",
		},
		{
			name:        "FACT_2026_005_REJET_CALCUL_TVA",
			file:        "FACT_2026_005_REJET_CALCUL_TVA.xml",
			profile:     validator.ProfileEN16931,
			expectValid: false,
			expectRule:  "BR-CO-17",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			xmlPath := filepath.Join("..", "..", "factures_test_lots", tc.file)
			xmlData, err := os.ReadFile(xmlPath)
			if err != nil {
				t.Skipf("fichier non trouvé (%s), skip", xmlPath)
			}

			report, err := engine.ValidateProfile(xmlData, tc.profile)
			if err != nil {
				t.Fatalf("erreur validation: %v", err)
			}

			if tc.expectValid && !report.Valid {
				t.Fatalf("attendu valide mais erreurs: %+v", report.Issues)
			}
			if !tc.expectValid {
				if report.Valid {
					t.Fatalf("attendu rejet %s mais document déclaré valide", tc.expectRule)
				}
				var found bool
				for _, issue := range report.Issues {
					if issue.RuleID == tc.expectRule {
						found = true
						t.Logf("Succès rejet: [%s] (%s) %s", issue.RuleID, issue.Severity, issue.Message)
						break
					}
				}
				if !found {
					t.Fatalf("règle %s attendue, obtenu: %+v", tc.expectRule, report.Issues)
				}
			}
		})
	}

	t.Run("CIUS_FR_REJET_SIRET_VENDEUR_MANQUANT", func(t *testing.T) {
		xmlPath := filepath.Join("..", "..", "factures_test_lots", "FACT_2026_001_CONFORME.xml")
		xmlData, err := os.ReadFile(xmlPath)
		if err != nil {
			t.Skipf("fichier non trouvé (%s), skip", xmlPath)
		}

		re := regexp.MustCompile(`(?s)<cac:PartyIdentification>\s*<cbc:ID[^>]*>\d+</cbc:ID>\s*</cac:PartyIdentification>`)
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

	t.Run("CIUS_FR_REJET_SIRET_ACHETEUR_MANQUANT", func(t *testing.T) {
		xmlPath := filepath.Join("..", "..", "factures_test_lots", "FACT_2026_001_CONFORME.xml")
		xmlData, err := os.ReadFile(xmlPath)
		if err != nil {
			t.Skipf("fichier non trouvé (%s), skip", xmlPath)
		}

		reCustomer := regexp.MustCompile(`(?s)(<cac:AccountingCustomerParty>.*?)(<cac:PartyIdentification>\s*<cbc:ID[^>]*>\d+</cbc:ID>\s*</cac:PartyIdentification>)(.*?</cac:AccountingCustomerParty>)`)
		corrupted := reCustomer.ReplaceAll(xmlData, []byte("${1}${3}"))

		report, err := engine.ValidateProfile(corrupted, validator.ProfileCIUSFR)
		if err != nil {
			t.Fatalf("erreur validation: %v", err)
		}

		if report.Valid {
			t.Errorf("attendu: rejet BR-FR-03 pour absence de SIRET acheteur sous CIUS-FR, obtenu: document valide")
		} else {
			var found bool
			for _, issue := range report.Issues {
				if issue.RuleID == "BR-FR-03" {
					found = true
					t.Logf("Succès rejet national: [%s] (%s) %s", issue.RuleID, issue.Severity, issue.Message)
					break
				}
			}
			if !found {
				t.Errorf("attendu: règle BR-FR-03 dans le rapport, obtenu: %+v", report.Issues)
			}
		}
	})
}