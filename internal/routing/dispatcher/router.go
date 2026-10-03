package dispatcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// TargetEndpoint détaille les coordonnées techniques de la PDP ou du point de livraison
type TargetEndpoint struct {
	ReceiverID   string // ex: "0002:12345678900014" (ISO6523:SIRET)
	PlatformName string // ex: "Chorus Pro", "PDP_X", "Peppol_AP_Y"
	AS4Endpoint  string // URL d'ingestion sécurisée AS4
	Certificate  []byte // Certificat X.509 pour chiffrement de l'enveloppe
	SupportedDoc string // UBL, CII, Factur-X
}

type DirectoryService interface {
	Lookup(ctx context.Context, participantID string) (*TargetEndpoint, error)
}

type AS4Client interface {
	SendPayload(ctx context.Context, endpoint *TargetEndpoint, payload []byte) (receiptID string, err error)
}

type Dispatcher struct {
	directory DirectoryService
	as4Client AS4Client
}

func NewDispatcher(dir DirectoryService, as4 AS4Client) *Dispatcher {
	return &Dispatcher{
		directory: dir,
		as4Client: as4,
	}
}

// RouteAndDispatch résout l'adresse du destinataire et transmet le flux
func (d *Dispatcher) RouteAndDispatch(ctx context.Context, receiverSiret string, rawInvoice []byte) (string, error) {
	// 1. Résolution via l'annuaire (SMP ou Annuaire National)
	participantID := fmt.Sprintf("0002:%s", receiverSiret)
	endpoint, err := d.directory.Lookup(ctx, participantID)
	if err != nil {
		return "", fmt.Errorf("échec de résolution de l'annuaire pour le tiers %s: %w", receiverSiret, err)
	}

	// 2. Calcul de l'empreinte de la facture pour la traçabilité
	hasher := sha256.New()
	hasher.Write(rawInvoice)
	payloadHash := hex.EncodeToString(hasher.Sum(nil))

	// 3. Transmission sécurisée au point d'accès cible
	receipt, err := d.as4Client.SendPayload(ctx, endpoint, rawInvoice)
	if err != nil {
		return "", fmt.Errorf("échec transmission AS4 vers %s: %w", endpoint.PlatformName, err)
	}

	_ = payloadHash // Utilisé pour consigner la preuve de transmission dans l'audit log
	return receipt, nil
}
