package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

var jwtTestSecret = []byte("test-secret-0123456789abcdef-test")

const (
	jwtTestIssuer = "einvoice-saas"
	jwtTestAud    = "einvoice-api"
)

func jwtB64(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func jwtSignRaw(secret, headerJSON, payloadJSON []byte) string {
	in := jwtB64(headerJSON) + "." + jwtB64(payloadJSON)

	m := hmac.New(sha256.New, secret)
	m.Write([]byte(in))

	return in + "." + jwtB64(m.Sum(nil))
}

func jwtSign(t *testing.T, secret []byte, claims JWTClaims) string {
	t.Helper()

	p, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}

	return jwtSignRaw(
		secret,
		[]byte(`{"alg":"HS256","typ":"JWT"}`),
		p,
	)
}

func jwtValidClaims() JWTClaims {
	now := time.Now().Unix()

	return JWTClaims{
		Issuer:    jwtTestIssuer,
		Subject:   "user-1",
		Audience:  jwtTestAud,
		TenantID:  "tenant-1",
		Role:      "admin",
		JTI:       "jti-1",
		ExpiresAt: now + 3600,
		NotBefore: now - 60,
		IssuedAt:  now - 60,
	}
}

func TestJWTValidator_ValidToken(t *testing.T) {
	v := NewJWTValidator(
		jwtTestSecret,
		jwtTestIssuer,
		jwtTestAud,
	)

	want := jwtValidClaims()

	got, err := v.ValidateToken(
		jwtSign(t, jwtTestSecret, want),
	)

	if err != nil {
		t.Fatalf("token valide rejete : %v", err)
	}

	if got == nil || *got != want {
		t.Fatalf("claims inattendus : %+v", got)
	}
}

func TestJWTValidator_EmptyExpectations_SkipIssuerAndAudience(t *testing.T) {
	v := NewJWTValidator(
		jwtTestSecret,
		"",
		"",
	)

	c := jwtValidClaims()
	c.Issuer = "autre"
	c.Audience = "autre"

	if _, err := v.ValidateToken(
		jwtSign(t, jwtTestSecret, c),
	); err != nil {
		t.Fatalf(
			"iss/aud ne devraient pas etre controles : %v",
			err,
		)
	}
}

func TestJWTValidator_NotBeforeZero_IsAccepted(t *testing.T) {
	v := NewJWTValidator(
		jwtTestSecret,
		jwtTestIssuer,
		jwtTestAud,
	)

	c := jwtValidClaims()
	c.NotBefore = 0

	if _, err := v.ValidateToken(
		jwtSign(t, jwtTestSecret, c),
	); err != nil {
		t.Fatalf(
			"nbf=0 devrait etre accepte : %v",
			err,
		)
	}
}

func TestJWTValidator_Rejects(t *testing.T) {
	v := NewJWTValidator(
		jwtTestSecret,
		jwtTestIssuer,
		jwtTestAud,
	)

	valid := jwtSign(
		t,
		jwtTestSecret,
		jwtValidClaims(),
	)

	validParts := strings.Split(valid, ".")

	payload := func(c JWTClaims) []byte {
		b, _ := json.Marshal(c)
		return b
	}

	expired := jwtValidClaims()
	expired.ExpiresAt = time.Now().Unix() - 10

	noExp := jwtValidClaims()
	noExp.ExpiresAt = 0

	future := jwtValidClaims()
	future.NotBefore = time.Now().Unix() + 3600

	badIss := jwtValidClaims()
	badIss.Issuer = "attaquant"

	badAud := jwtValidClaims()
	badAud.Audience = "autre-api"

	tampered := jwtValidClaims()
	tampered.Role = "superadmin"

	tests := []struct {
		name  string
		token string
	}{
		{
			"vide",
			"",
		},
		{
			"deux segments",
			"a.b",
		},
		{
			"quatre segments",
			"a.b.c.d",
		},
		{
			"header non base64",
			"!!!." + validParts[1] + "." + validParts[2],
		},
		{
			"header non JSON",
			jwtSignRaw(
				jwtTestSecret,
				[]byte("pas-du-json"),
				payload(jwtValidClaims()),
			),
		},
		{
			"alg none",
			jwtB64(
				[]byte(`{"alg":"none","typ":"JWT"}`),
			) + "." + validParts[1] + ".",
		},
		{
			"alg HS512",
			jwtSignRaw(
				jwtTestSecret,
				[]byte(`{"alg":"HS512"}`),
				payload(jwtValidClaims()),
			),
		},
		{
			"alg RS256",
			jwtSignRaw(
				jwtTestSecret,
				[]byte(`{"alg":"RS256"}`),
				payload(jwtValidClaims()),
			),
		},
		{
			"mauvais secret",
			jwtSign(
				t,
				[]byte("autre-secret-autre-secret-autre!"),
				jwtValidClaims(),
			),
		},
		{
			"signature non base64",
			validParts[0] + "." + validParts[1] + ".!!!",
		},
		{
			"payload altere",
			validParts[0] + "." +
				jwtB64(payload(tampered)) +
				"." + validParts[2],
		},
		{
			"payload non JSON",
			jwtSignRaw(
				jwtTestSecret,
				[]byte(`{"alg":"HS256"}`),
				[]byte("pas-du-json"),
			),
		},
		{
			"expire",
			jwtSign(t, jwtTestSecret, expired),
		},
		{
			"exp absent",
			jwtSign(t, jwtTestSecret, noExp),
		},
		{
			"nbf dans le futur",
			jwtSign(t, jwtTestSecret, future),
		},
		{
			"mauvais issuer",
			jwtSign(t, jwtTestSecret, badIss),
		},
		{
			"mauvaise audience",
			jwtSign(t, jwtTestSecret, badAud),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := v.ValidateToken(tc.token)

			if err == nil {
				t.Fatalf(
					"une erreur etait attendue (claims=%+v)",
					claims,
				)
			}

			if claims != nil {
				t.Fatalf(
					"claims doit etre nil en cas d'erreur, obtenu %+v",
					claims,
				)
			}
		})
	}
}
