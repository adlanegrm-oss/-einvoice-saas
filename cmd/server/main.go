// Exemple d'intégration dans cmd/server/main.go
import (
    "net/http"
    "github.com/adlanegrm-oss/einvoice-saas/internal/handler"
)

// Dans votre fonction main() ou configuration des routes :
func RegisterRoutes() {
    valHandler := handler.NewValidationHandler()

    // Route API protégée par votre middleware d'authentification existant
    http.HandleFunc("/api/v1/validate", valHandler.HandleValidateDocument)
}
