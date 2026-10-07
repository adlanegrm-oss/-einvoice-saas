package validator

import (
	"github.com/shopspring/decimal"
)

type DiagnosticResult struct {
	Code         string `json:"code"`
	Severity     string `json:"severity"`
	Path         string `json:"path"`
	Message      string `json:"message"`
	ActualValue  string `json:"actual_value,omitempty"`
	ExpectedRule string `json:"expected_rule,omitempty"`
}

type ValidationReport struct {
	Valid         bool               `json:"valid"`
	Syntax        string             `json:"syntax"`
	Profile       string             `json:"profile"`
	ErrorsCount   int                `json:"errors_count"`
	WarningsCount int                `json:"warnings_count"`
	Results       []DiagnosticResult `json:"results"`
}

type InvoiceMonetaryTotals struct {
	LineExtensionAmount decimal.Decimal
	TaxExclusiveAmount  decimal.Decimal
	TaxTotalAmount      decimal.Decimal
	TaxInclusiveAmount  decimal.Decimal
	PrepaidAmount       decimal.Decimal
	RoundingAmount      decimal.Decimal
	PayableAmount       decimal.Decimal
	AllowanceTotal      decimal.Decimal
	ChargeTotal         decimal.Decimal
}

func ValidateEN16931CoreRules(t InvoiceMonetaryTotals) ValidationReport {
	report := ValidationReport{
		Valid:   true,
		Syntax:  "UBL-2.1",
		Profile: "urn:cen.eu:en16931:2017",
		Results: make([]DiagnosticResult, 0),
	}

	addError := func(code, path, msg, actual, expected string) {
		report.Valid = false
		report.ErrorsCount++
		report.Results = append(report.Results, DiagnosticResult{
			Code:         code,
			Severity:     "ERROR",
			Path:         path,
			Message:      msg,
			ActualValue:  actual,
			ExpectedRule: expected,
		})
	}

	// BR-CO-10: BT-115 = BT-112 - BT-113 + BT-114
	expectedPayable := t.TaxInclusiveAmount.Sub(t.PrepaidAmount).Add(t.RoundingAmount)
	if !t.PayableAmount.Equal(expectedPayable) {
		addError(
			"BR-CO-10",
			"/Invoice/LegalMonetaryTotal/PayableAmount",
			"PayableAmount (BT-115) must equal TaxInclusiveAmount (BT-112) - PrepaidAmount (BT-113) + RoundingAmount (BT-114).",
			t.PayableAmount.StringFixed(2),
			expectedPayable.StringFixed(2),
		)
	}

	// BR-CO-11: BT-112 = BT-109 + BT-110
	expectedInclusive := t.TaxExclusiveAmount.Add(t.TaxTotalAmount)
	if !t.TaxInclusiveAmount.Equal(expectedInclusive) {
		addError(
			"BR-CO-11",
			"/Invoice/LegalMonetaryTotal/TaxInclusiveAmount",
			"TaxInclusiveAmount (BT-112) must equal TaxExclusiveAmount (BT-109) + TaxTotalAmount (BT-110).",
			t.TaxInclusiveAmount.StringFixed(2),
			expectedInclusive.StringFixed(2),
		)
	}

	// BR-CO-13: BT-109 = BT-106 - BT-107 + BT-108
	expectedExclusive := t.LineExtensionAmount.Sub(t.AllowanceTotal).Add(t.ChargeTotal)
	if !t.TaxExclusiveAmount.Equal(expectedExclusive) {
		addError(
			"BR-CO-13",
			"/Invoice/LegalMonetaryTotal/TaxExclusiveAmount",
			"TaxExclusiveAmount (BT-109) must equal LineExtensionAmount (BT-106) - AllowanceTotal (BT-107) + ChargeTotal (BT-108).",
			t.TaxExclusiveAmount.StringFixed(2),
			expectedExclusive.StringFixed(2),
		)
	}

	// BR-CO-16: BT-115 >= 0.00
	if t.PayableAmount.IsNegative() {
		addError(
			"BR-CO-16",
			"/Invoice/LegalMonetaryTotal/PayableAmount",
			"PayableAmount (BT-115) cannot be negative.",
			t.PayableAmount.StringFixed(2),
			">= 0.00",
		)
	}

	return report
}
