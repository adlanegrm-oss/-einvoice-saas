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
ErrRateLimited       = errors.New("429 TOO_MANY_REQUESTS: quota d'émission dépassé sur la plateforme tierce")
ErrServerError       = errors.New("503 SERVICE_UNAVAILABLE: passerelle PDP en maintenance ou indisponible")
ErrInvalidResponse   = errors.New("502 BAD_GATEWAY: réponse réseau malformée ou non parseable")
)

type Receipt struct {
MessageID   string    `json:"message_id"`
TrackingID  string    `json:"tracking_id"`
ReceiptHash string    `json:"receipt_hash"`
Status      string    `json:"status"` // ACCEPTED, DELIVERED, REJECTED
ReceivedAt  time.Time `json:"received_at"`
}

type Gateway interface {
Submit(ctx context.Context, tenantID, invoiceID string, payload []byte) (*Receipt, error)
Status(ctx context.Context, trackingID string) (string, error)
}

type MockMode string

const (
MockSuccess         MockMode = "SUCCESS"
MockTimeout         MockMode = "TIMEOUT"
MockDuplicate       MockMode = "DUPLICATE"
MockReject          MockMode = "REJECT"
MockRateLimit       MockMode = "RATE_LIMIT"
MockServerError     MockMode = "SERVER_ERROR"
MockInvalidResponse MockMode = "INVALID_RESPONSE"
)

type ContractMockGateway struct {
Mode MockMode
}

func NewContractMockGateway(mode MockMode) *ContractMockGateway {
return &ContractMockGateway{Mode: mode}
}

func (m *ContractMockGateway) Submit(ctx context.Context, tenantID, invoiceID string, payload []byte) (*Receipt, error) {
switch m.Mode {
case MockTimeout:
return nil, ErrNetworkTimeout
case MockDuplicate:
return nil, ErrDuplicateMessage
case MockReject:
return nil, ErrPartnerRejected
case MockRateLimit:
return nil, ErrRateLimited
case MockServerError:
return nil, ErrServerError
case MockInvalidResponse:
return nil, ErrInvalidResponse
case MockSuccess:
fallthrough
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

func (m *ContractMockGateway) Status(ctx context.Context, trackingID string) (string, error) {
if m.Mode == MockServerError {
return "", ErrServerError
}
return "DELIVERED", nil
}
