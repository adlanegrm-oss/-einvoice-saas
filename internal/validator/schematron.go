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

// ValidateProfile résout automatiquement l'artefact normatif embarqué selon le profil ciblé
func (e *SchematronEngine) ValidateProfile(xmlData []byte, profile ValidationProfile) (*SchematronReport, error) {
var xsltPath string
switch profile {
case ProfileEN16931:
xsltPath = rules.PathEN16931XSLT
default:
return nil, fmt.Errorf("schematron: profil non supporté: %s", profile)
}

xsltBytes, err := rules.LoadRuleAsset(xsltPath)
if err != nil {
return nil, fmt.Errorf("schematron: échec chargement règle [%s]: %w", xsltPath, err)
}

return e.ValidateSchematron(xmlData, xsltBytes, profile)
}

// ValidateSchematron applique la feuille XSLT compilée sur le document XML et parse le rapport SVRL
func (e *SchematronEngine) ValidateSchematron(xmlData []byte, xsltData []byte, profile ValidationProfile) (*SchematronReport, error) {
if len(xmlData) == 0 {
return nil, fmt.Errorf("schematron: document XML vide")
}
if len(xsltData) == 0 {
return nil, fmt.Errorf("schematron: feuille de style XSLT vide")
}

svrlBytes, err := e.xsltExecutor.Transform(xmlData, xsltData)
if err != nil {
return nil, fmt.Errorf("schematron: échec transformation XSLT: %w", err)
}

report, err := ParseSVRL(svrlBytes)
if err != nil {
return nil, fmt.Errorf("schematron: échec analyse du rapport SVRL: %w", err)
}

report.Profile = profile
return report, nil
}
