package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/logger"
)

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginSecurityMiddleware(limiter *LoginRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ExtractIP(r)

			// 1. Contrôle Rate-Limit IP
			allowedIP, remainingIP, retryIP := limiter.CheckIP(ip)
			w.Header().Set("X-RateLimit-Limit", "10")
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remainingIP))

			if !allowedIP {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryIP.Seconds())))
				slog.Warn("rate limit login dépassé (IP)", "ip", ip, "retry_after_sec", int(retryIP.Seconds()))
				http.Error(w, `{"error":"trop de tentatives, veuillez patienter"}`, http.StatusTooManyRequests)
				return
			}

			// 2. Lecture du body pour inspecter le compte sans fuite de mot de passe
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"corps invalide"}`, http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			var payload loginPayload
			_ = json.Unmarshal(bodyBytes, &payload)

			// 3. Contrôle Rate-Limit Compte
			if payload.Email != "" {
				allowedAcc, _, retryAcc := limiter.CheckAccount(payload.Email)
				if !allowedAcc {
					w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAcc.Seconds())))
					slog.Warn("rate limit login dépassé (Compte)", "email", payload.Email, "ip", ip)
					http.Error(w, `{"error":"compte temporairement bloqué pour brute-force"}`, http.StatusTooManyRequests)
					return
				}
			}

			// 4. Journalisation assainie (ZÉRO secret dans les logs)
			sanitizedBody := logger.SanitizePayload(bodyBytes)
			slog.Info("tentative de connexion reçue", "ip", ip, "email", payload.Email, "payload_sanitized", sanitizedBody)

			next.ServeHTTP(w, r)
		})
	}
}