package clearance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PPFConnector simule / prépare l'appel PISTE Chorus Pro.
// En prod: OAuth2 client_credentials + POST dépôt facture.
type PPFConnector struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	DryRun       bool
}

func NewPPFConnector(baseURL, clientID, clientSecret string, dryRun bool) *PPFConnector {
	if baseURL == "" {
		baseURL = "https://api.piste.gouv.fr"
	}
	return &PPFConnector{BaseURL: baseURL, ClientID: clientID, ClientSecret: clientSecret, DryRun: dryRun}
}

func (c *PPFConnector) Name() Channel { return ChannelPPF }

func (c *PPFConnector) Submit(ctx context.Context, req TransmissionRequest) (*TransmissionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if len(req.XMLPayload) == 0 {
		return nil, fmt.Errorf("ppf: XMLPayload vide")
	}

	result := &TransmissionResult{
		Channel:      ChannelPPF,
		MessageID:    req.MessageID,
		SubmittedAt:  time.Now().UTC(),
		Accepted:     true,
		RemoteStatus: "DEPOSEE",
	}

	if c.DryRun {
		payload, _ := json.Marshal(map[string]string{
			"status":     "DEPOSEE",
			"message_id": req.MessageID,
			"invoice_id": req.InvoiceID,
			"mode":       "dry-run",
		})
		result.RawResponse = payload
		return result, nil
	}

	// Point d'extension: appel HTTP réel PISTE
	return result, fmt.Errorf("ppf: mode live non configuré (renseigner credentials + HTTP client)")
}
