package clearance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"einvoice-saas/internal/model"
)

// Channel identifie le canal de transmission.
type Channel string

const (
	ChannelPPF  Channel = "PPF"  // Chorus Pro / Piste France
	ChannelAS4  Channel = "AS4"  // PEPPOL / ebMS3
	ChannelKSeF Channel = "KSEF" // Pologne
	ChannelNone Channel = "NONE"
)

// TransmissionRequest regroupe payload + métadonnées d'acheminement.
type TransmissionRequest struct {
	InvoiceID   string
	TenantID    string
	Channel     Channel
	Canonical   *model.CanonicalInvoice
	XMLPayload  []byte // UBL ou CII déjà sérialisé
	PDFPayload  []byte // optionnel Factur-X
	MessageID   string
	Correlation string
}

// TransmissionResult statut côté connecteur.
type TransmissionResult struct {
	Channel      Channel
	MessageID    string
	RemoteStatus string
	Accepted     bool
	SubmittedAt  time.Time
	RawResponse  []byte
	ErrorMessage string
}

// Connector contract minimal pour tous les canaux.
type Connector interface {
	Name() Channel
	Submit(ctx context.Context, req TransmissionRequest) (*TransmissionResult, error)
}

// Dispatcher route vers le bon connecteur selon Channel ou heuristique pays.
type Dispatcher struct {
	connectors map[Channel]Connector
}

func NewDispatcher(connectors ...Connector) *Dispatcher {
	d := &Dispatcher{connectors: make(map[Channel]Connector)}
	for _, c := range connectors {
		if c != nil {
			d.connectors[c.Name()] = c
		}
	}
	return d
}

func (d *Dispatcher) Register(c Connector) {
	if c != nil {
		d.connectors[c.Name()] = c
	}
}

// ResolveChannel déduit le canal si non fourni (FR→PPF, MA→AS4, PL→KSEF).
func ResolveChannel(inv *model.CanonicalInvoice, explicit Channel) Channel {
	if explicit != "" && explicit != ChannelNone {
		return explicit
	}
	if inv == nil {
		return ChannelPPF
	}
	cc := strings.ToUpper(strings.TrimSpace(inv.Seller.CountryCode))
	switch cc {
	case "PL":
		return ChannelKSeF
	case "MA":
		return ChannelAS4
	case "FR", "":
		return ChannelPPF
	default:
		return ChannelAS4
	}
}

// Transmit valide les préconditions minimales puis délègue au connecteur.
func (d *Dispatcher) Transmit(ctx context.Context, req TransmissionRequest) (*TransmissionResult, error) {
	if req.Canonical == nil && len(req.XMLPayload) == 0 {
		return nil, fmt.Errorf("clearance: canonical ou XMLPayload requis")
	}
	req.Channel = ResolveChannel(req.Canonical, req.Channel)
	conn, ok := d.connectors[req.Channel]
	if !ok {
		return nil, fmt.Errorf("clearance: aucun connecteur pour canal %s", req.Channel)
	}
	if req.MessageID == "" {
		req.MessageID = fmt.Sprintf("msg-%s-%d", req.InvoiceID, time.Now().UnixNano())
	}
	return conn.Submit(ctx, req)
}
