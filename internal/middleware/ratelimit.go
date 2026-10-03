package middleware

import (
	"context"
"net/http"
"sync"
"time"
)

type clientBucket struct {
tokens     float64
lastRefill time.Time
}

type LoginRateLimiter struct {
mu            sync.Mutex
ipBuckets     map[string]*clientBucket
userBuckets   map[string]*clientBucket
tenantBuckets map[string]*clientBucket
rate          float64
burst         float64
cleanupTick   time.Duration
}

func NewLoginRateLimiter() *LoginRateLimiter {
limiter := &LoginRateLimiter{
ipBuckets:     make(map[string]*clientBucket),
userBuckets:   make(map[string]*clientBucket),
tenantBuckets: make(map[string]*clientBucket),
rate:          10.0 / 60.0,
burst:         10.0,
cleanupTick:   5 * time.Minute,
}
go limiter.purgeInactive()
return limiter
}

func ExtractIP(r *http.Request) string {
return ClientIP(r)
}

func (l *LoginRateLimiter) allow(buckets map[string]*clientBucket, key string, burst float64) (bool, int, time.Duration) {
l.mu.Lock()
defer l.mu.Unlock()

now := time.Now()
b, ok := buckets[key]
if !ok {
buckets[key] = &clientBucket{
tokens:     burst - 1,
lastRefill: now,
}
return true, int(burst - 1), 0
}

elapsed := now.Sub(b.lastRefill).Seconds()
b.tokens += elapsed * l.rate
if b.tokens > burst {
b.tokens = burst
}
b.lastRefill = now

if b.tokens < 1.0 {
wait := time.Duration((1.0 - b.tokens) / l.rate * float64(time.Second))
return false, 0, wait
}

b.tokens -= 1.0
return true, int(b.tokens), 0
}

func (l *LoginRateLimiter) AllowIP(ip string) (bool, int, time.Duration) {
return l.allow(l.ipBuckets, ip, l.burst)
}

func (l *LoginRateLimiter) AllowUser(email string) (bool, int, time.Duration) {
return l.allow(l.userBuckets, email, 5.0)
}

func (l *LoginRateLimiter) CheckIP(ip string) (bool, int, time.Duration) {
return l.AllowIP(ip)
}

func (l *LoginRateLimiter) CheckAccount(email string) (bool, int, time.Duration) {
return l.AllowUser(email)
}

func (l *LoginRateLimiter) AllowTenant(tenantID string) (bool, int, time.Duration) {
return l.allow(l.tenantBuckets, tenantID, 100.0)
}

func (l *LoginRateLimiter) TenantRateLimit() func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
key := ""
isTenant := false

if tenant, ok := TenantFrom(r.Context()); ok && tenant != "" {
key = tenant
isTenant = true
} else {
key = ExtractIP(r)
}

allowed := false
if isTenant {
allowed, _, _ = l.AllowTenant(key)
} else {
allowed, _, _ = l.AllowIP(key)
}

if !allowed {
w.Header().Set("Retry-After", "10")
http.Error(w, `{"error":"quota de requêtes dépassé pour ce tenant"}`, http.StatusTooManyRequests)
return
}

next.ServeHTTP(w, r)
})
}
}

func (l *LoginRateLimiter) purgeInactive() {
ticker := time.NewTicker(l.cleanupTick)
for range ticker.C {
l.mu.Lock()
cutoff := time.Now().Add(-10 * time.Minute)
for k, v := range l.ipBuckets {
if v.lastRefill.Before(cutoff) {
delete(l.ipBuckets, k)
}
}
for k, v := range l.userBuckets {
if v.lastRefill.Before(cutoff) {
delete(l.userBuckets, k)
}
}
for k, v := range l.tenantBuckets {
if v.lastRefill.Before(cutoff) {
delete(l.tenantBuckets, k)
}
}
l.mu.Unlock()
}
}


type tenantContextKey struct{}

func TenantFrom(ctx context.Context) (string, bool) {
	if v, ok := ctx.Value(tenantContextKey{}).(string); ok && v != "" {
		return v, true
	}
	return "", false
}
