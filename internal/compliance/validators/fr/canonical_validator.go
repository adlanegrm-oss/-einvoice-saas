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

// 1. Contrôle SIREN / SIRET / TVA FR du Vendeur (CIUS-FR / Art. 242 nonies A)
sellerTaxID := strings.TrimSpace(inv.Seller.TaxID)
sellerNatID := strings.TrimSpace(inv.Seller.NationalID)
hasValidSellerTaxID := vatFRRegex.MatchString(sellerTaxID) || sirenRegex.MatchString(sellerTaxID) || siretRegex.MatchString(sellerTaxID)
hasValidSellerNatID := sirenRegex.MatchString(sellerNatID) || siretRegex.MatchString(sellerNatID)

if !hasValidSellerTaxID && !hasValidSellerNatID {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      "FR-RULE-SELLER-ID-01",
Description: "Identifiant fiscal ou légal du vendeur manquant ou invalide (SIREN, SIRET ou n° TVA FR obligatoire).",
Severity:    model.SeverityError,
Field:       "Seller.TaxID / Seller.NationalID",
Remediation: "Renseigner un numéro SIREN (9 chiffres), SIRET (14 chiffres) ou TVA intracommunautaire FR valide.",
})
}

// 2. Contrôle SIREN / SIRET de l'Acheteur B2B FR (BR-FR-03)
if inv.Buyer.Country == "FR" || inv.Buyer.Country == "" {
buyerNatID := strings.TrimSpace(inv.Buyer.NationalID)
buyerTaxID := strings.TrimSpace(inv.Buyer.TaxID)
hasValidBuyerNatID := sirenRegex.MatchString(buyerNatID) || siretRegex.MatchString(buyerNatID)
hasValidBuyerTaxID := vatFRRegex.MatchString(buyerTaxID) || sirenRegex.MatchString(buyerTaxID) || siretRegex.MatchString(buyerTaxID)

if !hasValidBuyerNatID && !hasValidBuyerTaxID {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      "FR-RULE-BUYER-ID-01",
Description: "Identifiant national de l'acheteur français manquant ou invalide (SIREN ou SIRET obligatoire sous CIUS-FR).",
Severity:    model.SeverityError,
Field:       "Buyer.NationalID",
Remediation: "Renseigner le SIREN (9 chiffres) ou SIRET (14 chiffres) de l'entreprise cliente.",
})
}
}

// 3. Contrôle du numéro de facture et de la date
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

// 4. Contrôle des avoirs (Credit Note 381 -> référence obligatoire de la facture d'origine)
if strings.TrimSpace(inv.InvoiceTypeCode) == "381" {
if strings.TrimSpace(inv.PrecedingInvoiceReference) == "" {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      "FR-RULE-CREDIT-NOTE-REF-01",
Description: "Un avoir (code 381) doit obligatoirement référencer la facture rectifiée d'origine.",
Severity:    model.SeverityError,
Field:       "PrecedingInvoiceReference",
Remediation: "Indiquer le numéro de la facture initiale rectifiée dans le champ PrecedingInvoiceReference.",
})
}
}

// 5. Contrôle de la catégorie d'opération fiscale (Mandat FR 2026 / BR-FR-15)
opCat := strings.ToLower(strings.TrimSpace(inv.OperationCategory))
if opCat != "" && !validOperationCategories[opCat] {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      "FR-RULE-OPERATION-CAT-01",
Description: "La catégorie d'opération fiscale spécifiée est invalide.",
Severity:    model.SeverityError,
Field:       "OperationCategory",
Remediation: "Utiliser une valeur autorisée : 'goods' (biens), 'services' (services) ou 'mixed' (mixte).",
})
}

// 6. Contrôle des taux de TVA français et motifs d'exonération
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

// Exonération ou autoliquidation (Codes E, AE, K) -> Motif légal obligatoire (BT-120)
code := strings.ToUpper(strings.TrimSpace(sub.CategoryCode))
if (code == "E" || code == "AE" || code == "K") && strings.TrimSpace(sub.ExemptionReason) == "" {
report.Valid = false
report.Issues = append(report.Issues, model.ValidationIssue{
RuleID:      "FR-RULE-VAT-EXEMPT-REASON-01",
Description: "Pour une opération exonérée ou en autoliquidation (catégorie " + code + "), le motif légal d'exonération est obligatoire.",
Severity:    model.SeverityError,
Field:       "TaxSubtotals.ExemptionReason",
Remediation: "Fournir la mention légale justifiant l'exonération ou l'autoliquidation de la TVA.",
})
}
}

return report
}
