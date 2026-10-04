package syntax

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

var (
	sirenRegex = regexp.MustCompile(`^[0-9]{9}$`)
	siretRegex = regexp.MustCompile(`^[0-9]{14}$`)
	vatFRRegex = regexp.MustCompile(`^FR[0-9A-Z]{2}[0-9]{9}$`)
)

type ValidationDiagnostic struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path"`
}

type Report struct {
	Valid       bool                   `json:"valid"`
	Diagnostics []ValidationDiagnostic `json:"diagnostics,omitempty"`
}

func ValidateEN16931Strict(inv *canonical.CanonicalInvoice) Report {
	var diags []ValidationDiagnostic

	if strings.TrimSpace(inv.InvoiceNumber) == "" {
		diags = append(diags, ValidationDiagnostic{
			RuleID:   "BR-01",
			Severity: "ERROR",
			Path:     "/Invoice/ID",
			Message:  "Le numero de facture est obligatoire",
		})
	}

	if inv.IssueDate.IsZero() {
		diags = append(diags, ValidationDiagnostic{
			RuleID:   "BR-02",
			Severity: "ERROR",
			Path:     "/Invoice/IssueDate",
			Message:  "La date d'emission est obligatoire",
		})
	}

	if !inv.DueDate.IsZero() && inv.DueDate.Before(inv.IssueDate) {
		diags = append(diags, ValidationDiagnostic{
			RuleID:   "BR-CO-04",
			Severity: "ERROR",
			Path:     "/Invoice/DueDate",
			Message:  "La date d'echeance ne peut pas etre anterieure a la date d'emission",
		})
	}

	if inv.Seller.CountryCode == "FR" {
		cleanedVat := strings.ReplaceAll(inv.Seller.TaxID, " ", "")
		if cleanedVat != "" && !vatFRRegex.MatchString(cleanedVat) {
			diags = append(diags, ValidationDiagnostic{
				RuleID:   "CIUS-FR-VAT",
				Severity: "ERROR",
				Path:     "/Invoice/Seller/TaxID",
				Message:  "Format de TVA FR invalide (attendu FR + 2 car. + 9 chiffres)",
			})
		}
		if inv.Seller.LegalEntityID != "" && !siretRegex.MatchString(inv.Seller.LegalEntityID) && !sirenRegex.MatchString(inv.Seller.LegalEntityID) {
			diags = append(diags, ValidationDiagnostic{
				RuleID:   "CIUS-FR-SIRET",
				Severity: "ERROR",
				Path:     "/Invoice/Seller/ID",
				Message:  "L'identifiant national francais doit etre un SIREN (9) ou SIRET (14)",
			})
		}
	}

	expectedTTC := inv.MonetaryTotals.NetHT + inv.MonetaryTotals.TaxAmount
	if inv.MonetaryTotals.GrossTTC != expectedTTC {
		diags = append(diags, ValidationDiagnostic{
			RuleID:   "BR-CO-15",
			Severity: "ERROR",
			Path:     "/Invoice/MonetaryTotals/GrossTTC",
			Message:  fmt.Sprintf("Incoherence montant total: NetHT (%d) + Tax (%d) != TTC (%d)", inv.MonetaryTotals.NetHT, inv.MonetaryTotals.TaxAmount, inv.MonetaryTotals.GrossTTC),
		})
	}

	var sumBaseHT, sumTax int64
	for _, tb := range inv.TaxBreakdowns {
		sumBaseHT += tb.BaseHT
		sumTax += tb.TaxAmount
	}

	if sumBaseHT != inv.MonetaryTotals.NetHT {
		diags = append(diags, ValidationDiagnostic{
			RuleID:   "BR-CO-13",
			Severity: "ERROR",
			Path:     "/Invoice/TaxBreakdowns",
			Message:  fmt.Sprintf("Somme des bases HT de TVA (%d) != NetHT facture (%d)", sumBaseHT, inv.MonetaryTotals.NetHT),
		})
	}

	if sumTax != inv.MonetaryTotals.TaxAmount {
		diags = append(diags, ValidationDiagnostic{
			RuleID:   "BR-CO-14",
			Severity: "ERROR",
			Path:     "/Invoice/TaxBreakdowns",
			Message:  fmt.Sprintf("Somme des montants TVA (%d) != Total TVA facture (%d)", sumTax, inv.MonetaryTotals.TaxAmount),
		})
	}

	return Report{
		Valid:       len(diags) == 0,
		Diagnostics: diags,
	}
}
