package ma

import (
"einvoice-saas/internal/model"
)

type MoroccoCanonicalValidator struct{}

func NewMoroccoCanonicalValidator() *MoroccoCanonicalValidator {
return &MoroccoCanonicalValidator{}
}

func (v *MoroccoCanonicalValidator) JurisdictionCode() string {
return "MA"
}

func (v *MoroccoCanonicalValidator) Validate(inv *model.CanonicalInvoice) model.ValidationReport {
report := model.ValidationReport{
Jurisdiction: "MA",
Valid:        true,
Issues:       []model.ValidationIssue{},
}

ice := inv.Seller.LegalID
if ice == "" {
ice = inv.Seller.VATID
}
if diag := ValidateICE(ice, "Seller"); diag != nil {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      diag.Code,
Description: diag.Message,
Severity:    model.SeverityError,
Field:       diag.Path,
Remediation: diag.ExpectedRule,
})
}

for _, sub := range inv.Taxes {
pct, _ := sub.Percent.Float64()
if diag := ValidateVATRate(pct, "TaxSubtotal"); diag != nil {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      diag.Code,
Description: diag.Message,
Severity:    model.SeverityError,
Field:       diag.Path,
Remediation: diag.ExpectedRule,
})
}
}

return report
}
