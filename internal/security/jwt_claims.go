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

var (
	ErrTokenInvalid = errors.New("security: token invalid")
	ErrTokenExpired = errors.New("security: token expired")
	ErrTokenClaims  = errors.New("security: claims invalid")
)

type Claims struct {
	Subject   string   `json:"sub"`
	Issuer    string   `json:"iss"`
	Audience  string   `json:"aud"`
	TenantID  string   `json:"tenant_id"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	NotBefore int64    `json:"nbf"`
	IssuedAt  int64    `json:"iat"`
}

func SignHS256(secret []byte, c Claims) (string, error) {
	header := base64.RawURLEncoding.EncodeToString(
		[]byte(`{"alg":"HS256","typ":"JWT"}`),
	)

	raw, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	payload := base64.RawURLEncoding.EncodeToString(raw)
	signingInput := header + "." + payload

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signingInput))

	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

// ParseAndValidate conserve l'API attendue par les nouveaux tests,
// sans créer un second JWTValidator.
func (v *JWTValidator) ParseAndValidate(token string) (*Claims, error) {
	claims, err := v.ValidateToken(token)
	if err != nil {
		msg := err.Error()

		switch {
		case strings.Contains(msg, "expired"):
			return nil, ErrTokenExpired
		case strings.Contains(msg, "issuer"),
			strings.Contains(msg, "audience"),
			strings.Contains(msg, "tenant"),
			strings.Contains(msg, "claims"):
			return nil, ErrTokenClaims
		default:
			return nil, ErrTokenInvalid
		}
	}

	return &Claims{
		Subject:   claims.Subject,
		Issuer:    claims.Issuer,
		Audience:  claims.Audience,
		TenantID:  claims.TenantID,
		Roles:     []string{claims.Role},
		ExpiresAt: claims.ExpiresAt,
		NotBefore: claims.NotBefore,
		IssuedAt:  claims.IssuedAt,
	}, nil
}

// évite un import inutilisé si le fichier est modifié ultérieurement.
var _ = fmt.Sprintf
var _ = time.Now
