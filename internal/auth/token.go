package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Role identifie le profil d'un utilisateur.
type Role string

const (
	RoleAdmin  Role = "ADMIN"
	RoleClient Role = "CLIENT_DO"
)

// Claims est le contenu signé du jeton.
type Claims struct {
	Subject string `json:"sub"`    // e-mail
	Role    Role   `json:"role"`   // ADMIN | CLIENT_DO
	Tenant  string `json:"tenant"` // identifiant du client (dossier d'archives)
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

var (
	ErrInvalidToken = errors.New("jeton invalide")
	ErrExpiredToken = errors.New("jeton expiré")
)

// TokenManager émet et vérifie des jetons JWT HS256.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewTokenManager exige un secret d'au moins 32 octets.
func NewTokenManager(secret []byte, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET doit contenir au moins 32 caractères")
	}
	if ttl <= 0 {
		ttl = 8 * time.Hour
	}
	return &TokenManager{secret: secret, ttl: ttl, now: time.Now}, nil
}

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

func (m *TokenManager) sign(signingInput string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Issue crée un jeton signé pour l'utilisateur donné.
func (m *TokenManager) Issue(email string, role Role, tenant string) (string, time.Time, error) {
	now := m.now()
	exp := now.Add(m.ttl)
	payload, err := json.Marshal(Claims{
		Subject: email, Role: role, Tenant: tenant,
		Issued: now.Unix(), Expires: exp.Unix(),
	})
	if err != nil {
		return "", time.Time{}, err
	}
	input := jwtHeader + "." + base64.RawURLEncoding.EncodeToString(payload)
	return input + "." + m.sign(input), exp, nil
}

// Verify contrôle la signature et l'expiration. L'algorithme est imposé (HS256) :
// les jetons "alg":"none" ou signés autrement sont refusés.
func (m *TokenManager) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != jwtHeader {
		return nil, ErrInvalidToken
	}
	input := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	want, _ := base64.RawURLEncoding.DecodeString(m.sign(input))
	if !hmac.Equal(sig, want) {
		return nil, ErrInvalidToken
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, ErrInvalidToken
	}
	if c.Subject == "" || (c.Role != RoleAdmin && c.Role != RoleClient) {
		return nil, ErrInvalidToken
	}
	if m.now().Unix() >= c.Expires {
		return nil, ErrExpiredToken
	}
	return &c, nil
}
