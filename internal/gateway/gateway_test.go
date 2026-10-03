package gateway_test

import (
"context"
"testing"

"github.com/adlanegrm-oss/einvoice-saas/internal/gateway"
)

func TestGatewayMock_IncidentSimulation(t *testing.T) {
ctx := context.Background()

// 1. Test Timeout
gwTimeout := gateway.NewMockGateway(gateway.MockTimeout)
_, err := gwTimeout.Submit(ctx, "tenant-1", "inv-1", []byte("xml"))
if err != gateway.ErrNetworkTimeout {
t.Fatalf("Attendu ErrNetworkTimeout, obtenu %v", err)
}

// 2. Test Doublon
gwDup := gateway.NewMockGateway(gateway.MockDuplicate)
_, err = gwDup.Submit(ctx, "tenant-1", "inv-1", []byte("xml"))
if err != gateway.ErrDuplicateMessage {
t.Fatalf("Attendu ErrDuplicateMessage, obtenu %v", err)
}

// 3. Test Succès
gwOk := gateway.NewMockGateway(gateway.MockSuccess)
receipt, err := gwOk.Submit(ctx, "tenant-1", "inv-1", []byte("xml"))
if err != nil || receipt.Status != "ACCEPTED" {
t.Fatalf("Reçu non conforme en mode succès : %v", err)
}
}
