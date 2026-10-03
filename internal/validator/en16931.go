package validator

import (
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/domain"
)

type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
)

type Diagnostic struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Location string   `json:"location"`
	Message  string   `json:"message"`
	BTCode   string   `json:"bt_code"`
}

type ValidationReport struct {
	Valid       bool         `json:"valid"`
	Profile     string       `json:"profile"`
	Ruleset     string       `json:"ruleset"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type EN16931Engine struct {
	rulesetVersion string
}

func NewEN16931Engine(version string) *EN16931Engine {
	return &EN16931Engine{rulesetVersion: version}
}

func (e *EN16931Engine) ValidateFacture(
	invoiceNumber string,
	buyerSIRET string,
	sellerSIRET string,
	totals domain.MonetaryTotals,
	linesCount int,
) ValidationReport {
	report := ValidationReport{
		Valid:       true,
		Profile:     "EN16931-CIUS-FR",
		Ruleset:     e.rulesetVersion,
		Diagnostics: make([]Diagnostic, 0),
	}

	addError := func(rule, bt, location, msg string) {
		report.Valid = false
		report.Diagnostics = append(report.Diagnostics, Diagnostic{
			RuleID:   rule,
			Severity: SeverityError,
			Location: location,
			Message:  msg,
			BTCode:   bt,
		})
	}

	if strings.TrimSpace(invoiceNumber) == "" {
		addError("BR-01", "BT-1", "/Invoice/ID", "Une facture doit comporter un numéro d'identification unique.")
	}

	if totals.GrossTotal != totals.NetTotal.Add(totals.TaxTotal) {
		addError("BR-CO-09", "BT-112", "/Invoice/LegalMonetaryTotal/TaxInclusiveAmount",
			"Le total TTC (BT-112) doit être strictement égal à la somme du montant net HT (BT-109) et du total TVA (BT-110).")
	}

	if linesCount == 0 {
		addError("BR-16", "BG-25", "/Invoice/InvoiceLine", "Une facture doit comporter au moins une ligne de facturation.")
	}

	if len(strings.TrimSpace(sellerSIRET)) != 14 {
		addError("FR-R-01", "BT-29", "/Invoice/AccountingSupplierParty/Party/PartyLegalEntity/CompanyID",
			"Le vendeur doit être identifié par un numéro SIRET valide à 14 chiffres pour le profil français.")
	}

	return report
}
