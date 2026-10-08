package model

import (
	"fmt"
	"github.com/shopspring/decimal"
)

type CanonicalInvoice struct {
	ID                string
	CustomizationID   string
	ProfileID         string
	IssueDate         string
	TypeCode          string
	DocumentCurrency  string
	TaxCurrency       string
	BuyerReference    string
	PrecedingInvoices []PrecedingInvoiceRef
	Seller            Party
	Buyer             Party
	Lines             []Line
	AllowancesCharges []AllowanceCharge
	Taxes             []TaxSubtotal
	Totals            Totals
	PaymentMeans      []PaymentMeans
	OperationCategory string
}

type PrecedingInvoiceRef struct {
	ID        string
	IssueDate string
}

type Party struct {
	Name              string
	TradingName       string
	VATID             string
	TaxRegistrationID string
	LegalID           string
	LegalIDScheme     string
	RegistrationName  string
	CountryCode       string
	City              string
	PostalCode        string
	Street            string
}

type Line struct {
	ID                  string
	Quantity            decimal.Decimal
	UnitCode            string
	LineTotalAmount     decimal.Decimal
	NetPrice            decimal.Decimal
	ItemName            string
	VATRatePercent      decimal.Decimal
	VATCategoryCode     string
	ExemptionReason     string
	ExemptionReasonCode string
	AllowancesCharges   []AllowanceCharge
}

type AllowanceCharge struct {
	IsCharge        bool
	BaseAmount      decimal.Decimal
	Percentage      decimal.Decimal
	Amount          decimal.Decimal
	VATCategoryCode string
	Reason          string
	ReasonCode      string
}

type TaxSubtotal struct {
	TaxableAmount       decimal.Decimal
	TaxAmount           decimal.Decimal
	TaxCategoryCode     string
	Percent             decimal.Decimal
	ExemptionReason     string
	ExemptionReasonCode string
}

type Totals struct {
	LineTotalAmount      decimal.Decimal
	AllowanceTotalAmount decimal.Decimal
	ChargeTotalAmount    decimal.Decimal
	TaxInclusiveAmount   decimal.Decimal
	TaxExclusiveAmount   decimal.Decimal
	PrepaidAmount        decimal.Decimal
	RoundingAmount       decimal.Decimal
	PayableAmount        decimal.Decimal
	TotalTaxAmount       decimal.Decimal
}

type PaymentMeans struct {
	TypeCode string
	IBAN     string
}

func ParseDecimal(val string) (decimal.Decimal, error) {
	if val == "" {
		return decimal.Zero, nil
	}
	d, err := decimal.NewFromString(val)
	if err != nil {
		return decimal.Zero, fmt.Errorf("montant décimal invalide '%s': %w", val, err)
	}
	return d.Round(2), nil
}

type IssueSeverity string

const (
	SeverityError   IssueSeverity = "ERROR"
	SeverityWarning IssueSeverity = "WARNING"
)

type ValidationIssue struct {
	RuleID      string        `json:"rule_id"`
	Description string        `json:"description"`
	Severity    IssueSeverity `json:"severity"`
	Field       string        `json:"field,omitempty"`
	Remediation string        `json:"remediation,omitempty"`
}

type ValidationReport struct {
	Jurisdiction string            `json:"jurisdiction"`
	Valid        bool              `json:"valid"`
	Issues       []ValidationIssue `json:"issues,omitempty"`
}

type JurisdictionValidator interface {
	JurisdictionCode() string
	Validate(invoice *CanonicalInvoice) ValidationReport
}

func MustParseDecimal(val string) decimal.Decimal {
	d, err := ParseDecimal(val)
	if err != nil {
		panic(err)
	}
	return d
}
