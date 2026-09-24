package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/exporter"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/worker"
)

type InvoiceHandler struct {
	repo       *repository.SQLiteInvoiceRepository
	workerPool *worker.Pool
}

func NewInvoiceHandler(repo *repository.SQLiteInvoiceRepository, pool *worker.Pool) *InvoiceHandler {
	return &InvoiceHandler{
		repo:       repo,
		workerPool: pool,
	}
}

func (h *InvoiceHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status": "ok"}`)
}

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

// GetDailyReport retourne la synthèse financière d'une journée
func (h *InvoiceHandler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	report, err := h.repo.GetDailyReport(dateStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du calcul du rapport"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

// TriggerAsyncCronTask reçoit l'appel du Cron et libère l'exécution immédiatement
func (h *InvoiceHandler) TriggerAsyncCronTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// Soumission de la tâche au Worker Pool en arrière-plan (non-bloquant pour le Cron)
	submitted := h.workerPool.Submit(func() {
		report, err := h.repo.GetDailyReport(dateStr)
		if err != nil {
			fmt.Printf("[Async Job Error] Échec du rapport journalier: %v\n", err)
			return
		}
		fmt.Printf("[Async Job Success] Rapport du %s généré: %d factures, Total TTC: %.2f EUR\n",
			report.Date, report.TotalInvoices, report.TotalTTC)
	})

	if !submitted {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "busy", "message": "File de traitement saturée"})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Traitement de nuit planifié en arrière-plan",
	})
}
