package en16931

import (
"strings"
)

type Severity string

const (
SeverityError   Severity = "ERROR"
SeverityWarning Severity = "WARNING"
SeverityInfo    Severity = "INFO"
)

type RuleDiagnostic struct {
RuleID   string   `json:"rule_id"`
Severity Severity `json:"severity"`
Location string   `json:"location"`
Message  string   `json:"message"`
BTCode   string   `json:"bt_code"`
}

type ComplianceReport struct {
Valid       bool             `json:"valid"`
Ruleset     string           `json:"ruleset"`
Diagnostics []RuleDiagnostic `json:"diagnostics"`
}

type ComplianceEngine struct {
Version string
}

func NewComplianceEngine(version string) *ComplianceEngine {
return &ComplianceEngine{Version: version}
}

func (v *ComplianceEngine) Validate(invoiceNumber, sellerSIRET, buyerSIRET string, netCents, taxCents, grossCents int64, linesCount int) ComplianceReport {
report := ComplianceReport{
Valid:       true,
Ruleset:     v.Version,
Diagnostics: make([]RuleDiagnostic, 0),
}

addErr := func(rule, bt, loc, msg string) {
report.Valid = false
report.Diagnostics = append(report.Diagnostics, RuleDiagnostic{
RuleID: rule, Severity: SeverityError, Location: loc, Message: msg, BTCode: bt,
})
}

if strings.TrimSpace(invoiceNumber) == "" {
addErr("BR-01", "BT-1", "/Invoice/ID", "Numéro de facture obligatoire manquant.")
}

if grossCents != netCents+taxCents {
addErr("BR-CO-09", "BT-112", "/Invoice/LegalMonetaryTotal/TaxInclusiveAmount", "TTC != Net + TVA.")
}

if linesCount <= 0 {
addErr("BR-16", "BG-25", "/Invoice/InvoiceLine", "Au moins une ligne de facture requise.")
}

if len(strings.TrimSpace(sellerSIRET)) != 14 {
addErr("FR-R-01", "BT-29", "/Invoice/SupplierParty", "SIRET émetteur français invalide (14 chiffres requis).")
}

if len(strings.TrimSpace(buyerSIRET)) == 0 {
report.Diagnostics = append(report.Diagnostics, RuleDiagnostic{
RuleID: "FR-W-01", Severity: SeverityWarning, Location: "/Invoice/BuyerParty", Message: "Identifiant client recommandé pour routage.", BTCode: "BT-46",
})
}

return report
}
