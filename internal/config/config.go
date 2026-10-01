// Package config lit la configuration depuis les variables d'environnement.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Environnements reconnus (APP_ENV).
const (
	EnvDev      = "DEV"
	EnvRecette  = "RECETTE"
	EnvPreprod  = "PREPROD"
	EnvProd     = "PROD"
	minSecretLn = 32
)

type Config struct {
	AppEnv        string
	Port          string
	DataDir       string
	DBPath        string
	ArchiveDir    string
	WebDir        string
	PublicBaseURL string

	JWTSecret string
	TokenTTL  time.Duration

	AdminEmail     string
	AdminPassword  string // vide en DEV => généré au démarrage
	ClientEmail    string
	ClientPassword string

	SMTP SMTPConfig

	// GeneratedAdminPassword est renseigné quand le mot de passe admin a été généré (DEV).
	GeneratedAdminPassword string
	// GeneratedJWTSecret est vrai quand le secret a été généré (DEV).
	GeneratedJWTSecret bool
}

type SMTPConfig struct {
	Host, Port, User, Pass, From string
}

func (s SMTPConfig) Enabled() bool { return s.Host != "" && s.From != "" }

// IsProdLike indique un environnement où les valeurs par défaut de développement sont interdites.
func (c *Config) IsProdLike() bool { return c.AppEnv != EnvDev }

// Load lit l'environnement. Hors DEV, JWT_SECRET et ADMIN_PASSWORD sont obligatoires.
func Load() (*Config, error) {
	c := &Config{
		AppEnv:         strings.ToUpper(get("APP_ENV", EnvDev)),
		Port:           strings.TrimPrefix(get("PORT", "8080"), ":"),
		DataDir:        get("DATA_DIR", "data"),
		WebDir:         get("WEB_DIR", "web"),
		PublicBaseURL:  strings.TrimRight(get("PUBLIC_BASE_URL", ""), "/"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AdminEmail:     strings.ToLower(get("ADMIN_EMAIL", "admin@example.com")),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
		ClientEmail:    strings.ToLower(os.Getenv("CLIENT_EMAIL")),
		ClientPassword: os.Getenv("CLIENT_PASSWORD"),
		SMTP: SMTPConfig{
			Host: os.Getenv("SMTP_HOST"), Port: get("SMTP_PORT", "587"),
			User: os.Getenv("SMTP_USER"), Pass: os.Getenv("SMTP_PASS"), From: os.Getenv("SMTP_FROM"),
		},
	}

	switch c.AppEnv {
	case EnvDev, EnvRecette, EnvPreprod, EnvProd:
	default:
		return nil, fmt.Errorf("APP_ENV inconnu : %q (DEV, RECETTE, PREPROD ou PROD)", c.AppEnv)
	}

	if _, err := strconv.Atoi(c.Port); err != nil {
		return nil, fmt.Errorf("PORT invalide : %q", c.Port)
	}

	ttl, err := time.ParseDuration(get("TOKEN_TTL", "8h"))
	if err != nil || ttl <= 0 {
		return nil, fmt.Errorf("TOKEN_TTL invalide (exemple : 8h) : %v", err)
	}
	c.TokenTTL = ttl

	dbName := get("DB_NAME", "invoices.db")
	if filepath.Base(dbName) != dbName {
		return nil, errors.New("DB_NAME doit être un nom de fichier, sans chemin")
	}
	c.DBPath = filepath.Join(c.DataDir, dbName)
	c.ArchiveDir = get("ARCHIVE_DIR", filepath.Join(c.DataDir, "archives"))

	if c.PublicBaseURL == "" {
		c.PublicBaseURL = "http://localhost:" + c.Port
	}

	if c.JWTSecret == "" {
		if c.IsProdLike() {
			return nil, errors.New("JWT_SECRET est obligatoire hors DEV (32 caractères minimum)")
		}
		c.GeneratedJWTSecret = true
	} else if len(c.JWTSecret) < minSecretLn {
		return nil, fmt.Errorf("JWT_SECRET doit contenir au moins %d caractères", minSecretLn)
	}

	if c.AdminPassword == "" && c.IsProdLike() {
		return nil, errors.New("ADMIN_PASSWORD est obligatoire hors DEV (10 caractères minimum)")
	}
	if (c.ClientEmail == "") != (c.ClientPassword == "") {
		return nil, errors.New("CLIENT_EMAIL et CLIENT_PASSWORD vont ensemble")
	}
	return c, nil
}

func get(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
