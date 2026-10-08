package fr

import (
	"regexp"
	"strings"

	"einvoice-saas/internal/model"
)

var (
	sirenRegex = regexp.MustCompile(`^[0-9]{9}$`)
	siretRegex = regexp.MustCompile(`^[0-9]{14}$`)
	vatFRRegex = regexp.MustCompile(`^FR[0-9A-Z]{2}[0-9]{9}$`)

	validFRVATRates = map[string]bool{
		"0":   true,
		"2.1": true,
		"5.5": true,
		"10":  true,
		"20":  true,
	}

	validOperationCategories = map[string]bool{
		"goods":    true,
		"services": true,
		"mixed":    true,
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

	sellerVAT := strings.TrimSpace(inv.Seller.VATID)
	sellerLegal := strings.TrimSpace(inv.Seller.LegalID)
	hasValidSellerVAT := vatFRRegex.MatchString(sellerVAT) || sirenRegex.MatchString(sellerVAT) || siretRegex.MatchString(sellerVAT)
	hasValidSellerLegal := sirenRegex.MatchString(sellerLegal) || siretRegex.MatchString(sellerLegal)

	if !hasValidSellerVAT && !hasValidSellerLegal {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      "FR-RULE-SELLER-ID-01",
			Description: "Identifiant fiscal ou légal du vendeur manquant ou invalide (SIREN, SIRET ou n° TVA FR obligatoire).",
			Severity:    model.SeverityError,
			Field:       "Seller.VATID / Seller.LegalID",
			Remediation: "Renseigner un numéro SIREN (9 chiffres), SIRET (14 chiffres) ou TVA intracommunautaire FR valide.",
		})
	}

	if inv.Buyer.CountryCode == "FR" || inv.Buyer.CountryCode == "" {
		buyerLegal := strings.TrimSpace(inv.Buyer.LegalID)
		buyerVAT := strings.TrimSpace(inv.Buyer.VATID)
		hasValidBuyerLegal := sirenRegex.MatchString(buyerLegal) || siretRegex.MatchString(buyerLegal)
		hasValidBuyerVAT := vatFRRegex.MatchString(buyerVAT) || sirenRegex.MatchString(buyerVAT) || siretRegex.MatchString(buyerVAT)

		if !hasValidBuyerLegal && !hasValidBuyerVAT {
			report.Valid = false
			report.Issues = append(report.Issues, model.ValidationIssue{
				RuleID:      "FR-RULE-BUYER-ID-01",
				Description: "Identifiant national de l'acheteur français manquant ou invalide (SIREN ou SIRET obligatoire sous CIUS-FR).",
				Severity:    model.SeverityError,
				Field:       "Buyer.LegalID",
				Remediation: "Renseigner le SIREN (9 chiffres) ou SIRET (14 chiffres) de l'entreprise cliente.",
			})
		}
	}

	if strings.TrimSpace(inv.ID) == "" {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      "FR-RULE-INVOICE-NUM-01",
			Description: "Le numéro de facture est obligatoire.",
			Severity:    model.SeverityError,
			Field:       "ID",
			Remediation: "Spécifier un numéro séquentiel unique de facture.",
		})
	}

	if strings.TrimSpace(inv.IssueDate) == "" {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      "FR-RULE-DATE-01",
			Description: "La date d'émission de la facture est obligatoire.",
			Severity:    model.SeverityError,
			Field:       "IssueDate",
			Remediation: "Indiquer la date d'émission au format ISO 8601 (AAAA-MM-JJ).",
		})
	}

	if strings.TrimSpace(inv.TypeCode) == "381" {
		hasPreceding := false
		for _, p := range inv.PrecedingInvoices {
			if strings.TrimSpace(p.ID) != "" {
				hasPreceding = true
				break
			}
		}
		if !hasPreceding {
			report.Valid = false
			report.Issues = append(report.Issues, model.ValidationIssue{
				RuleID:      "FR-RULE-CREDIT-NOTE-REF-01",
				Description: "Un avoir (code 381) doit obligatoirement référencer la facture rectifiée d'origine.",
				Severity:    model.SeverityError,
				Field:       "PrecedingInvoices",
				Remediation: "Indiquer le numéro de la facture initiale rectifiée dans PrecedingInvoices.",
			})
		}
	}

	opCat := strings.ToLower(strings.TrimSpace(inv.OperationCategory))
	if opCat != "" && !validOperationCategories[opCat] {
		report.Valid = false
		report.Issues = append(report.Issues, model.ValidationIssue{
			RuleID:      "FR-RULE-OPERATION-CAT-01",
			Description: "La catégorie d'opération fiscale spécifiée est invalide.",
			Severity:    model.SeverityError,
			Field:       "OperationCategory",
			Remediation: "Utiliser une valeur autorisée : 'goods', 'services' ou 'mixed'.",
		})
	}

	for _, sub := range inv.Taxes {
		pctKey := sub.Percent.StringFixed(1)
		switch pctKey {
		case "0.0":
			pctKey = "0"
		case "2.1":
			pctKey = "2.1"
		case "5.5":
			pctKey = "5.5"
		case "10.0":
			pctKey = "10"
		case "20.0":
			pctKey = "20"
		default:
			// garde la valeur pour faire échouer le check si non listée
			pctKey = strings.TrimRight(strings.TrimRight(sub.Percent.StringFixed(2), "0"), ".")
			if pctKey == "" {
				pctKey = "0"
			}
		}
		if !validFRVATRates[pctKey] {
			report.Valid = false
			report.Issues = append(report.Issues, model.ValidationIssue{
				RuleID:      "FR-RULE-VAT-RATE-01",
				Description: "Taux de TVA non conforme à la législation française.",
				Severity:    model.SeverityError,
				Field:       "Taxes.Percent",
				Remediation: "Appliquer un taux légal français (0%, 2.1%, 5.5%, 10%, 20%).",
			})
		}

		code := strings.ToUpper(strings.TrimSpace(sub.TaxCategoryCode))
		if (code == "E" || code == "AE" || code == "K") && strings.TrimSpace(sub.ExemptionReason) == "" {
			report.Valid = false
			report.Issues = append(report.Issues, model.ValidationIssue{
				RuleID:      "FR-RULE-VAT-EXEMPT-REASON-01",
				Description: "Pour une opération exonérée ou en autoliquidation (catégorie " + code + "), le motif légal d'exonération est obligatoire.",
				Severity:    model.SeverityError,
				Field:       "Taxes.ExemptionReason",
				Remediation: "Fournir la mention légale justifiant l'exonération ou l'autoliquidation de la TVA.",
			})
		}
	}

	return report
}
