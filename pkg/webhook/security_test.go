package webhook

import (
"crypto/hmac"
"crypto/sha256"
"encoding/hex"
"fmt"
"testing"
"time"
)

func TestVerifyHMACWebhook(t *testing.T) {
secret := "secret-token-key-2026"
payload := []byte(`{"event":"INVOICE_ACCEPTED","invoice_id":"INV-123"}`)
nowTs := fmt.Sprintf("%d", time.Now().Unix())

mac := hmac.New(sha256.New, []byte(secret))
mac.Write([]byte(nowTs))
mac.Write([]byte("."))
mac.Write(payload)
sigHex := hex.EncodeToString(mac.Sum(nil))

err := VerifyHMACWebhook(payload, sigHex, nowTs, secret, 300)
if err != nil {
t.Fatalf("signature valide rejetee: %v", err)
}

// Test mauvaise signature
errBadSig := VerifyHMACWebhook(payload, "badf00d", nowTs, secret, 300)
if errBadSig == nil {
t.Fatal("mauvaise signature acceptee a tort")
}

// Test skew timestamp (rejeu ancien)
oldTs := fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())
errSkew := VerifyHMACWebhook(payload, sigHex, oldTs, secret, 300)
if errSkew != ErrTimestampSkew {
t.Fatalf("timestamp expire non detecte: %v", errSkew)
}
}