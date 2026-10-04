package pdp

import (
"bytes"
"context"
"encoding/json"
"errors"
"fmt"
"net/http"
"time"
)

type PDPClient struct {
BaseURL    string
APIKey     string
HTTPClient *http.Client
}

func NewPDPClient(baseURL, apiKey string) *PDPClient {
return &PDPClient{
BaseURL: baseURL,
APIKey:  apiKey,
HTTPClient: &http.Client{
Timeout: 15 * time.Second,
},
}
}

type SubmissionPayload struct {
TenantID       string `json:"tenant_id"`
InvoiceNumber  string `json:"invoice_number"`
Format         string `json:"format"` // FACTURX, UBL, CII
RawData        string `json:"raw_data"`
IdempotencyKey string `json:"idempotency_key"`
}

type SubmissionResult struct {
TransmissionID string `json:"transmission_id"`
LifecycleCode  string `json:"lifecycle_code"` // 200, 201, 400, etc.
Status         string `json:"status"`         // DEPOSE, REJETE, VALIDE
TrackingURL    string `json:"tracking_url,omitempty"`
}

func (c *PDPClient) SubmitInvoice(ctx context.Context, payload SubmissionPayload) (*SubmissionResult, error) {
bodyBytes, err := json.Marshal(payload)
if err != nil {
return nil, err
}

endpoint := fmt.Sprintf("%s/api/v1/invoices/submit", strings.TrimRight(c.BaseURL, "/"))
req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
if err != nil {
return nil, err
}

req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+c.APIKey)
req.Header.Set("X-Idempotency-Key", payload.IdempotencyKey)

// Exécution HTTP avec capture fine du statut
resp, err := c.HTTPClient.Do(req)
if err != nil {
return nil, fmt.Errorf("echec reseau transmission PDP : %w", err)
}
defer resp.Body.Close()

if resp.StatusCode == http.StatusConflict {
return nil, errors.New("pdp: doublon facture detecte (deja traitee)")
}

if resp.StatusCode >= 400 {
return nil, fmt.Errorf("pdp: rejet partenaire HTTP %d", resp.StatusCode)
}

var res SubmissionResult
if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
return nil, fmt.Errorf("pdp: reponse JSON illisible : %w", err)
}

return &res, nil
}