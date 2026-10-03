package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/service"
)

type PipelineHandler struct {
	pipeline *service.InvoicePipeline
	repo     *repository.SQLiteInvoiceRepository
}

func NewPipelineHandler(p *service.InvoicePipeline, repo *repository.SQLiteInvoiceRepository) *PipelineHandler {
	return &PipelineHandler{
		pipeline: p,
		repo:     repo,
	}
}

type EmitInvoiceRequest struct {
	Invoice invoice.Invoice `json:"invoice"`
}

type StatusUpdateRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func (h *PipelineHandler) EmitInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Méthode non autorisée"}`, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"Lecture corps de requête impossible"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var targetInv invoice.Invoice
	// 1. Essai de désérialisation directe de l'objet facture (comme envoyé dans les tests)
	if err := json.Unmarshal(body, &targetInv); err != nil || targetInv.Number == "" {
		// 2. Si non concluant, tentative avec l'enveloppe {"invoice": {...}}
		var wrapped EmitInvoiceRequest
		if errWrap := json.Unmarshal(body, &wrapped); errWrap == nil && wrapped.Invoice.Number != "" {
			targetInv = wrapped.Invoice
		}
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default-tenant"
	}

	res, err := h.pipeline.ProcessAndEmit(r.Context(), tenantID, targetInv)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
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

func (h *PipelineHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"Méthode non autorisée"}`, http.StatusMethodNotAllowed)
		return
	}

	// Découpage attendu : /api/v1/invoices/{id}/status
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 || parts[4] != "status" {
		http.Error(w, `{"error":"URL invalide"}`, http.StatusBadRequest)
		return
	}
	invoiceID := parts[3]

	var req StatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Corps JSON invalide"}`, http.StatusBadRequest)
		return
	}

	actor := r.Header.Get("X-User-Email")
	if actor == "" {
		actor = "PDP_SYSTEM"
	}
	if err := h.pipeline.TransitionInvoiceStatus(r.Context(), invoiceID, status.InvoiceState(req.Status), actor, req.Reason); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"invoice_id": invoiceID,
		"new_status": req.Status,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PipelineHandler) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Méthode non autorisée"}`, http.StatusMethodNotAllowed)
		return
	}

	// Découpage attendu : /api/v1/invoices/{id}/audit-trail
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 || parts[4] != "audit-trail" {
		http.Error(w, `{"error":"URL invalide"}`, http.StatusBadRequest)
		return
	}
	invoiceID := parts[3]

	history, err := h.repo.GetStatusHistory(r.Context(), invoiceID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Erreur lecture audit trail : %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if len(history) == 0 {
		http.Error(w, `{"error":"Aucun historique trouvé pour cette facture"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(history)
}
