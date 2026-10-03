package middleware

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/adlanegrm-oss/einvoice-saas/internal/logger"
)

func TestRateLimiter_TokenBucket_IP(t *testing.T) {
	limiter := NewLoginRateLimiter()
	ip := "192.168.1.100"

	// 10 premières tentatives doivent passer
	for i := 0; i < 10; i++ {
		allowed, rem, _ := limiter.CheckIP(ip)
		if !allowed {
			t.Fatalf("requête %d aurait dû être autorisée (rem=%d)", i+1, rem)
		}
	}

	// La 11e doit être rejetée immédiatement
	allowed, _, retryAfter := limiter.CheckIP(ip)
	if allowed {
		t.Fatalf("la 11e requête aurait dû être bloquée (429)")
	}
	if retryAfter <= 0 {
		t.Errorf("Retry-After doit être supérieur à 0, reçu %v", retryAfter)
	}
}

func TestRateLimiter_AccountProtection(t *testing.T) {
	limiter := NewLoginRateLimiter()
	email := "target@entreprise.fr"

	// Seuil à 5 tentatives pour le compte
	for i := 0; i < 5; i++ {
		allowed, _, _ := limiter.CheckAccount(email)
		if !allowed {
			t.Fatalf("tentative compte %d autorisée attendue", i+1)
		}
	}

	allowed, _, _ := limiter.CheckAccount(email)
	if allowed {
		t.Fatalf("le compte aurait dû être bloqué après 5 échecs rapides")
	}
}

func TestLogSanitizer_MasksSecrets(t *testing.T) {
	rawJSON := []byte(`{"email":"admin@corp.fr","password":"SuperSecretPassword!123","token":"jwt.xyz"}`)
	clean := logger.SanitizePayload(rawJSON)

	if bytes.Contains([]byte(clean), []byte("SuperSecretPassword!123")) {
		t.Errorf("Le mot de passe en clair n'a pas été masqué: %s", clean)
	}
	if !bytes.Contains([]byte(clean), []byte("[REDACTED]")) {
		t.Errorf("Le libellé [REDACTED] est absent: %s", clean)
	}

	// En-tête Authorization
	headers := http.Header{}
	headers.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.secret")
	sanitizedHeaders := logger.SanitizeHeaders(headers)

	if sanitizedHeaders["Authorization"] != "Bearer [REDACTED]" {
		t.Errorf("L'en-tête Authorization n'a pas été correctement masqué: %s", sanitizedHeaders["Authorization"])
	}
}