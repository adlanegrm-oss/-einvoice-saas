package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"sync"

	"einvoice-saas/internal/middleware"
)

var (
	ErrKeyNotFound = errors.New("key not found in memory store")
	ErrKeyRevoked  = errors.New("key has been revoked")
	ErrKeyInactive = errors.New("key is inactive")
)

type KeyRecord struct {
	TenantID string
	Active   bool
	Revoked  bool
}

type InMemoryKeyStore struct {
	mu   sync.RWMutex
	keys map[string]KeyRecord
}

func NewInMemoryKeyStore() *InMemoryKeyStore {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "production" || appEnv == "prod" {
		panic("VIOLATION DE SECURITE CRITIQUE: InMemoryKeyStore est strictement interdit en production. Configurez PostgresAPIKeyStore.")
	}
	return &InMemoryKeyStore{
		keys: make(map[string]KeyRecord),
	}
}

func (m *InMemoryKeyStore) AddKey(rawToken string, tenantID string, active bool) {
	hashBytes := sha256.Sum256([]byte(rawToken))
	keyHash := hex.EncodeToString(hashBytes[:])

	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[keyHash] = KeyRecord{
		TenantID: tenantID,
		Active:   active,
		Revoked:  false,
	}
}

func (m *InMemoryKeyStore) RevokeKey(rawToken string) {
	hashBytes := sha256.Sum256([]byte(rawToken))
	keyHash := hex.EncodeToString(hashBytes[:])

	m.mu.Lock()
	defer m.mu.Unlock()
	if rec, exists := m.keys[keyHash]; exists {
		rec.Revoked = true
		m.keys[keyHash] = rec
	}
}

func (m *InMemoryKeyStore) FindTenantByKeyHash(ctx context.Context, hash string) (*middleware.TenantRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, exists := m.keys[hash]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if rec.Revoked {
		return nil, ErrKeyRevoked
	}
	if !rec.Active {
		return nil, ErrKeyInactive
	}

	return &middleware.TenantRecord{
		ID:      rec.TenantID,
		Active:  rec.Active,
		KeyHash: hash,
	}, nil
}
