package connector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type TransmissionReceipt struct {
	MessageID   string    `json:"message_id"`
	PartnerID   string    `json:"partner_id"`
	Status      string    `json:"status"` // ACCEPTED, REJECTED
	ReceiptHash string    `json:"receipt_hash"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type NetworkConnector interface {
	Submit(ctx context.Context, tenantID, invoiceID string, payload []byte) (*TransmissionReceipt, error)
	GetStatus(ctx context.Context, messageID string) (string, error)
}

type SandboxPDPConnector struct {
	PartnerName string
	FailRate    bool
}

func NewSandboxPDPConnector(partnerName string) *SandboxPDPConnector {
	return &SandboxPDPConnector{PartnerName: partnerName}
}

func (c *SandboxPDPConnector) Submit(ctx context.Context, tenantID, invoiceID string, payload []byte) (*TransmissionReceipt, error) {
	if c.FailRate {
		return nil, errors.New("503 SERVICE_UNAVAILABLE: Sandbox PDP temporarily down")
	}

	raw := fmt.Sprintf("%s:%s:%s", c.PartnerName, invoiceID, time.Now().Format(time.RFC3339Nano))
	h := sha256.Sum256([]byte(raw))
	receiptHash := hex.EncodeToString(h[:])

	return &TransmissionReceipt{
		MessageID:   fmt.Sprintf("msg-%s-%d", invoiceID, time.Now().UnixNano()),
		PartnerID:   c.PartnerName,
		Status:      "ACCEPTED",
		ReceiptHash: receiptHash,
		SubmittedAt: time.Now().UTC(),
	}, nil
}

func (c *SandboxPDPConnector) GetStatus(ctx context.Context, messageID string) (string, error) {
	return "ACCEPTED", nil
}
