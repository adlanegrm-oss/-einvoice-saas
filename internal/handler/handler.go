package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
	"github.com/adlanegrm-oss/einvoice-saas/internal/exporter"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
)

type InvoiceHandler struct {
	repo       *repository.SQLiteInvoiceRepository
	workerPool *worker.Pool
}

func NewInvoiceHandler(repo *repository.SQLiteInvoiceRepository, pool *worker.Pool) *InvoiceHandler {
	return &InvoiceHandler{repo: repo, workerPool: pool}
}

// ownerOf renvoie le filtre de propriétaire : vide pour l'administrateur (voit tout),
// le tenant pour un client. ok=false si la requête n'est pas authentifiée.
func ownerOf(r *http.Request) (owner string, ok bool) {
	c, found := middleware.ClaimsFrom(r.Context())
	if !found {
		return "", false
	}
	if c.Role == auth.RoleAdmin {
		return "", true
	}
	return c.Tenant, true
}

func newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "inv-" + hex.EncodeToString(b)
}

// Health répond à la supervision (route publique) et vérifie la base.
func (h *InvoiceHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.repo.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "down", "database": "unreachable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "connected"})
}

// Validate contrôle puis enregistre une facture structurée (JSON).
func (h *InvoiceHandler) Validate(w http.ResponseWriter, r *http.Request) {
	owner, ok := ownerOf(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var inv invoice.Invoice
	if !decodeJSON(w, r, &inv) {
		return
	}

	if err := inv.Validate(); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"status": "invalid", "error": err.Error()})
		return
	}

	// L'identifiant est attribué par le serveur : un client ne choisit pas la clé primaire.
	inv.ID = newID()
	if inv.IssueDate.IsZero() {
		inv.IssueDate = time.Now().UTC()
	}

	// Un administrateur n'a pas de tenant : ses factures sont rattachées à "admin".
	saveOwner := owner
	if saveOwner == "" {
		saveOwner = "admin"
	}

	switch err := h.repo.SaveFor(saveOwner, inv); {
	case errors.Is(err, repository.ErrDuplicate):
		writeJSON(w, http.StatusConflict, map[string]string{"status": "duplicate", "error": "Ce numéro de facture existe déjà"})
		return
	case err != nil:
		slog.Error("enregistrement de la facture", "error", err)
		writeError(w, http.StatusInternalServerError, "Erreur lors de l'enregistrement")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"status": "valid", "invoice": inv})
}

// List renvoie les factures structurées visibles par l'utilisateur.
func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	owner, ok := ownerOf(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	invoices, err := h.repo.ListFor(owner)
	if err != nil {
		slog.Error("liste des factures", "error", err)
		writeError(w, http.StatusInternalServerError, "Erreur lors de la récupération")
		return
	}
	if invoices == nil {
		invoices = []invoice.Invoice{}
	}
	writeJSON(w, http.StatusOK, invoices)
}

// ExportXML renvoie la facture au format Factur-X (XML CII).
func (h *InvoiceHandler) ExportXML(w http.ResponseWriter, r *http.Request) {
	owner, ok := ownerOf(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Paramètre id manquant")
		return
	}

	target, err := h.repo.GetFor(id, owner)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Facture introuvable")
		return
	}
	if err != nil {
		slog.Error("lecture de la facture", "error", err)
		writeError(w, http.StatusInternalServerError, "Erreur serveur")
		return
	}

	xmlBytes, err := exporter.GenerateFacturXXML(*target)
	if err != nil {
		slog.Error("génération XML", "error", err)
		writeError(w, http.StatusInternalServerError, "Erreur lors de la génération XML")
		return
	}

	// Le numéro sert de nom de fichier : on le nettoie (injection d'en-tête).
	safe := unsafeName.ReplaceAllString(target.Number, "_")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="factur-x-%s.xml"`, strings.Trim(safe, "._- ")))
	w.WriteHeader(http.StatusOK)
	w.Write(xmlBytes)
}

// GetDailyReport retourne la synthèse financière d'une journée (AAAA-MM-JJ).
func (h *InvoiceHandler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	owner, ok := ownerOf(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		writeError(w, http.StatusBadRequest, "Date invalide (format AAAA-MM-JJ attendu)")
		return
	}

	report, err := h.repo.DailyReportFor(owner, dateStr)
	if err != nil {
		slog.Error("rapport journalier", "error", err)
		writeError(w, http.StatusInternalServerError, "Erreur lors du calcul du rapport")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// TriggerAsyncCronTask déclenche le rapport global en arrière-plan (route réservée à l'administrateur).
func (h *InvoiceHandler) TriggerAsyncCronTask(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		writeError(w, http.StatusBadRequest, "Date invalide (format AAAA-MM-JJ attendu)")
		return
	}

	submitted := h.workerPool.Submit(func() {
		report, err := h.repo.GetDailyReport(dateStr)
		if err != nil {
			slog.Error("échec du rapport journalier", "error", err)
			return
		}
		slog.Info("rapport journalier généré", "date", report.Date, "factures", report.TotalInvoices, "total_ttc", report.TotalTTC)
	})

	if !submitted {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "busy", "message": "File de traitement saturée"})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "message": "Traitement planifié en arrière-plan"})
}
