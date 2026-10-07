package repository

import (
	"context"
	"testing"
)

func TestRepository_Smoke(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Fatal("repository context is nil")
	}
}
