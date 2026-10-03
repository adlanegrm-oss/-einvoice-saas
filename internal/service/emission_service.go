package service

import (
"context"
"database/sql"
"encoding/json"
"errors"
"fmt"
"time"

"github.com/adlanegrm-oss/einvoice-saas/internal/idempotency"
"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

var (
ErrDuplicateInvoiceNumber = errors.New("numéro de facture déjà utilisé pour cette organisation")
)

type MultiTenantEmissionService struct {
db          *sql.DB
idempotency *idempotency.Store
}

func NewMultiTenantEmissionService(db *sql.DB) *MultiTenantEmissionService {
return &MultiTenantEmissionService{
db:          db,
idempotency: idempotency.NewStore(db),
}
}

type EmissionResult struct {
InvoiceID string `json:"invoice_id"`
Status    string `json:"status"`
Cached    bool   `json:"cached"`
}

type itemInputHash struct {
Description string  `json:"desc"`
Quantity    float64 `json:"qty"`
UnitPrice   int64   `json:"price_amount"`
VATRate     int64   `json:"vat_amount"`
}

type invoiceInputHash struct {
Number   string          `json:"number"`
Customer string          `json:"customer"`
Items    []itemInputHash `json:"items"`
}

func (s *MultiTenantEmissionService) EmitInvoiceTx(
ctx context.Context, 
tenantID string, 
idempKey string, 
inv *invoice.Invoice,
) (*EmissionResult, error) {
if tenantID == "" {
return nil, errors.New("tenant_id obligatoire")
}

// 1. Calcul du hash d'idempotence canonique et immuable
itemsH := make([]itemInputHash, len(inv.Items))
for idx, it := range inv.Items {
itemsH[idx] = itemInputHash{
Description: it.Description,
Quantity:    it.Quantity,
UnitPrice:   it.UnitPrice.Amount,
VATRate:     it.VATRate.Amount,
}
}
canonicalInput := invoiceInputHash{
Number:   inv.Number,
Customer: inv.Customer.Name,
Items:    itemsH,
}
payloadRaw, _ := json.Marshal(canonicalInput)
reqHash := idempotency.ComputeHash(payloadRaw)

tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
if err != nil {
return nil, fmt.Errorf("begin tx: %w", err)
}
defer tx.Rollback()

// 2. Contrôle d'Idempotence persistant scellé au tenant
if idempKey != "" {
cached, err := s.idempotency.LockKey(ctx, tx, tenantID, idempKey, reqHash, 24*time.Hour)
if err != nil {
return nil, err
}
if cached != nil {
var res EmissionResult
if err := json.Unmarshal([]byte(cached.ResponseBody), &res); err == nil {
res.Cached = true
return &res, nil
}
}
}

// 3. Validation métier
if err := inv.Validate(); err != nil {
return nil, fmt.Errorf("validation facture: %w", err)
}

// 4. Clé primaire composite/cloisonnée
invoiceID := fmt.Sprintf("%s-%s-%d", tenantID, inv.Number, time.Now().UnixNano())
inv.ID = invoiceID

insertInvoice := `
INSERT INTO invoices_v2 (
id, tenant_id, invoice_number, status, currency, issue_date, 
seller_json, customer_json, total_ht_cents, total_vat_cents, total_ttc_cents, is_validated
) VALUES (?, ?, ?, 'SUBMISSION_PENDING', ?, ?, ?, ?, ?, ?, ?, 1)`

customerJSON, _ := json.Marshal(inv.Customer)
sellerJSON, _ := json.Marshal(inv.Seller)

_, err = tx.ExecContext(ctx, insertInvoice,
invoiceID, tenantID, inv.Number, string(inv.Currency), inv.IssueDate,
string(sellerJSON), string(customerJSON),
inv.TotalHT.Amount, inv.TotalVAT.Amount, inv.TotalTTC.Amount,
)
if err != nil {
return nil, ErrDuplicateInvoiceNumber
}

// 5. Outbox transactionnel rattaché au tenant
outboxEventID := fmt.Sprintf("outbox-%d", time.Now().UnixNano())
insertOutbox := `
INSERT INTO outbox_events (id, tenant_id, aggregate_id, event_type, payload_json, status, created_at)
VALUES (?, ?, ?, 'INVOICE_SUBMISSION_REQUESTED', ?, 'PENDING', ?)`

now := time.Now().UTC()
if _, err := tx.ExecContext(ctx, insertOutbox, outboxEventID, tenantID, invoiceID, string(payloadRaw), now); err != nil {
return nil, fmt.Errorf("insertion outbox: %w", err)
}

// 6. Enregistrement réponse d'idempotence
finalResult := EmissionResult{
InvoiceID: invoiceID,
Status:    "SUBMISSION_PENDING",
Cached:    false,
}
resBytes, _ := json.Marshal(finalResult)

if idempKey != "" {
if err := s.idempotency.Complete(ctx, tx, tenantID, idempKey, 201, string(resBytes)); err != nil {
return nil, fmt.Errorf("finalisation idempotence: %w", err)
}
}

if err := tx.Commit(); err != nil {
return nil, fmt.Errorf("commit tx: %w", err)
}

return &finalResult, nil
}
