package fr

import (
	"regexp"
	"strings"

	"einvoice-saas/internal/model"
)

var (
	// SIREN (9 chiffres), SIRET (14 chiffres) ou TVA intracommunautaire FR (FR + 11 caractères)
	sirenRegex = regexp.MustCompile(`^[0-9]{9}$`)
	siretRegex = regexp.MustCompile(`^[0-9]{14}$`)
	vatFRRegex = regexp.MustCompile(`^FR[0-9A-Z]{2}[0-9]{9}$`)

	// Taux légaux de TVA en France (Normal, Intermédiaire, Réduit, Super-réduit, Exonéré)
	validFRVATRates = map[float64]bool{
		0.0:  true,
		2.1:  true,
		5.5:  true,
		10.0: true,
		20.0: true,
	}
)

type FranceCanonicalValidator struct{}

func NewFranceCanonicalValidator() *FranceCanonicalValidator {
	return &FranceCanonicalValidator{}
}

func (v *FranceCanonicalValidator) JurisdictionCode() string {
	return "FR"
}

func (v *FranceCanonicalValidator) Validate(inv *model.CanonicalInvoice) model.ValidationReport {
	report := model.ValidationReport{
		Jurisdiction: "FR",
		Valid:        true,
		Issues:       []model.ValidationIssue{},
	}

	// 1. Contrôle SIREN / SIRET / TVA FR du Vendeur (CIUS-FR / Art. 242 nonies A)
	sellerTaxID := strings.TrimSpace(inv.Seller.TaxID)
	sellerNatID := strings.TrimSpace(inv.Seller.NationalID)
	hasValidTaxID := vatFRRegex.MatchString(sellerTaxID) || sirenRegex.MatchString(sellerTaxID) || siretRegex.MatchString(sellerTaxID)
	hasValidNatID := sirenRegex.MatchString(sellerNatID) || siretRegex.MatchString(sellerNatID)

	if !hasValidTaxID && !hasValidNatID {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      "FR-RULE-SELLER-ID-01",
			Description: "Identifiant fiscal ou légal du vendeur manquant ou invalide (SIREN, SIRET ou n° TVA FR obligatoire).",
			Severity:    model.SeverityError,
			Field:       "Seller.TaxID / Seller.NationalID",
			Remediation: "Renseigner un numéro SIREN (9 chiffres), SIRET (14 chiffres) ou TVA intracommunautaire FR valide.",
		})
	}

	// 2. Contrôle du numéro de facture et de la date
	if strings.TrimSpace(inv.InvoiceNumber) == "" {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      "FR-RULE-INVOICE-NUM-01",
			Description: "Le numéro de facture est obligatoire.",
			Severity:    model.SeverityError,
			Field:       "InvoiceNumber",
			Remediation: "Spécifier un numéro séquentiel unique de facture.",
		})
	}

	if inv.IssueDate.IsZero() {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      "FR-RULE-DATE-01",
			Description: "La date d'émission de la facture est obligatoire.",
			Severity:    model.SeverityError,
			Field:       "IssueDate",
			Remediation: "Indiquer la date d'émission au format ISO 8601 (AAAA-MM-JJ).",
		})
	}

	// 3. Contrôle des taux de TVA français
	for _, sub := range inv.TaxSubtotals {
		if !validFRVATRates[sub.Percent] {
			report.Valid = false
			report.Issues = append(report.Issues, model.ValidationIssue{
				RuleID:      "FR-RULE-VAT-RATE-01",
				Description: "Taux de TVA non conforme à la législation française.",
				Severity:    model.SeverityError,
				Field:       "TaxSubtotals.Percent",
				Remediation: "Appliquer un taux légal français (0%, 2.1%, 5.5%, 10%, 20%).",
			})
		}
	}

	return report
}