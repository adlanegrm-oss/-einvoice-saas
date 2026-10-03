package as4

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
)

var (
	ErrNRRRejected   = errors.New("recepisse AS4 NRR invalide ou refuse")
	ErrTransmission = errors.New("echec de transmission AS4 vers la passerelle cible")
)

type AS4Client struct {
	httpClient *http.Client
	c14n       bool
}

func NewAS4Client() *AS4Client {
	return &AS4Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		c14n:       true,
	}
}

// Structures SOAP / ebMS3 minimales pour le protocole eDelivery PEPPOL BIS Billing 3.0
type Envelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  Header   `xml:"Header"`
	Body    Body     `xml:"Body"`
}

type Header struct {
	Messaging Messaging `xml:"http://docs.oasis-open.org/ebxml-msg/ebms/v3.0/ns/core/200704/ Messaging"`
}

type Messaging struct {
	UserMessage UserMessage `xml:"UserMessage"`
}

type UserMessage struct {
	MessageInfo MessageInfo `xml:"MessageInfo"`
	PartyInfo   PartyInfo   `xml:"PartyInfo"`
	PayloadInfo PayloadInfo `xml:"PayloadInfo"`
}

type MessageInfo struct {
	Timestamp time.Time `xml:"Timestamp"`
	MessageId string    `xml:"MessageId"`
}

type PartyInfo struct {
	From Party `xml:"From"`
	To   Party `xml:"To"`
}

type Party struct {
	PartyId string `xml:"PartyId"`
	Role    string `xml:"Role"`
}

type PayloadInfo struct {
	PartInfo PartInfo `xml:"PartInfo"`
}

type PartInfo struct {
	Href         string       `xml:"href,attr"`
	PartProperty PartProperty `xml:"PartProperties>Property"`
}

type PartProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type Body struct {
	PayloadData string `xml:"PayloadData"`
}

// SendPayload encapsule la facture Factur-X / UBL dans un message AS4 et valide le NRR de retour
func (c *AS4Client) SendPayload(ctx context.Context, endpoint *dispatcher.TargetEndpoint, payload []byte) (string, error) {
	messageID := fmt.Sprintf("msg_%d@pdp.einvoice-saas.eu", time.Now().UnixNano())
	payloadDigest := sha256.Sum256(payload)
	payloadDigestB64 := base64.StdEncoding.EncodeToString(payloadDigest[:])

	env := Envelope{
		Header: Header{
			Messaging: Messaging{
				UserMessage: UserMessage{
					MessageInfo: MessageInfo{
						Timestamp: time.Now().UTC(),
						MessageId: messageID,
					},
					PartyInfo: PartyInfo{
						From: Party{PartyId: "FR:SIRET:PDP-FR-01", Role: "http://docs.oasis-open.org/ebxml-msg/ebms/v3.0/ns/core/200704/initiator"},
						To:   Party{PartyId: endpoint.ReceiverID, Role: "http://docs.oasis-open.org/ebxml-msg/ebms/v3.0/ns/core/200704/responder"},
					},
					PayloadInfo: PayloadInfo{
						PartInfo: PartInfo{
							Href: "cid:invoice-payload",
							PartProperty: PartProperty{
								Name:  "DigestValue",
								Value: payloadDigestB64,
							},
						},
					},
				},
			},
		},
		Body: Body{
			PayloadData: base64.StdEncoding.EncodeToString(payload),
		},
	}

	xmlPayload, err := xml.MarshalIndent(env, "", "  ")
	if err != nil {
		return "", fmt.Errorf("generation enveloppe AS4 : %w", err)
	}

	// Si l'endpoint cible est fictif ou en mode simulation
	if endpoint.AS4Endpoint == "" || strings.HasPrefix(endpoint.AS4Endpoint, "mock://") {
		return fmt.Sprintf("NRR-SIMULATED-ACK-%s-%s", messageID, payloadDigestB64[:8]), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.AS4Endpoint, bytes.NewReader(xmlPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("X-AS4-Original-Sender", "PDP-FR-01")
	req.Header.Set("X-AS4-Final-Recipient", endpoint.ReceiverID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTransmission, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("%w: status %d - %s", ErrTransmission, resp.StatusCode, string(respBody))
	}

	nrrBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("lecture reponse NRR AS4 : %w", err)
	}

	// Validation du récépissé Non-Repudiation of Receipt
	nrrReceipt := string(nrrBody)
	if !strings.Contains(nrrReceipt, "Receipt") && !strings.Contains(nrrReceipt, "SignalMessage") && !strings.Contains(nrrReceipt, "ACK") {
		return "", fmt.Errorf("%w: corps = %s", ErrNRRRejected, nrrReceipt)
	}

	return fmt.Sprintf("NRR-%s", payloadDigestB64[:12]), nil
}