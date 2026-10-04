package as4

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AccessPointGateway route les messages vers un Access Point PEPPOL / AS4 e-Delivery certifié (Oxalis, Holodeck B2B, Phase4)
type AccessPointGateway struct {
	apEndpointURL string
	httpClient    *http.Client
}

type PeppolTransmissionRequest struct {
	SenderID       string `json:"sender_id"`       // ex: iso6523-actorid-upis::0002:73204903600045
	ReceiverID     string `json:"receiver_id"`     // ex: iso6523-actorid-upis::0002:80214589000012
	DocumentTypeID string `json:"document_type_id"` // ex: busdox-docid-qns::urn:oasis:names:specification:ubl:schema:xsd:Invoice-2::Invoice##urn:cen.eu:en16931:2017...
	ProcessID      string `json:"process_id"`      // ex: cenbii-procid-ubl::urn:fdc:peppol.eu:2017:poacc:billing:01:1.0
	PayloadXML     string `json:"payload_xml"`
}

type PeppolTransmissionReceipt struct {
	MessageID      string    `json:"message_id"`
	TransmissionID string    `json:"transmission_id"`
	Status         string    `json:"status"` // DELIVERED, REJECTED
	DeliveredAt    time.Time `json:"delivered_at"`
	NRRReceipt     string    `json:"nrr_receipt"` // Non-repudiation of receipt XML/Signature
}

func NewAccessPointGateway(apEndpointURL string, tlsClientCert *tls.Certificate) *AccessPointGateway {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if tlsClientCert != nil {
		tlsConfig.Certificates = []tls.Certificate{*tlsClientCert}
	}

	return &AccessPointGateway{
		apEndpointURL: apEndpointURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}
}

func (g *AccessPointGateway) DispatchPeppolInvoice(ctx context.Context, req PeppolTransmissionRequest) (*PeppolTransmissionReceipt, error) {
	raw, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.apEndpointURL+"/as4/outbound", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("access point communication failure: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("access point rejected with status %d: %s", resp.StatusCode, string(body))
	}

	var receipt PeppolTransmissionReceipt
	if err := json.NewDecoder(resp.Body).Decode(&receipt); err != nil {
		return nil, fmt.Errorf("decoding access point receipt: %w", err)
	}

	return &receipt, nil
}
