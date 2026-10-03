package gateway_test

import (
	"context"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/gateway"
)

func TestGatewayContract_ExhaustiveScenarios(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		mode        gateway.MockMode
		expectedErr error
	}{
		{gateway.MockTimeout, gateway.ErrNetworkTimeout},
		{gateway.MockDuplicate, gateway.ErrDuplicateMessage},
		{gateway.MockReject, gateway.ErrPartnerRejected},
		{gateway.MockRateLimit, gateway.ErrRateLimited},
		{gateway.MockServerError, gateway.ErrServerError},
		{gateway.MockInvalidResponse, gateway.ErrInvalidResponse},
	}

	for _, tc := range cases {
		t.Run(string(tc.mode), func(t *testing.T) {
			gw := gateway.NewContractMockGateway(tc.mode)
			_, err := gw.Submit(ctx, "tenant-test", "inv-test", []byte("xml-content"))
			if err != tc.expectedErr {
				t.Fatalf("Mode %s: attendu %v, obtenu %v", tc.mode, tc.expectedErr, err)
			}
		})
	}

	// Cas de succès nominal
	gwOk := gateway.NewContractMockGateway(gateway.MockSuccess)
	receipt, err := gwOk.Submit(ctx, "tenant-test", "inv-test", []byte("xml-content"))
	if err != nil || receipt.Status != "ACCEPTED" || receipt.ReceiptHash == "" {
		t.Fatalf("Mode SUCCESS échoué: receipt=%+v, err=%v", receipt, err)
	}
}
