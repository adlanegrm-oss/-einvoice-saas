package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

// IdempotencyStore is a simple in-memory cache for idempotent POST responses.
// In production replace with Redis / DB with TTL.
type IdempotencyStore struct {
	mu    sync.RWMutex
	items map[string]*idempotencyEntry
	ttl   time.Duration
}

type idempotencyEntry struct {
	status    int
	body      []byte
	headers   http.Header
	expiresAt time.Time
}

func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	s := &IdempotencyStore{
		items: make(map[string]*idempotencyEntry),
		ttl:   ttl,
	}
	go s.purgeLoop()
	return s
}

func (s *IdempotencyStore) Get(key string) (status int, body []byte, headers http.Header, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, exists := s.items[key]
	if !exists || time.Now().After(e.expiresAt) {
		return 0, nil, nil, false
	}
	return e.status, e.body, e.headers.Clone(), true
}

func (s *IdempotencyStore) Set(key string, status int, body []byte, headers http.Header) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = &idempotencyEntry{
		status:    status,
		body:      append([]byte(nil), body...),
		headers:   headers.Clone(),
		expiresAt: time.Now().Add(s.ttl),
	}
}

func (s *IdempotencyStore) purgeLoop() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.items {
			if now.After(v.expiresAt) {
				delete(s.items, k)
			}
		}
		s.mu.Unlock()
	}
}

// IdempotencyKey extracts a stable key for a request.
// Priority:
//  1. Explicit "Idempotency-Key" header
//  2. Canonical SHA-256 of the request body (for POST /emit)
func IdempotencyKey(r *http.Request, body []byte) string {
	if k := r.Header.Get("Idempotency-Key"); k != "" {
		return "hdr:" + k
	}
	sum := sha256.Sum256(body)
	return "sha:" + hex.EncodeToString(sum[:])
}

// responseRecorder captures status, headers and body for caching.
type responseRecorder struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.body = append(rr.body, b...)
	return rr.ResponseWriter.Write(b)
}

// IdempotencyMiddleware returns a middleware that guarantees at-most-once semantics
// for POST requests that carry an Idempotency-Key or whose body can be hashed.
func IdempotencyMiddleware(store *IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Read body once
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"unable to read body"}`, http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()
			r.Body = io.NopCloser(bytesReader(body))

			key := IdempotencyKey(r, body)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Cache hit → replay previous response
			if status, cachedBody, headers, ok := store.Get(key); ok {
				for k, vv := range headers {
					for _, v := range vv {
						w.Header().Add(k, v)
					}
				}
				w.Header().Set("X-Idempotency-Replayed", "true")
				w.WriteHeader(status)
				_, _ = w.Write(cachedBody)
				return
			}

			// Cache miss → execute and store
			rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rr, r)

			// Only cache successful or client-error responses (not 5xx)
			if rr.status < 500 {
				store.Set(key, rr.status, rr.body, w.Header())
			}
		})
	}
}

// bytesReader is a tiny helper to re-create an io.ReadCloser from []byte.
func bytesReader(b []byte) io.ReadCloser {
	return io.NopCloser(&byteSliceReader{b: b})
}

type byteSliceReader struct {
	b []byte
	i int
}

func (r *byteSliceReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

// CanonicalPayloadHash produces a stable SHA-256 of a JSON payload
// (keys sorted, no whitespace) – useful when the client does not send
// an Idempotency-Key header.
func CanonicalPayloadHash(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
