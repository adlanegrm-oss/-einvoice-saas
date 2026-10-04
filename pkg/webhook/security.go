package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

var (
	ErrInvalidSignature = errors.New("signature webhook HMAC invalide")
	ErrTimestampSkew    = errors.New("ecart de timestamp webhook trop eleve (risque de rejeu)")
)

func VerifyHMACWebhook(payload []byte, signatureHex string, timestampHeader string, secret string, maxSkewSeconds float64) error {
	tsInt, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("timestamp invalide : %w", err)
	}

	msgTime := time.Unix(tsInt, 0)
	diff := math.Abs(time.Since(msgTime).Seconds())
	if diff > maxSkewSeconds {
		return ErrTimestampSkew
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestampHeader))
	mac.Write([]byte("."))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)

	actualMAC, err := hex.DecodeString(signatureHex)
	if err != nil {
		return ErrInvalidSignature
	}

	if !hmac.Equal(actualMAC, expectedMAC) {
		return ErrInvalidSignature
	}

	return nil
}
