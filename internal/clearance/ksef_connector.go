package clearance

import (
"context"
"encoding/json"
"fmt"
"time"
)

// KSeFConnector stub session d'autorisation + envoi FA(2).
type KSeFConnector struct {
BaseURL string
DryRun  bool
}

func NewKSeFConnector(baseURL string, dryRun bool) *KSeFConnector {
if baseURL == "" {
baseURL = "https://ksef.mf.gov.pl"
}
return &KSeFConnector{BaseURL: baseURL, DryRun: dryRun}
}

func (c *KSeFConnector) Name() Channel { return ChannelKSeF }

func (c *KSeFConnector) Submit(ctx context.Context, req TransmissionRequest) (*TransmissionResult, error) {
select {
case <-ctx.Done():
return nil, ctx.Err()
default:
}
if len(req.XMLPayload) == 0 {
return nil, fmt.Errorf("ksef: XMLPayload vide")
}
result := &TransmissionResult{
Channel:      ChannelKSeF,
MessageID:    req.MessageID,
SubmittedAt:  time.Now().UTC(),
Accepted:     true,
RemoteStatus: "ACCEPTED",
}
if c.DryRun {
payload, _ := json.Marshal(map[string]string{
"status":     "ACCEPTED",
"message_id": req.MessageID,
"mode":       "dry-run",
})
result.RawResponse = payload
return result, nil
}
return result, fmt.Errorf("ksef: mode live non configuré")
}
