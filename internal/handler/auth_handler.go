package handler

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
)

// AuthHandler expose connexion, mot de passe oublié et réinitialisation.
type AuthHandler struct {
	Store         *auth.Store
	Tokens        *auth.TokenManager
	Notifier      ResetNotifier
	PublicURL     string // base des liens envoyés, ex. https://app.exemple.com
	loginAttempts *auth.Limiter
}

func NewAuthHandler(store *auth.Store, tokens *auth.TokenManager, notifier ResetNotifier, publicURL string) *AuthHandler {
	return &AuthHandler{
		Store: store, Tokens: tokens, Notifier: notifier, PublicURL: publicURL,
		// plafond par adresse e-mail, en plus du limiteur par IP posé sur la route
		loginAttempts: auth.NewLimiter(15, 15*time.Minute),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &creds) {
		return
	}

	// Plafond par compte : limite le "password guessing" distribué sur plusieurs IP.
	if !h.loginAttempts.Allow("login:" + strings.ToLower(strings.TrimSpace(creds.Email))) {
		w.Header().Set("Retry-After", "900")
		writeError(w, http.StatusTooManyRequests, "Trop de tentatives, réessayez plus tard")
		return
	}

	user, err := h.Store.Authenticate(creds.Email, creds.Password)
	if err != nil {
		slog.Warn("échec de connexion", "ip", middleware.ClientIP(r))
		writeError(w, http.StatusUnauthorized, "Identifiant ou mot de passe incorrect.")
		return
	}

	token, exp, err := h.Tokens.Issue(user.Email, user.Role, user.Tenant)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Erreur interne")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "success",
		"token":      token,
		"email":      user.Email,
		"role":       user.Role,
		"expires_at": exp.UTC().Format(time.RFC3339),
	})
}

// Me renvoie l'identité portée par le jeton (sert au contrôle d'accès côté interface).
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	c, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"email": c.Subject, "role": c.Role})
}

const forgotMessage = "Si un compte correspond à cette adresse, un lien de réinitialisation vient d'être envoyé."

// ForgotPassword répond TOUJOURS pareil, que le compte existe ou non, et ne
// renvoie jamais le jeton : il n'est transmis que par le canal de notification.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}

	if token, ok, err := h.Store.IssueResetToken(payload.Email); err != nil {
		slog.Error("émission du jeton de réinitialisation", "error", err)
	} else if ok {
		link := h.PublicURL + "/reset-password.html?token=" + url.QueryEscape(token)
		email, _ := auth.NormalizeEmail(payload.Email)
		if err := h.Notifier.SendResetLink(email, link); err != nil {
			slog.Error("envoi du lien de réinitialisation", "error", err)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success", "message": forgotMessage})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}

	switch err := h.Store.ConsumeResetToken(payload.Token, payload.Password); err {
	case nil:
		writeJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Mot de passe mis à jour avec succès !"})
	case auth.ErrWeakPassword:
		writeError(w, http.StatusBadRequest, err.Error())
	case auth.ErrInvalidResetToken:
		writeError(w, http.StatusBadRequest, "Lien invalide ou expiré. Refaites une demande de réinitialisation.")
	default:
		slog.Error("réinitialisation du mot de passe", "error", err)
		writeError(w, http.StatusInternalServerError, "Erreur interne")
	}
}
