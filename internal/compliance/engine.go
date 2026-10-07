package compliance

import (
"fmt"
"strings"

"einvoice-saas/internal/compliance/validators/fr"
"einvoice-saas/internal/compliance/validators/ma"
"einvoice-saas/internal/model"
"einvoice-saas/pkg/compliance/en16931"
)

// Dispatcher gère la validation transverse EN 16931 et le routage juridictionnel
type Dispatcher struct {
enValidator *en16931.Validator
validators  map[string]model.JurisdictionValidator
}

// NewDispatcher initialise le répartiteur avec le socle EN 16931 et les profils nationaux
func NewDispatcher() *Dispatcher {
d := &Dispatcher{
enValidator: en16931.NewValidator(),
validators:  make(map[string]model.JurisdictionValidator),
}
d.Register(ma.NewMoroccoCanonicalValidator())
d.Register(fr.NewFranceCanonicalValidator())
return d
}

// Register ajoute ou remplace un validateur de juridiction
func (d *Dispatcher) Register(v model.JurisdictionValidator) {
d.validators[strings.ToUpper(v.JurisdictionCode())] = v
}

// Validate exécute d'abord les règles socles EN 16931 puis les règles nationales
func (d *Dispatcher) Validate(inv *model.CanonicalInvoice) (model.ValidationReport, error) {
if inv == nil {
return model.ValidationReport{}, fmt.Errorf("facture canonique nulle")
}

jurisdiction := strings.ToUpper(strings.TrimSpace(inv.TargetJurisdiction))
if jurisdiction == "" {
jurisdiction = "FR"
}

report := model.ValidationReport{
Jurisdiction: jurisdiction,
Valid:        true,
Issues:       make([]model.ValidationIssue, 0),
}

// 1. Validation transverse EN 16931
enViolations := d.enValidator.Validate(inv)
for _, v := range enViolations {
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      v.RuleID,
Description: v.Message,
Severity:    v.Severity,
Field:       v.Path,
})
if v.Severity == model.SeverityError {
report.Valid = false
}
}

// 2. Validation spécifique de juridiction
validator, exists := d.validators[jurisdiction]
if !exists {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      "SYS-JURISDICTION-UNSUPPORTED",
Description: fmt.Sprintf("Aucun validateur actif configuré pour la juridiction: %s", jurisdiction),
Severity:    model.SeverityError,
Field:       "TargetJurisdiction",
Remediation: "Vérifier le code pays cible ou activer le CountryProfile correspondant.",
})
return report, nil
}

natReport := validator.Validate(inv)
if !natReport.Valid {
report.Valid = false
}
report.Issues = append(report.Issues, natReport.Issues...)

return report, nil
}
