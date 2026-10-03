package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/service"
)

type PipelineHandler struct {
	pipeline *service.InvoicePipeline
	repo     *repository.SQLiteInvoiceRepository
}

func NewPipelineHandler(p *service.InvoicePipeline, r *repository.SQLiteInvoiceRepository) *PipelineHandler {
	return &PipelineHandler{
		pipeline: p,
		repo:     r,
	}
}

// EmitInvoice reçoit une facture structurée et lance la pipeline de validation et de routage
func (h *PipelineHandler) EmitInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	owner := r.Header.Get("X-Tenant-ID")
	if owner == "" {
		owner = "default-tenant"
	}

	var inv invoice.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		http.Error(w, "Format JSON invalide : "+err.Error(), http.StatusBadRequest)
		return
	}

	res, err := h.pipeline.ProcessAndEmit(r.Context(), owner, inv)
	if err != nil {
		http.Error(w, "Erreur de traitement : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if res.Status == status.StateRejected {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(res)
}

// GetAuditTrail renvoie l'historique immuable des statuts (Piste d'Audit Fiable) d'une facture
func (h *PipelineHandler) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Analyse du chemin : /api/v1/invoices/{id}/audit-trail
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/invoices/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "audit-trail" || parts[0] == "" {
		http.Error(w, "Chemin invalide. Format attendu : /api/v1/invoices/{id}/audit-trail", http.StatusBadRequest)
		return
	}
	invoiceID := parts[0]

	history, err := h.repo.GetStatusHistory(r.Context(), invoiceID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de l'historique : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(history)
}