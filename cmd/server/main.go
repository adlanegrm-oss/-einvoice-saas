package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"einvoice-saas/internal/app"
	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/middleware"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/repository"
	"einvoice-saas/internal/repository/postgres"
	"einvoice-saas/internal/security"
	"einvoice-saas/internal/service"
	"einvoice-saas/internal/validator"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	schematronEngine := validator.NewSchematronEngine(nil)
	normativeValidator := validator.NewNormativeValidator(true)
	frFiscalValidator := fr.NewFranceCanonicalValidator()

	valFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		resp, err := app.ExecuteValidationPipeline(xmlData, profile, schematronEngine, normativeValidator, frFiscalValidator)
		if err != nil {
			return "", nil, false, resp, err
		}
		return resp.Syntax, resp.CanonicalInvoice, resp.Valid, resp, nil
	}

	var invRepo repository.InvoiceRepository
	var evtRepo repository.EventRepository
	var idemRepo repository.IdempotencyRepository
	var db *sql.DB

	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("ENVIRONMENT"))
	}
	isProduction := env == "production" || env == "prod"

	dbURL := os.Getenv("DATABASE_URL")
	if isProduction && dbURL == "" {
		log.Fatalf("[FATAL] Demarrage impossible en production : DATABASE_URL est obligatoire")
	}

	if dbURL != "" {
		var err error
		db, err = postgres.OpenDB(dbURL)
		if err != nil {
			log.Fatalf("PostgreSQL connection failed: %v", err)
		}

		if err := postgres.ApplyMigrations(context.Background(), db); err != nil {
			log.Fatalf("Schema migration failed: %v", err)
		}
		log.Printf("[PERSISTENCE] Migrations appliquees avec succes")

		invRepo = postgres.NewInvoiceRepo(db)
		evtRepo = postgres.NewEventRepo(db)
		idemRepo = postgres.NewIdempotencyRepo(db)
		log.Printf("[PERSISTENCE] PostgreSQL connecte avec succes (env=%s)", env)
	} else {
		invRepo = app.NewInMemInvoiceRepo()
		evtRepo = app.NewInMemEventRepo()
		idemRepo = app.NewInMemIdemRepo()
		log.Printf("[PERSISTENCE] Mode In-Memory actif (dev/test uniquement)")
	}

	invoiceSvc := service.NewInvoiceService(invRepo, evtRepo, idemRepo, valFn)

	var keyStore middleware.APIKeyStore
	if db != nil {
		keyStore = postgres.NewPostgresAPIKeyStore(db)
		log.Printf("[AUTH] API key store PostgreSQL actif")
	} else {
		keyStore = security.NewInMemoryKeyStore()
		log.Printf("[AUTH] API key store In-Memory actif (dev/test uniquement)")
	}

	handler := app.SetupRouter(keyStore, invoiceSvc)

	log.Printf("[READY] E-Invoicing Gateway demarree sur le port %s", port)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}
