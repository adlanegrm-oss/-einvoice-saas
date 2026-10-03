package en16931

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// SVRLReport modélise la racine d'un rapport de validation Schematron (ISO SVRL)
type SVRLReport struct {
	XMLName         xml.Name         `xml:"schematron-output"`
	Title           string           `xml:"title,attr"`
	FiredRules      []SVRLFiredRule  `xml:"fired-rule"`
	FailedAsserts   []SVRLAssert     `xml:"failed-assert"`
	SuccessfulReports []SVRLReportItem `xml:"successful-report"`
}

type SVRLFiredRule struct {
	Context string `xml:"context,attr"`
	ID      string `xml:"id,attr"`
	Flag    string `xml:"flag,attr"`
}

type SVRLAssert struct {
	ID       string `xml:"id,attr"`
	Flag     string `xml:"flag,attr"`
	Location string `xml:"location,attr"`
	Test     string `xml:"test,attr"`
	Text     string `xml:"text"`
}

type SVRLReportItem struct {
	ID       string `xml:"id,attr"`
	Flag     string `xml:"flag,attr"`
	Location string `xml:"location,attr"`
	Test     string `xml:"test,attr"`
	Text     string `xml:"text"`
}

// DiagnosticItem représente une anomalie formatée pour l'API / PAF
type DiagnosticItem struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"` // "FATAL", "ERROR", "WARNING"
	Location string `json:"location"`
	Message  string `json:"message"`
}

// ValidationResult synthétise l'issue du contrôle sémantique
type ValidationResult struct {
	IsValid     bool             `json:"is_valid"`
	Diagnostics []DiagnosticItem `json:"diagnostics,omitempty"`
}

// ParseSVRL analyse un flux XML SVRL brut et génère le rapport typé
func ParseSVRL(rawSVRL []byte) (*ValidationResult, error) {
	if len(rawSVRL) == 0 {
		return nil, fmt.Errorf("rapport SVRL vide")
	}

	var report SVRLReport
	if err := xml.Unmarshal(rawSVRL, &report); err != nil {
		return nil, fmt.Errorf("échec de désérialisation du SVRL : %w", err)
	}

	result := &ValidationResult{
		IsValid:     true,
		Diagnostics: make([]DiagnosticItem, 0),
	}

	for _, assert := range report.FailedAsserts {
		severity := normalizeFlag(assert.Flag)
		if severity == "FATAL" || severity == "ERROR" {
			result.IsValid = false
		}

		result.Diagnostics = append(result.Diagnostics, DiagnosticItem{
			RuleID:   assert.ID,
			Severity: severity,
			Location: assert.Location,
			Message:  strings.TrimSpace(assert.Text),
		})
	}

	return result, nil
}

func normalizeFlag(flag string) string {
	switch strings.ToLower(flag) {
	case "fatal":
		return "FATAL"
	case "warning":
		return "WARNING"
	default:
		return "ERROR"
	}
}