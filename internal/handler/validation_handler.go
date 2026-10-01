package handler

import (
	"io"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/validator"
)

const maxValidateBody = 10 << 20 // 10 Mo

type ValidationHandler struct {
	fileValidator *validator.FileValidator
}

func NewValidationHandler() *ValidationHandler {
	return &ValidationHandler{fileValidator: validator.NewFileValidator()}
}

// HandleValidateDocument valide le fichier envoyé tel quel dans le corps de la requête (POST).
// Rien n'est enregistré : c'est un simple contrôle.
func (h *ValidationHandler) HandleValidateDocument(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxValidateBody)
	content, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "Fichier illisible ou supérieur à 10 Mo")
		return
	}

	result := h.fileValidator.ValidateFileContent(content)
	if !result.IsValid {
		writeJSON(w, http.StatusUnprocessableEntity, result)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
