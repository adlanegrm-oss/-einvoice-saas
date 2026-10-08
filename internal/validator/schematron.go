package validator

import (
	"fmt"

	"einvoice-saas/internal/validator/rules"
)

type ValidationProfile string

const (
	ProfileEN16931 ValidationProfile = "EN16931"
	ProfileCIUSFR  ValidationProfile = "CIUS-FR-2.0"
)

type Severity string

const (
	SeverityFatal   Severity = "FATAL"
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

type SchematronIssue struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	XPath    string   `json:"xpath"`
}

type SchematronReport struct {
	Profile ValidationProfile `json:"profile"`
	Valid   bool              `json:"valid"`
	Issues  []SchematronIssue `json:"issues"`
}

type SchematronEngine struct {
	xsltExecutor XSLTExecutor
}

func NewSchematronEngine(executor XSLTExecutor) *SchematronEngine {
	if executor == nil {
		executor = NewDefaultXSLTExecutor()
	}
	return &SchematronEngine{
		xsltExecutor: executor,
	}
}

// ValidateProfile rÃ©sout automatiquement l'artefact normatif selon le profil ciblÃ©
func (e *SchematronEngine) ValidateProfile(xmlData []byte, profile ValidationProfile) (*SchematronReport, error) {
	var xsltPath string
	switch profile {
	case ProfileEN16931:
		xsltPath = rules.PathEN16931XSLT
	case ProfileCIUSFR:
		xsltPath = rules.PathCIUSFRXSLT
	default:
		return nil, fmt.Errorf("schematron: profil non supportÃ©: %s", profile)
	}

	xsltBytes, err := rules.LoadRuleAsset(xsltPath)
	if err != nil {
		return nil, fmt.Errorf("schematron: Ã©chec chargement rÃ¨gle [%s]: %w", xsltPath, err)
	}

	return e.ValidateSchematron(xmlData, xsltBytes, profile)
}

func (e *SchematronEngine) ValidateSchematron(xmlData []byte, xsltData []byte, profile ValidationProfile) (*SchematronReport, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("schematron: document XML vide")
	}
	if len(xsltData) == 0 {
		return nil, fmt.Errorf("schematron: feuille de style XSLT vide")
	}

	svrlBytes, err := e.xsltExecutor.Transform(xmlData, xsltData)
	if err != nil {
		return nil, fmt.Errorf("schematron: Ã©chec transformation XSLT: %w", err)
	}

	report, err := ParseSVRL(svrlBytes)
	if err != nil {
		return nil, fmt.Errorf("schematron: Ã©chec analyse du rapport SVRL: %w", err)
	}

	report.Profile = profile
	return report, nil
}

// NewNativeBackedSchematronEngine expose le moteur Go natif derrière l'API SchematronEngine.
// Utile pour les environnements sans xsltproc, tout en gardant l'API ValidateProfile.
func NewNativeBackedSchematronEngine() *SchematronEngine {
return NewSchematronEngine(&NativeEN16931Executor{})
}
