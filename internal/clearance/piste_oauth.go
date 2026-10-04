package clearance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type PISTEOAuthConfig struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
}

type PISTETokenManager struct {
	cfg        PISTEOAuthConfig
	httpClient *http.Client
	mu         sync.Mutex
	cachedTok  string
	expiresAt  time.Time
}

func NewPISTETokenManager(cfg PISTEOAuthConfig) *PISTETokenManager {
	if cfg.TokenURL == "" {
		cfg.TokenURL = "https://oauth.piste.gouv.fr/api/oauth/token"
	}
	return &PISTETokenManager{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// GetToken retourne un bearer token valide avec rafraîchissement automatique
func (m *PISTETokenManager) GetToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Si le token est encore valide avec une marge de sécurité de 60 secondes
	if m.cachedTok != "" && time.Now().Add(60*time.Second).Before(m.expiresAt) {
		return m.cachedTok, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", m.cfg.ClientID)
	data.Set("client_secret", m.cfg.ClientSecret)
	if m.cfg.Scope != "" {
		data.Set("scope", m.cfg.Scope)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("piste oauth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("piste oauth failed with HTTP %d: %s", resp.StatusCode, string(b))
	}

	var tokResp oauthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokResp); err != nil {
		return "", fmt.Errorf("decoding piste oauth token response: %w", err)
	}

	m.cachedTok = tokResp.AccessToken
	m.expiresAt = time.Now().Add(time.Duration(tokResp.ExpiresIn) * time.Second)

	return m.cachedTok, nil
}
