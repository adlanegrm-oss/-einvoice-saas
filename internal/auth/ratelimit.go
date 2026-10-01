package auth

import (
	"sync"
	"time"
)

// Limiter est un limiteur à fenêtre fixe, en mémoire, par clé (ex. adresse IP).
type Limiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	buckets map[string]*bucket
	now     func() time.Time
	calls   int
}

type bucket struct {
	count int
	reset time.Time
}

func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{max: max, window: window, buckets: make(map[string]*bucket), now: time.Now}
}

// Allow enregistre une tentative et indique si elle est autorisée.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()

	// nettoyage périodique pour borner la mémoire
	l.calls++
	if l.calls%256 == 0 {
		for k, b := range l.buckets {
			if now.After(b.reset) {
				delete(l.buckets, k)
			}
		}
	}

	b := l.buckets[key]
	if b == nil || now.After(b.reset) {
		l.buckets[key] = &bucket{count: 1, reset: now.Add(l.window)}
		return true
	}
	if b.count >= l.max {
		return false
	}
	b.count++
	return true
}
