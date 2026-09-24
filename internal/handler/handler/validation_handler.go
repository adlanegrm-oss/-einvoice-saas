package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/validator"
)

type ValidationHandler struct {
	fileValidator *validator.FileValidator
}

func NewValidationHandler() *ValidationHandler {
	return &ValidationHandler{
		fileValidator: validator.NewFileValidator(),
	}
}

// HandleValidateDocument gère l'appel HTTP pour valider tout type de fichier e-invoicing
func (h *ValidationHandler) HandleValidateDocument(w http.ResponseWriter, r *http.Request) {
	// Restreindre aux requêtes POST
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée (utiliser POST)", http.StatusMethodNotAllowed)
		return
	}

	// Limiter la taille du corps de la requête à 10 Mo pour la sécurité
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du fichier", http.StatusBadRequest)
		return
	}

	// Appel du validateur unifié
	result := h.fileValidator.ValidateFileContent(content)

	// Retourner le résultat au format JSON
	w.Header().Set("Content-Type", "application/json")
	if !result.IsValid {
		w.WriteHeader(http.StatusUnprocessableEntity) // Code 422 si le document contient des erreurs
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(result)
}
