package observability

import (
	"context"
	"testing"
)

func TestObservability_Smoke(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Fatal("observability context is nil")
	}
}
