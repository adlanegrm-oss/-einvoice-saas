package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	go func() {
		for range time.Tick(window) {
			rl.mu.Lock()
			now := time.Now()
			for ip, times := range rl.requests {
				var valid []time.Time
				for _, t := range times {
					if now.Sub(t) <= window {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.requests, ip)
				} else {
					rl.requests[ip] = valid
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	times := rl.requests[ip]

	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) <= rl.window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

// SecurityMiddleware applique l'authentification X-API-Key et le rate-limiting
func SecurityMiddleware(next http.Handler) http.Handler {
	limiter := newRateLimiter(60, time.Minute)
	expectedKey := os.Getenv("API_KEY")
	if expectedKey == "" {
		expectedKey = "dev-secret-key-2026"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Liste blanche : Endpoints publics, doc Swagger et ressources web statiques
		isPublic := r.URL.Path == "/" ||
			r.URL.Path == "/health" ||
			r.URL.Path == "/api/v1/health" ||
			r.URL.Path == "/swagger.yaml" ||
			strings.HasSuffix(r.URL.Path, ".html") ||
			strings.HasSuffix(r.URL.Path, ".js") ||
			strings.HasSuffix(r.URL.Path, ".css") ||
			strings.HasSuffix(r.URL.Path, ".json") ||
			strings.HasSuffix(r.URL.Path, ".png") ||
			strings.HasSuffix(r.URL.Path, ".ico")

		// 2. Headers de durcissement HTTP (OWASP)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Laisser passer les ressources web sans bloquer l'interface
		if isPublic {
			w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline';")
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Security-Policy", "default-src 'none'")

		// 3. Rate Limiting par IP sur les routes protégées
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		if !limiter.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too many requests. Rate limit exceeded."}`))
			return
		}

		// 4. Contrôle de la clé API pour les routes métier
		clientKey := r.Header.Get("X-API-Key")
		if clientKey == "" || clientKey != expectedKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Non autorisé. En-tête X-API-Key invalide ou manquant."}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}