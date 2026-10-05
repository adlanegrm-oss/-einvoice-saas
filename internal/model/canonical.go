package model

import "time"

// Party représente une entité commerciale (vendeur ou acheteur)
type Party struct {
	Name            string            `json:"name"`
	TaxID           string            `json:"tax_id,omitempty"`           // SIREN / SIRET / TVA FR
	NationalID      string            `json:"national_id,omitempty"`      // ICE (MA), etc.
	Country         string            `json:"country"`                    // Code pays ISO 3166-1 alpha-2 (FR, MA, etc.)
	AddressLine     string            `json:"address_line,omitempty"`
	City            string            `json:"city,omitempty"`
	PostalZone      string            `json:"postal_zone,omitempty"`
	Identifiers     map[string]string `json:"identifiers,omitempty"`      // SchemeID -> Value
}

// TaxSubtotal détaille les montants par catégorie ou taux de taxe
type TaxSubtotal struct {
	TaxableAmount float64 `json:"taxable_amount"`
	TaxAmount     float64 `json:"tax_amount"`
	Percent       float64 `json:"percent"`
	CategoryCode  string  `json:"category_code,omitempty"` // Standard S, Z, E, etc.
}

// InvoiceLine détaille un article ou une prestation facturée
type InvoiceLine struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	LineTotal   float64 `json:"line_total"`
	VatPercent  float64 `json:"vat_percent"`
}

// MonetaryTotals regroupe l'équilibre financier de la facture
type MonetaryTotals struct {
	LineExtensionAmount float64 `json:"line_extension_amount"` // Total HT des lignes
	TaxExclusiveAmount  float64 `json:"tax_exclusive_amount"`  // Base imposable totale
	TaxInclusiveAmount  float64 `json:"tax_inclusive_amount"`  // Total TTC
	AllowanceTotal      float64 `json:"allowance_total,omitempty"`
	ChargeTotal         float64 `json:"charge_total,omitempty"`
	PrepaidAmount       float64 `json:"prepaid_amount,omitempty"`
	PayableAmount       float64 `json:"payable_amount"`        // Reste net à payer
}

// CanonicalInvoice constitue le modèle universel pivot CC&VP
type CanonicalInvoice struct {
	ID                 string          `json:"id"`
	InvoiceNumber      string          `json:"invoice_number"`
	IssueDate          time.Time       `json:"issue_date"`
	DueDate            *time.Time      `json:"due_date,omitempty"`
	Currency           string          `json:"currency"`
	SourceSyntax       string          `json:"source_syntax"` // UBL, CII, FACTUR-X, EDIFACT
	TargetJurisdiction string          `json:"target_jurisdiction,omitempty"` // FR, MA, etc.
	Seller             Party           `json:"seller"`
	Buyer              Party           `json:"buyer"`
	Lines              []InvoiceLine   `json:"lines"`
	TaxSubtotals       []TaxSubtotal   `json:"tax_subtotals"`
	Totals             MonetaryTotals  `json:"totals"`
	PaymentMeansCode   string          `json:"payment_means_code,omitempty"`
	PaymentReference   string          `json:"payment_reference,omitempty"`
}
// IssueSeverity définit la criticité d'une non-conformité
type IssueSeverity string

const (
	SeverityError   IssueSeverity = "ERROR"
	SeverityWarning IssueSeverity = "WARNING"
)

// ValidationIssue décrit une non-conformité détectée
type ValidationIssue struct {
	RuleID      string        `json:"rule_id"`
	Description string        `json:"description"`
	Severity    IssueSeverity `json:"severity"`
	Field       string        `json:"field,omitempty"`
	Remediation string        `json:"remediation,omitempty"`
}

// ValidationReport synthétise le verdict de validation
type ValidationReport struct {
	Jurisdiction string            `json:"jurisdiction"`
	Valid        bool              `json:"valid"`
	Issues       []ValidationIssue `json:"issues,omitempty"`
}

// JurisdictionValidator est l'interface pivot implémentée par chaque pays
type JurisdictionValidator interface {
	JurisdictionCode() string
	Validate(invoice *CanonicalInvoice) ValidationReport
}