package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type JWTClaims struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	Audience  string   `json:"aud"`
	TenantID  string   `json:"tenant_id"`
	Role      string   `json:"role"`
	JTI       string   `json:"jti"`
	ExpiresAt int64    `json:"exp"`
	NotBefore int64    `json:"nbf"`
	IssuedAt  int64    `json:"iat"`
}

type JWTValidator struct {
	secret         []byte
	expectedIssuer string
	expectedAud    string
}

func NewJWTValidator(secret []byte, expectedIssuer, expectedAud string) *JWTValidator {
	return &JWTValidator{
		secret:         secret,
		expectedIssuer: expectedIssuer,
		expectedAud:    expectedAud,
	}
}

// ValidateToken effectue une vÃ©rification cryptographique stricte (iss, aud, exp, nbf, HMAC whitelist)
func (v *JWTValidator) ValidateToken(tokenString string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("jwt: invalid token format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid header encoding: %w", err)
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("jwt: invalid header json: %w", err)
	}

	// Whitelist stricte d'algorithme (prÃ©vention attaque 'none' ou injection asymÃ©trique)
	if header.Alg != "HS256" {
		return nil, fmt.Errorf("jwt: algorithm %s is not permitted (only HS256 allowed)", header.Alg)
	}

	// VÃ©rification de la signature HMAC-SHA256
	signingInput := parts[0] + "." + parts[1]
	expectedMac := hmac.New(sha256.New, v.secret)
	expectedMac.Write([]byte(signingInput))
	expectedSignature := expectedMac.Sum(nil)

	actualSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("jwt: invalid signature encoding")
	}

	if !hmac.Equal(actualSignature, expectedSignature) {
		return nil, errors.New("jwt: signature mismatch")
	}

	// DÃ©codage des claims
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("jwt: invalid payload encoding: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("jwt: invalid payload json: %w", err)
	}

	now := time.Now().Unix()

	// VÃ©rification exp
	if claims.ExpiresAt <= now {
		return nil, errors.New("jwt: token is expired")
	}
	// VÃ©rification nbf
	if claims.NotBefore > 0 && claims.NotBefore > now {
		return nil, errors.New("jwt: token not active yet")
	}
	// VÃ©rification issuer
	if v.expectedIssuer != "" && claims.Issuer != v.expectedIssuer {
		return nil, fmt.Errorf("jwt: invalid issuer %s", claims.Issuer)
	}
	// VÃ©rification audience
	if v.expectedAud != "" && claims.Audience != v.expectedAud {
		return nil, errors.New("jwt: invalid audience")
	}

	return &claims, nil
}

