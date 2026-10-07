package ma

import (
	"regexp"
	"strings"
)

type DiagnosticResult struct {
	Code         string `json:"code"`
	Severity     string `json:"severity"`
	Path         string `json:"path"`
	Message      string `json:"message"`
	ActualValue  string `json:"actual_value,omitempty"`
	ExpectedRule string `json:"expected_rule,omitempty"`
}

var (
	iceRegex = regexp.MustCompile(`^[0-9]{15}$`)
	ifRegex  = regexp.MustCompile(`^[0-9]{7,9}$`)
)

// AllowedVATRatesMaroc liste les taux légaux de TVA au Maroc (CGI art. 98/99)
var AllowedVATRatesMaroc = map[float64]bool{
	0.0:  true, // Exonéré / Export
	7.0:  true, // Eau, pharmacie
	10.0: true, // Restauration, hôtellerie, banques
	14.0: true, // Énergie électrique, transport
	20.0: true, // Taux normal
}

// ValidateICE contrôle l'Identifiant Commun de l'Entreprise (15 chiffres obligatoires)
func ValidateICE(ice string, path string) *DiagnosticResult {
	cleanICE := strings.TrimSpace(ice)
	if cleanICE == "" {
		return &DiagnosticResult{
			Code:         "MA-RULE-ICE-01",
			Severity:     "ERROR",
			Path:         path,
			Message:      "L'ICE est obligatoire pour les assujettis au Maroc.",
			ActualValue:  "",
			ExpectedRule: "15 chiffres numériques",
		}
	}
	if !iceRegex.MatchString(cleanICE) {
		return &DiagnosticResult{
			Code:         "MA-RULE-ICE-02",
			Severity:     "ERROR",
			Path:         path,
			Message:      "Format d'ICE non conforme.",
			ActualValue:  cleanICE,
			ExpectedRule: "Exactement 15 chiffres sans lettre ni séparateur",
		}
	}
	return nil
}

// ValidateIF contrôle l'Identifiant Fiscal délivré par la DGI
func ValidateIF(identifiantFiscal string, path string) *DiagnosticResult {
	cleanIF := strings.TrimSpace(identifiantFiscal)
	if cleanIF == "" {
		return &DiagnosticResult{
			Code:         "MA-RULE-IF-01",
			Severity:     "WARNING",
			Path:         path,
			Message:      "L'Identifiant Fiscal (IF) est manquant.",
			ActualValue:  "",
			ExpectedRule: "7 à 9 chiffres",
		}
	}
	if !ifRegex.MatchString(cleanIF) {
		return &DiagnosticResult{
			Code:         "MA-RULE-IF-02",
			Severity:     "WARNING",
			Path:         path,
			Message:      "Format de l'Identifiant Fiscal suspect.",
			ActualValue:  cleanIF,
			ExpectedRule: "Numéro IF valide délivré par la DGI",
		}
	}
	return nil
}

// ValidateVATRate vérifie la conformité des taux de taxe appliqués aux lignes
func ValidateVATRate(rate float64, lineID string) *DiagnosticResult {
	if !AllowedVATRatesMaroc[rate] {
		return &DiagnosticResult{
			Code:         "MA-RULE-VAT-01",
			Severity:     "ERROR",
			Path:         "InvoiceLine[" + lineID + "].TaxPercent",
			Message:      "Taux de TVA non conforme au CGI marocain.",
			ActualValue:  strings.TrimRight(strings.TrimRight(string(rune(int(rate))), "0"), "."),
			ExpectedRule: "0%, 7%, 10%, 14% ou 20%",
		}
	}
	return nil
}
