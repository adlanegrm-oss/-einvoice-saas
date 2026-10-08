package clearance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AS4Connector enveloppe un envoi ebMS3/AS4 (stub dry-run).
type AS4Connector struct {
	Endpoint string
	DryRun   bool
}

func NewAS4Connector(endpoint string, dryRun bool) *AS4Connector {
	return &AS4Connector{Endpoint: endpoint, DryRun: dryRun}
}

func (c *AS4Connector) Name() Channel { return ChannelAS4 }

func (c *AS4Connector) Submit(ctx context.Context, req TransmissionRequest) (*TransmissionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if len(req.XMLPayload) == 0 {
		return nil, fmt.Errorf("as4: XMLPayload vide")
	}
	result := &TransmissionResult{
		Channel:      ChannelAS4,
		MessageID:    req.MessageID,
		SubmittedAt:  time.Now().UTC(),
		Accepted:     true,
		RemoteStatus: "SENT",
	}
	if c.DryRun {
		payload, _ := json.Marshal(map[string]string{
			"status":     "SENT",
			"message_id": req.MessageID,
			"endpoint":   c.Endpoint,
			"mode":       "dry-run",
		})
		result.RawResponse = payload
		return result, nil
	}
	return result, fmt.Errorf("as4: mode live non configuré")
}
