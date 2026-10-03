package service

import (
"context"
"database/sql"
"encoding/json"
"time"

"github.com/adlanegrm-oss/einvoice-saas/internal/audit"
"github.com/adlanegrm-oss/einvoice-saas/internal/domain"
"github.com/adlanegrm-oss/einvoice-saas/internal/tenant"
)

type InvoiceService struct {
db *sql.DB
}

func NewInvoiceService(db *sql.DB) *InvoiceService {
return &InvoiceService{db: db}
}

type EmitInvoiceCommand struct {
InvoiceNumber string                `json:"invoice_number"`
Profile       string                `json:"profile"`
Currency      string                `json:"currency"`
Totals        domain.MonetaryTotals `json:"totals"`
BuyerID       string                `json:"buyer_id"`
SellerID      string                `json:"seller_id"`
RawPayload    string                `json:"raw_payload"`
}

func (s *InvoiceService) EmitInvoice(ctx context.Context, cmd EmitInvoiceCommand) (string, error) {
tenantID, err := tenant.FromContext(ctx)
if err != nil {
return "", err
}

if err := cmd.Totals.Validate(); err != nil {
return "", err
}

tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
if err != nil {
return "", err
}
defer tx.Rollback()

var invoiceID string
insertInvoiceQuery := `
INSERT INTO invoices (
tenant_id, invoice_number, profile, 
compliance_status, transmission_status, payment_status,
currency, net_amount_cents, tax_amount_cents, gross_amount_cents,
buyer_identifier, seller_identifier, raw_payload_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id
`
err = tx.QueryRowContext(ctx, insertInvoiceQuery,
tenantID, cmd.InvoiceNumber, cmd.Profile,
domain.ComplianceValid, domain.TransmissionPending, domain.PaymentNotDue,
cmd.Currency, cmd.Totals.NetTotal.Cents(), cmd.Totals.TaxTotal.Cents(), cmd.Totals.GrossTotal.Cents(),
cmd.BuyerID, cmd.SellerID, cmd.RawPayload,
).Scan(&invoiceID)
if err != nil {
return "", err
}

payloadHash := audit.ComputePayloadHash([]byte(cmd.RawPayload))
eventHash := audit.ComputeRawEventHash("0000000000000000000000000000000000000000000000000000000000000000", payloadHash, "INVOICE_INGESTED", time.Now())

insertAuditQuery := `
INSERT INTO invoice_audit_events (
tenant_id, invoice_id, event_type, payload_hash, prev_event_hash, event_hash
) VALUES ($1, $2, 'INVOICE_INGESTED', $3, '0000000000000000000000000000000000000000000000000000000000000000', $4)
`
if _, err = tx.ExecContext(ctx, insertAuditQuery, tenantID, invoiceID, payloadHash, eventHash); err != nil {
return "", err
}

outboxPayload, _ := json.Marshal(map[string]interface{}{
"invoice_id":     invoiceID,
"invoice_number": cmd.InvoiceNumber,
"tenant_id":      tenantID,
"destination":    "PDP_ROUTING",
})

insertOutboxQuery := `
INSERT INTO outbox_events (
tenant_id, aggregate_type, aggregate_id, event_type, payload, status
) VALUES ($1, 'INVOICE', $2, 'DISPATCH_REQUESTED', $3, 'PENDING')
`
if _, err = tx.ExecContext(ctx, insertOutboxQuery, tenantID, invoiceID, outboxPayload); err != nil {
return "", err
}

if err := tx.Commit(); err != nil {
return "", err
}

return invoiceID, nil
}
