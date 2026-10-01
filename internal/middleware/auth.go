// Package middleware contient l'authentification par jeton signé, l'autorisation
// par rôle, la limitation de débit et les en-têtes de sécurité.
package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
)

type ctxKey struct{}

// ClaimsFrom renvoie l'identité authentifiée placée dans le contexte par Authenticate.
func ClaimsFrom(ctx context.Context) (*auth.Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(*auth.Claims)
	return c, ok && c != nil
}

// WithClaims est utile pour les tests unitaires de handlers.
func WithClaims(ctx context.Context, c *auth.Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "bearer "
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}

// Authenticate exige un jeton "Authorization: Bearer ..." valide.
// Aucun en-tête fourni par le client (comme X-User-Role) n'est jamais cru.
func Authenticate(tm *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearer(r)
			if token == "" {
				writeErr(w, http.StatusUnauthorized, "Authentification requise")
				return
			}
			claims, err := tm.Verify(token)
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "Jeton invalide ou expiré")
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

// RequireRole autorise uniquement les rôles listés (à placer après Authenticate).
func RequireRole(roles ...auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := ClaimsFrom(r.Context())
			if !ok {
				writeErr(w, http.StatusUnauthorized, "Authentification requise")
				return
			}
			for _, allowed := range roles {
				if c.Role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeErr(w, http.StatusForbidden, "Accès refusé")
		})
	}
}

// Protect combine Authenticate et RequireRole autour d'un handler.
func Protect(tm *auth.TokenManager, next http.Handler, roles ...auth.Role) http.Handler {
	return Authenticate(tm)(RequireRole(roles...)(next))
}

// ProtectFunc est la variante pour http.HandlerFunc.
func ProtectFunc(tm *auth.TokenManager, next http.HandlerFunc, roles ...auth.Role) http.Handler {
	return Protect(tm, next, roles...)
}

// ClientIP renvoie l'adresse distante directe. X-Forwarded-For n'est volontairement
// pas utilisé : il est falsifiable sans reverse proxy de confiance.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimit répond 429 quand la clé (adresse IP) dépasse le quota du limiteur.
func RateLimit(l *auth.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow(ClientIP(r)) {
				w.Header().Set("Retry-After", "60")
				writeErr(w, http.StatusTooManyRequests, "Trop de tentatives, réessayez plus tard")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders ajoute des en-têtes défensifs et interdit la mise en cache des réponses d'API.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// Recover évite qu'une panique ne coupe la connexion sans réponse et la journalise.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panique dans un handler", "path", r.URL.Path, "panic", rec, "stack", string(debug.Stack()))
				writeErr(w, http.StatusInternalServerError, "Erreur interne")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
