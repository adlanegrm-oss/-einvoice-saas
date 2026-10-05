package model

import "time"

type Party struct {
Name        string            `json:"name"`
TaxID       string            `json:"tax_id,omitempty"`
NationalID  string            `json:"national_id,omitempty"`
Country     string            `json:"country"`
AddressLine string            `json:"address_line,omitempty"`
City        string            `json:"city,omitempty"`
PostalZone  string            `json:"postal_zone,omitempty"`
Identifiers map[string]string `json:"identifiers,omitempty"`
}

type TaxSubtotal struct {
TaxableAmount float64 `json:"taxable_amount"`
TaxAmount     float64 `json:"tax_amount"`
Percent       float64 `json:"percent"`
CategoryCode  string  `json:"category_code,omitempty"`
}

type InvoiceLine struct {
ID          string  `json:"id"`
Description string  `json:"description"`
Quantity    float64 `json:"quantity"`
UnitPrice   float64 `json:"unit_price"`
LineTotal   float64 `json:"line_total"`
VatPercent  float64 `json:"vat_percent"`
VatCategory string  `json:"vat_category,omitempty"`
}

type MonetaryTotals struct {
LineExtensionAmount float64 `json:"line_extension_amount"`
TaxExclusiveAmount  float64 `json:"tax_exclusive_amount"`
TaxInclusiveAmount  float64 `json:"tax_inclusive_amount"`
AllowanceTotal      float64 `json:"allowance_total,omitempty"`
ChargeTotal         float64 `json:"charge_total,omitempty"`
PrepaidAmount       float64 `json:"prepaid_amount,omitempty"`
PayableAmount       float64 `json:"payable_amount"`
}

type CanonicalInvoice struct {
ID                 string         `json:"id"`
InvoiceNumber      string         `json:"invoice_number"`
IssueDate          time.Time      `json:"issue_date"`
DueDate            *time.Time     `json:"due_date,omitempty"`
Currency           string         `json:"currency"`
SourceSyntax       string         `json:"source_syntax"`
TargetJurisdiction string         `json:"target_jurisdiction,omitempty"`
Seller             Party          `json:"seller"`
Buyer              Party          `json:"buyer"`
Lines              []InvoiceLine  `json:"lines"`
TaxSubtotals       []TaxSubtotal  `json:"tax_subtotals"`
Totals             MonetaryTotals `json:"totals"`
PaymentMeansCode   string         `json:"payment_means_code,omitempty"`
PaymentReference   string         `json:"payment_reference,omitempty"`
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
