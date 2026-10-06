package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"einvoice-saas/internal/app"
	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/repository"
	"einvoice-saas/internal/repository/postgres"
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
		db, err := postgres.OpenDB(dbURL)
		if err != nil {
			log.Fatalf("PostgreSQL connection failed: %v", err)
		}

		if _, err := os.Stat("migrations"); err == nil {
			if err := postgres.ApplyMigrations(context.Background(), db, "migrations"); err != nil {
				log.Fatalf("Schema migration failed: %v", err)
			}
			log.Printf("[PERSISTENCE] Migrations appliquees avec succes")
		}

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
	keyStore := &app.InMemoryKeyStore{}
	handler := app.SetupRouter(keyStore, invoiceSvc)

	log.Printf("[READY] E-Invoicing Gateway demarree sur le port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
