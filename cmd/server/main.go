package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
	"github.com/adlanegrm-oss/einvoice-saas/internal/config"
	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/logger"
	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
)

// app assemble toutes les dépendances ; séparé de main() pour être testable.
type app struct {
	cfg     *config.Config
	db      *sql.DB
	pool    *worker.Pool
	store   *auth.Store
	archive *handler.ArchiveHandler
	Handler http.Handler
}

func (a *app) Close() {
	a.pool.Stop()
	a.db.Close()
}

// disabledNotifier : hors DEV sans SMTP, on n'écrit jamais le lien dans les journaux.
type disabledNotifier struct{}

func (disabledNotifier) SendResetLink(email, link string) error {
	return errors.New("SMTP non configuré : lien de réinitialisation non envoyé")
}

func newApp(cfg *config.Config, notifier handler.ResetNotifier) (*app, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return nil, fmt.Errorf("dossier de données : %w", err)
	}

	// --- base de données (SQLite, une seule connexion : pas de verrous concurrents) ---
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base : %w", err)
	}
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA busy_timeout=5000", "PRAGMA foreign_keys=ON"} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("%s : %w", pragma, err)
		}
	}
	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	// --- authentification ---
	secret := cfg.JWTSecret
	if secret == "" { // DEV uniquement (config.Load l'interdit ailleurs)
		if secret, err = auth.RandomSecret(32); err != nil {
			db.Close()
			return nil, err
		}
		slog.Warn("JWT_SECRET absent : secret temporaire généré (les sessions sont perdues à chaque redémarrage)")
	}
	tokens, err := auth.NewTokenManager([]byte(secret), cfg.TokenTTL)
	if err != nil {
		db.Close()
		return nil, err
	}

	store, err := auth.NewStore()
	if err != nil {
		db.Close()
		return nil, err
	}
	adminPassword := cfg.AdminPassword
	if adminPassword == "" { // DEV uniquement
		if adminPassword, err = auth.RandomSecret(12); err != nil {
			db.Close()
			return nil, err
		}
		cfg.GeneratedAdminPassword = adminPassword
	}
	if _, err := store.AddUser(cfg.AdminEmail, adminPassword, auth.RoleAdmin); err != nil {
		db.Close()
		return nil, fmt.Errorf("compte administrateur : %w", err)
	}
	if cfg.ClientEmail != "" {
		if _, err := store.AddUser(cfg.ClientEmail, cfg.ClientPassword, auth.RoleClient); err != nil {
			db.Close()
			return nil, fmt.Errorf("compte client : %w", err)
		}
	}

	if notifier == nil {
		switch {
		case cfg.SMTP.Enabled():
			notifier = handler.SMTPNotifier{Host: cfg.SMTP.Host, Port: cfg.SMTP.Port, User: cfg.SMTP.User, Pass: cfg.SMTP.Pass, From: cfg.SMTP.From}
		case !cfg.IsProdLike():
			notifier = handler.LogNotifier{}
		default:
			slog.Warn("SMTP_HOST / SMTP_FROM absents : la réinitialisation de mot de passe est inopérante")
			notifier = disabledNotifier{}
		}
	}

	// --- handlers ---
	archive, err := handler.NewArchiveHandler(cfg.ArchiveDir)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("dossier d'archives : %w", err)
	}
	pool := worker.NewPool(2, 16)
	invH := handler.NewInvoiceHandler(repo, pool)
	authH := handler.NewAuthHandler(store, tokens, notifier, cfg.PublicBaseURL)
	valH := handler.NewValidationHandler()

	loginLimit := middleware.RateLimit(auth.NewLimiter(20, time.Minute))
	forgotLimit := middleware.RateLimit(auth.NewLimiter(5, 15*time.Minute))
	resetLimit := middleware.RateLimit(auth.NewLimiter(10, 15*time.Minute))
	both := []auth.Role{auth.RoleAdmin, auth.RoleClient}
	protect := func(h http.HandlerFunc, roles ...auth.Role) http.Handler {
		return middleware.Protect(tokens, h, roles...)
	}

	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("GET /health", invH.Health)
	mux.HandleFunc("GET /api/v1/health", invH.Health)
	mux.Handle("POST /api/v1/auth/login", loginLimit(http.HandlerFunc(authH.Login)))
	mux.Handle("POST /api/v1/auth/forgot-password", forgotLimit(http.HandlerFunc(authH.ForgotPassword)))
	mux.Handle("POST /api/v1/auth/reset-password", resetLimit(http.HandlerFunc(authH.ResetPassword)))

	// Authentifié (client ou administrateur)
	mux.Handle("GET /api/v1/auth/me", protect(authH.Me, both...))
	mux.Handle("POST /api/v1/validate", protect(valH.HandleValidateDocument, both...))
	mux.Handle("POST /api/v1/invoices", protect(invH.Validate, both...))
	mux.Handle("GET /api/v1/invoices", protect(invH.List, both...))
	mux.Handle("GET /api/v1/invoices/export", protect(invH.ExportXML, both...))
	mux.Handle("GET /api/v1/invoices/list", protect(archive.List, both...))
	mux.Handle("POST /api/v1/invoices/deposit", protect(archive.Deposit, both...))
	mux.Handle("GET /api/v1/invoices/download", protect(archive.Download, both...))
	mux.Handle("GET /api/v1/reports/daily", protect(invH.GetDailyReport, both...))

	// Administrateur uniquement
	mux.Handle("POST /api/v1/jobs/daily-report", protect(invH.TriggerAsyncCronTask, auth.RoleAdmin))

	// Pages statiques (l'accès aux données reste contrôlé par l'API)
	mux.Handle("GET /", http.FileServer(http.Dir(cfg.WebDir)))

	return &app{
		cfg: cfg, db: db, pool: pool, store: store, archive: archive,
		Handler: middleware.Recover(middleware.SecurityHeaders(mux)),
	}, nil
}

func main() {
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration invalide", "error", err)
		os.Exit(1)
	}

	a, err := newApp(cfg, nil)
	if err != nil {
		slog.Error("démarrage impossible", "error", err)
		os.Exit(1)
	}
	defer a.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go a.archive.RunPurger(ctx)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           a.Handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute, // dépôts de fichiers volumineux
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    64 << 10,
	}

	slog.Info("serveur eInvoice SaaS démarré", "env", cfg.AppEnv, "addr", srv.Addr, "db", cfg.DBPath, "archives", cfg.ArchiveDir)
	if cfg.GeneratedAdminPassword != "" {
		slog.Warn("[DEV] compte administrateur généré", "email", cfg.AdminEmail, "password", cfg.GeneratedAdminPassword)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("erreur serveur", "error", err)
			a.Close()
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("arrêt demandé")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("arrêt forcé", "error", err)
	}
}
