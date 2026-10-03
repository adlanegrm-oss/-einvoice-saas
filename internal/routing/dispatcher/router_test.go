package dispatcher

import (
	"context"
	"errors"
	"testing"
)

// Mock du service d'annuaire
type mockDirectory struct {
	found bool
}

func (m *mockDirectory) Lookup(ctx context.Context, participantID string) (*TargetEndpoint, error) {
	if !m.found {
		return nil, errors.New("participant non enregistré")
	}
	return &TargetEndpoint{
		ReceiverID:   participantID,
		PlatformName: "Mock PDP Alpha",
		AS4Endpoint:  "https://as4.mock-pdp.fr/ebms",
	}, nil
}

// Mock du client AS4
type mockAS4Client struct{}

func (m *mockAS4Client) SendPayload(ctx context.Context, endpoint *TargetEndpoint, payload []byte) (string, error) {
	return "AS4-RECEIPT-OK-987654321", nil
}

func TestDispatcher_RouteAndDispatch(t *testing.T) {
	ctx := context.Background()

	// Cas 1 : Succès de routage et transmission
	disp := NewDispatcher(&mockDirectory{found: true}, &mockAS4Client{})
	receipt, err := disp.RouteAndDispatch(ctx, "12345678900014", []byte("<Invoice/>"))
	if err != nil {
		t.Fatalf("Le routage aurait dû réussir: %v", err)
	}
	if receipt == "" {
		t.Errorf("L'accusé de réception (receipt) ne doit pas être vide")
	}

	// Cas 2 : SIRET absent de l'annuaire
	dispNotFound := NewDispatcher(&mockDirectory{found: false}, &mockAS4Client{})
	_, err = dispNotFound.RouteAndDispatch(ctx, "00000000000000", []byte("<Invoice/>"))
	if err == nil {
		t.Fatalf("Le routage d'un SIRET inexistant aurait dû lever une erreur")
	}
}
