package directory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
)

var (
	ErrParticipantNotFound = errors.New("participant introuvable dans l'annuaire")
	ErrDirectoryAPI        = errors.New("erreur de communication avec l'annuaire centralise")
)

type CachedDirectory struct {
	db         *sql.DB
	httpClient *http.Client
	apiBaseURL string
	apiKey     string
	cacheTTL   time.Duration
}

func NewCachedDirectory(db *sql.DB, apiBaseURL, apiKey string, cacheTTL time.Duration) (*CachedDirectory, error) {
	cd := &CachedDirectory{
		db:         db,
		httpClient: &http.Client{Timeout: 8 * time.Second},
		apiBaseURL: strings.TrimRight(apiBaseURL, "/"),
		apiKey:     apiKey,
		cacheTTL:   cacheTTL,
	}

	if err := cd.initSchema(); err != nil {
		return nil, fmt.Errorf("initialisation schema annuaire : %w", err)
	}

	return cd, nil
}

func (cd *CachedDirectory) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS directory_routing_cache (
		identifier TEXT PRIMARY KEY,
		platform_name TEXT NOT NULL,
		as4_endpoint TEXT NOT NULL,
		certificate_pem TEXT,
		cached_at TIMESTAMP NOT NULL,
		expires_at TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_directory_expires ON directory_routing_cache(expires_at);
	`
	_, err := cd.db.Exec(query)
	return err
}

func (cd *CachedDirectory) Lookup(ctx context.Context, participantID string) (*dispatcher.TargetEndpoint, error) {
	id := strings.TrimSpace(participantID)
	now := time.Now().UTC()

	// 1. Consultation du cache local
	var ep dispatcher.TargetEndpoint
	var expiresAt time.Time
	var cert sql.NullString

	query := `
		SELECT platform_name, as4_endpoint, certificate_pem, expires_at 
		FROM directory_routing_cache 
		WHERE identifier = ?
	`
	err := cd.db.QueryRowContext(ctx, query, id).Scan(
		&ep.PlatformName,
		&ep.AS4Endpoint,
		&cert,
		&expiresAt,
	)

	if err == nil {
		if now.Before(expiresAt) {
			ep.ReceiverID = id
			return &ep, nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("lecture cache annuaire : %w", err)
	}

	// 2. Interrogation de l'annuaire distant (PPF ou SMP PEPPOL)
	remoteEP, err := cd.fetchRemote(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Mise en cache local avec TTL
	exp := now.Add(cd.cacheTTL)
	upsertQuery := `
		INSERT INTO directory_routing_cache (identifier, platform_name, as4_endpoint, cached_at, expires_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(identifier) DO UPDATE SET
			platform_name = excluded.platform_name,
			as4_endpoint = excluded.as4_endpoint,
			cached_at = excluded.cached_at,
			expires_at = excluded.expires_at
	`
	_, _ = cd.db.ExecContext(ctx, upsertQuery, id, remoteEP.PlatformName, remoteEP.AS4Endpoint, now, exp)

	return remoteEP, nil
}

type directoryAPIResponse struct {
	ReceiverID   string `json:"receiver_id"`
	PlatformName string `json:"platform_name"`
	AS4Endpoint  string `json:"as4_endpoint"`
}

func (cd *CachedDirectory) fetchRemote(ctx context.Context, participantID string) (*dispatcher.TargetEndpoint, error) {
	// Fallback de démonstration si aucune URL d'annuaire configurée
	if cd.apiBaseURL == "" {
		return &dispatcher.TargetEndpoint{
			ReceiverID:   participantID,
			PlatformName: "PPF_CHORUS_PRO",
			AS4Endpoint:  "https://edelivery.chorus-pro.gouv.fr/as4",
		}, nil
	}

	url := fmt.Sprintf("%s/api/v1/directory/%s", cd.apiBaseURL, participantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if cd.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+cd.apiKey)
	}

	res, err := cd.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDirectoryAPI, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %s", ErrParticipantNotFound, participantID)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: code HTTP %d", ErrDirectoryAPI, res.StatusCode)
	}

	var data directoryAPIResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decodage reponse annuaire : %w", err)
	}

	return &dispatcher.TargetEndpoint{
		ReceiverID:   participantID,
		PlatformName: data.PlatformName,
		AS4Endpoint:  data.AS4Endpoint,
	}, nil
}