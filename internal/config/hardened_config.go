package config

import (
	"errors"
	"os"
)

type HardenedConfig struct {
	AppEnv            string
	BehindSecureProxy bool
	AdminEmail        string
	AdminPassword     string
	DatabasePath      string
}

func LoadHardenedConfig() (*HardenedConfig, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	cfg := &HardenedConfig{
		AppEnv:            env,
		BehindSecureProxy: os.Getenv("BEHIND_SECURE_PROXY") == "true",
		AdminEmail:        os.Getenv("ADMIN_EMAIL"),
		AdminPassword:     os.Getenv("ADMIN_PASSWORD"),
		DatabasePath:      os.Getenv("DATABASE_PATH"),
	}

	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "./data/einvoice.db"
	}

	if env == "production" {
		if !cfg.BehindSecureProxy && os.Getenv("TLS_CERT_FILE") == "" {
			return nil, errors.New("FATAL: En production, TLS direct ou BEHIND_SECURE_PROXY=true est obligatoire")
		}
		if cfg.AdminEmail == "" || cfg.AdminEmail == "admin@example.com" {
			return nil, errors.New("FATAL: ADMIN_EMAIL reel obligatoire en production")
		}
		if len(cfg.AdminPassword) < 16 {
			return nil, errors.New("FATAL: ADMIN_PASSWORD doit contenir au moins 16 caracteres en production")
		}
	}

	return cfg, nil
}
