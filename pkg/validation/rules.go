package validation

import (
	"fmt"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/money"
	"github.com/adlanegrm-oss/einvoice-saas/pkg/tax"
)

const (
	RulesetVersion = "EN16931-1:2017/A1:2019-core-subset-v0.5"
	CoverageScope  = "INTERNAL_CORE_RULES_WITHOUT_SCHEMATRON"
)

type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "ERROR"
	SeverityWarning DiagnosticSeverity = "WARNING"
)

type Diagnostic struct {
	RuleID      string             `json:"rule_id"`
	Description string             `json:"description"`
	Severity    DiagnosticSeverity `json:"severity"`
	TargetField string             `json:"target_field"`
}

type ValidationReport struct {
	RulesetVersion string       `json:"ruleset_version"`
	CoverageScope  string       `json:"coverage_scope"`
	RulesChecked   int          `json:"rules_checked"`
	Valid          bool         `json:"valid"`
	Diagnostics    []Diagnostic `json:"diagnostics"`
}

type InvoiceTotals struct {
	InvoiceNumber       string
	Currency            money.Currency
	SumInvoiceLines     money.Money    // BT-106
	AllowanceTotal      money.Money    // BT-107
	ChargeTotal         money.Money    // BT-108
	TaxExclusiveAmount  money.Money    // BT-109
	TaxInclusiveAmount  money.Money    // BT-110
	TaxTotalAmount      money.Money    // BT-111
	VATBreakdowns       []tax.Subtotal // BG-23
}

// ValidateStrictEN16931 checks compliance against official BR-xx rules without arbitrary tolerances.
func ValidateStrictEN16931(inv InvoiceTotals) ValidationReport {
	report := ValidationReport{
		RulesetVersion: RulesetVersion,
		CoverageScope:  CoverageScope,
		RulesChecked:   5,
		Valid:          true,
		Diagnostics:    make([]Diagnostic, 0),
	}

	addError := func(ruleID, desc, field string) {
		report.Valid = false
		report.Diagnostics = append(report.Diagnostics, Diagnostic{
			RuleID:      ruleID,
			Description: desc,
			Severity:    SeverityError,
			TargetField: field,
		})
	}

	// BR-02: An Invoice shall have an Invoice number (BT-1)
	if inv.InvoiceNumber == "" {
		addError("BR-02", "An Invoice shall have an Invoice number (BT-1).", "BT-1")
	}

	// BR-05: The currency code shall be valid ISO 4217
	if inv.Currency != money.EUR {
		addError("BR-05", fmt.Sprintf("Currency %s is not supported or invalid ISO 4217.", inv.Currency), "BT-5")
	}

	// BR-CO-13: BT-109 = BT-106 - BT-107 + BT-108 (Zero tolerance)
	expectedBT109 := inv.SumInvoiceLines.Cents() - inv.AllowanceTotal.Cents() + inv.ChargeTotal.Cents()
	if inv.TaxExclusiveAmount.Cents() != expectedBT109 {
		addError("BR-CO-13", fmt.Sprintf("Invoice total amount without VAT (BT-109: %d cents) must equal Sum of line amounts (BT-106: %d) - Allowances (BT-107: %d) + Charges (BT-108: %d). Expected: %d cents.",
			inv.TaxExclusiveAmount.Cents(), inv.SumInvoiceLines.Cents(), inv.AllowanceTotal.Cents(), inv.ChargeTotal.Cents(), expectedBT109), "BT-109")
	}

	// BR-CO-15: BT-110 = BT-109 + BT-111 (Zero tolerance - removed 0.05/0.02 tolerances)
	expectedBT110 := inv.TaxExclusiveAmount.Cents() + inv.TaxTotalAmount.Cents()
	if inv.TaxInclusiveAmount.Cents() != expectedBT110 {
		addError("BR-CO-15", fmt.Sprintf("Invoice total amount with VAT (BT-110: %d cents) must equal BT-109 (%d) + BT-111 (%d). Expected: %d cents.",
			inv.TaxInclusiveAmount.Cents(), inv.TaxExclusiveAmount.Cents(), inv.TaxTotalAmount.Cents(), expectedBT110), "BT-110")
	}

	// BR-CO-17: BT-111 = sum of all BT-117 in BG-23
	var sumVatBreakdown int64
	for _, b := range inv.VATBreakdowns {
		sumVatBreakdown += b.TaxAmount.Cents()
	}
	if inv.TaxTotalAmount.Cents() != sumVatBreakdown {
		addError("BR-CO-17", fmt.Sprintf("Invoice total VAT amount (BT-111: %d cents) must equal sum of VAT category amounts (%d cents).",
			inv.TaxTotalAmount.Cents(), sumVatBreakdown), "BT-111")
	}

	return report
}
