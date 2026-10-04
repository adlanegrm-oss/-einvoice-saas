package clearance

import (
	"context"
	"fmt"
	"time"
)

type KSeFConnector struct {
	BaseURL string
	NIP     string
}

func NewKSeFConnector(baseURL, nip string) *KSeFConnector {
	return &KSeFConnector{BaseURL: baseURL, NIP: nip}
}

type KSeFReceipt struct {
	KSeFReferenceNumber string    `json:"ksef_reference_number"`
	AcquisitionTimestamp time.Time `json:"acquisition_timestamp"`
	Status               string    `json:"status"` // ACCEPTED, PENDING, REJECTED
}

func (k *KSeFConnector) SendFA2Invoice(ctx context.Context, invoiceXML []byte) (*KSeFReceipt, error) {
	if len(invoiceXML) == 0 {
		return nil, fmt.Errorf("ksef: empty FA(2) invoice xml")
	}

	// Simulation conforme à l'API asynchrone KSeF 2.0 (Pologne)
	refNum := fmt.Sprintf("%s-20261004-%08d", k.NIP, time.Now().Unix()%100000000)
	return &KSeFReceipt{
		KSeFReferenceNumber:  refNum,
		AcquisitionTimestamp: time.Now().UTC(),
		Status:               "ACCEPTED",
	}, nil
}
