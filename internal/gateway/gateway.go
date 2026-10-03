package gateway

import (
"context"
"crypto/sha256"
"encoding/hex"
"errors"
"fmt"
"time"
)

var (
ErrNetworkTimeout    = errors.New("504 GATEWAY_TIMEOUT: le tiers PDP n'a pas répondu dans le délai imparti")
ErrDuplicateMessage  = errors.New("409 CONFLICT: message déjà reçu et traité par le réseau")
ErrPartnerRejected   = errors.New("422 UNPROCESSABLE_ENTITY: facture rejetée par la plateforme destinataire")
ErrServiceDown       = errors.New("503 SERVICE_UNAVAILABLE: passerelle indisponible")
)

type Receipt struct {
MessageID   string    `json:"message_id"`
TrackingID  string    `json:"tracking_id"`
ReceiptHash string    `json:"receipt_hash"`
Status      string    `json:"status"`
ReceivedAt  time.Time `json:"received_at"`
}

type Gateway interface {
Submit(ctx context.Context, tenantID, invoiceID string, payload []byte) (*Receipt, error)
CheckStatus(ctx context.Context, trackingID string) (string, error)
}

type MockMode string

const (
MockSuccess   MockMode = "SUCCESS"
MockTimeout   MockMode = "TIMEOUT"
MockDuplicate MockMode = "DUPLICATE"
MockReject    MockMode = "REJECT"
MockDown      MockMode = "DOWN"
)

type AdvancedMockGateway struct {
Mode MockMode
}

func NewMockGateway(mode MockMode) *AdvancedMockGateway {
return &AdvancedMockGateway{Mode: mode}
}

func (m *AdvancedMockGateway) Submit(ctx context.Context, tenantID, invoiceID string, payload []byte) (*Receipt, error) {
switch m.Mode {
case MockTimeout:
return nil, ErrNetworkTimeout
case MockDuplicate:
return nil, ErrDuplicateMessage
case MockReject:
return nil, ErrPartnerRejected
case MockDown:
return nil, ErrServiceDown
default:
h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", tenantID, invoiceID, time.Now().UnixNano())))
return &Receipt{
MessageID:   fmt.Sprintf("msg-%d", time.Now().UnixNano()),
TrackingID:  fmt.Sprintf("trk-%s", invoiceID),
ReceiptHash: hex.EncodeToString(h[:]),
Status:      "ACCEPTED",
ReceivedAt:  time.Now().UTC(),
}, nil
}
}

func (m *AdvancedMockGateway) CheckStatus(ctx context.Context, trackingID string) (string, error) {
return "DELIVERED", nil
}
