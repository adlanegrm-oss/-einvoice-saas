package config

import (
	"strings"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"APP_ENV", "PORT", "DATA_DIR", "DB_NAME", "JWT_SECRET", "ADMIN_EMAIL",
		"ADMIN_PASSWORD", "CLIENT_EMAIL", "CLIENT_PASSWORD", "TOKEN_TTL", "PUBLIC_BASE_URL", "ARCHIVE_DIR"} {
		t.Setenv(k, "")
	}
}

func TestLoadDevDefaults(t *testing.T) {
	clearEnv(t)
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.AppEnv != EnvDev || c.Port != "8080" || !c.GeneratedJWTSecret {
		t.Errorf("valeurs par défaut inattendues : %+v", c)
	}
}

func TestLoadProdRequiresSecrets(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", "PROD")
	if _, err := Load(); err == nil {
		t.Fatal("PROD sans JWT_SECRET doit échouer")
	}
	t.Setenv("JWT_SECRET", strings.Repeat("k", 32))
	if _, err := Load(); err == nil {
		t.Fatal("PROD sans ADMIN_PASSWORD doit échouer")
	}
	t.Setenv("ADMIN_PASSWORD", "un-mot-de-passe-admin")
	if _, err := Load(); err != nil {
		t.Fatalf("configuration PROD complète refusée : %v", err)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "trop-court")
	if _, err := Load(); err == nil {
		t.Error("secret trop court accepté")
	}
	clearEnv(t)
	t.Setenv("APP_ENV", "STAGING")
	if _, err := Load(); err == nil {
		t.Error("APP_ENV inconnu accepté")
	}
	clearEnv(t)
	t.Setenv("DB_NAME", "../evil.db")
	if _, err := Load(); err == nil {
		t.Error("DB_NAME avec chemin accepté")
	}
	clearEnv(t)
	t.Setenv("CLIENT_EMAIL", "c@example.com")
	if _, err := Load(); err == nil {
		t.Error("CLIENT_EMAIL sans CLIENT_PASSWORD accepté")
	}
}
