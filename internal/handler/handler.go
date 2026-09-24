package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/exporter"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
)

// InvoiceHandler gère les requêtes HTTP liées aux factures
type InvoiceHandler struct {
	repo repository.InvoiceRepository
}

// NewInvoiceHandler initialise un nouvel InvoiceHandler
func NewInvoiceHandler(repo repository.InvoiceRepository) *InvoiceHandler {
	return &InvoiceHandler{repo: repo}
}

// Health gère la vérification d'état de l'API
func (h *InvoiceHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "ok"}`)
}

// Validate gère la validation et la sauvegarde d'une facture
func (h *InvoiceHandler) Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Méthode non autorisée"}`, http.StatusMethodNotAllowed)
		return
	}

	var inv invoice.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		http.Error(w, `{"error": "Format JSON invalide"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := inv.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "invalid",
			"error":  err.Error(),
		})
		return
	}

	if err := h.repo.Save(inv); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erreur lors de l'enregistrement",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "valid",
		"invoice": inv,
	})
}

// List retourne toutes les factures enregistrées
func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Méthode non autorisée"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	invoices, err := h.repo.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erreur lors de la récupération",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoices)
}

// ExportXML génère et renvoie le fichier XML Factur-X pour une facture donnée (?id=...)
func (h *InvoiceHandler) ExportXML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Méthode non autorisée"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error": "Paramètre id manquant"}`, http.StatusBadRequest)
		return
	}

	invoices, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}

	var target *invoice.Invoice
	for _, inv := range invoices {
		if inv.ID == id {
			target = &inv
			break
		}
	}

	if target == nil {
		http.Error(w, `{"error": "Facture introuvable"}`, http.StatusNotFound)
		return
	}

	xmlBytes, err := exporter.GenerateFacturXXML(*target)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de la génération XML"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=factur-x-%s.xml", target.Number))
	w.WriteHeader(http.StatusOK)
	w.Write(xmlBytes)
}
