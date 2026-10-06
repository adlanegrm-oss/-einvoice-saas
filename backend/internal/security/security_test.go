package security

import (
"crypto/sha256"
"testing"
)

func TestSecurity_Smoke(t *testing.T) {
h := sha256.Sum256([]byte("payload"))
if len(h) != 32 {
t.Fatalf("invalid sha256 length: %d", len(h))
}
}