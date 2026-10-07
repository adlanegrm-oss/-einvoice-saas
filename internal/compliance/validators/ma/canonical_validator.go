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

	// 1. Contrôle ICE Vendeur
	if diag := ValidateICE(inv.Seller.NationalID, "Seller"); diag != nil {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      diag.Code,
			Description: diag.Message,
			Severity:    model.SeverityError,
			Field:       diag.Path,
			Remediation: diag.ExpectedRule,
		})
	}

	// 2. Contrôle Taux TVA
	for _, sub := range inv.TaxSubtotals {
		if diag := ValidateVATRate(sub.Percent, "TaxSubtotal"); diag != nil {
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
