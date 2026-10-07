package clearance

import (
	"context"
	"testing"
)

func TestClearance_Smoke(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Fatal("context is nil")
	}
}
