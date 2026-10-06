package middleware

import (
"context"
"crypto/sha256"
"encoding/hex"
"errors"
"net/http"
"strings"
)

type contextKey string

const (
TenantIDContextKey contextKey = "tenant_id"
KeyIDContextKey    contextKey = "api_key_id"
)

type TenantRecord struct {
ID      string
KeyID   string
Active  bool
KeyHash string
}

type APIKeyStore interface {
FindTenantByKeyHash(ctx context.Context, hash string) (*TenantRecord, error)
}

func RequireAPIKey(store APIKeyStore) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
http.Error(w, `{"error":"unauthorized","message":"missing Authorization header"}`, http.StatusUnauthorized)
return
}

parts := strings.SplitN(authHeader, " ", 2)
if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
http.Error(w, `{"error":"unauthorized","message":"invalid authorization format, expected Bearer <token>"}`, http.StatusUnauthorized)
return
}

rawToken := strings.TrimSpace(parts[1])
if len(rawToken) < 32 {
http.Error(w, `{"error":"unauthorized","message":"invalid api key length"}`, http.StatusUnauthorized)
return
}

hashBytes := sha256.Sum256([]byte(rawToken))
keyHash := hex.EncodeToString(hashBytes[:])

tenant, err := store.FindTenantByKeyHash(r.Context(), keyHash)
if err != nil || tenant == nil || !tenant.Active {
http.Error(w, `{"error":"unauthorized","message":"invalid or inactive api key"}`, http.StatusUnauthorized)
return
}

ctx := context.WithValue(r.Context(), TenantIDContextKey, tenant.ID)
ctx = context.WithValue(ctx, KeyIDContextKey, tenant.KeyID)

next.ServeHTTP(w, r.WithContext(ctx))
})
}
}

func GetTenantID(ctx context.Context) (string, error) {
val := ctx.Value(TenantIDContextKey)
if val == nil {
return "", errors.New("security fault: missing tenant in execution context")
}
tID, ok := val.(string)
if !ok || tID == "" {
return "", errors.New("security fault: invalid tenant format in context")
}
return tID, nil
}
