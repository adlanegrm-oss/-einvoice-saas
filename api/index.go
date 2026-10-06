package handler

import (
	"net/http"

	"einvoice-saas/internal/app"
	"einvoice-saas/internal/compliance/validators/fr"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/security"
	"einvoice-saas/internal/service"
	"einvoice-saas/internal/validator"
)

var appHandler http.Handler

func init() {
	schematronEngine := validator.NewSchematronEngine(nil)
	normativeValidator := validator.NewNormativeValidator(true)
	frFiscalValidator := fr.NewFranceCanonicalValidator()

	valFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		resp, err := app.ExecuteValidationPipeline(xmlData, profile, schematronEngine, normativeValidator, frFiscalValidator)
		if err != nil {
			return "", nil, false, resp, err
		}
		return resp.Syntax, resp.CanonicalInvoice, resp.Valid, resp, nil
	}

	invRepo := app.NewInMemInvoiceRepo()
	evtRepo := app.NewInMemEventRepo()
	idemRepo := app.NewInMemIdemRepo()

	invoiceSvc := service.NewInvoiceService(invRepo, evtRepo, idemRepo, valFn)
	keyStore := security.NewInMemoryKeyStore()

	appHandler = app.SetupRouter(keyStore, invoiceSvc)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	appHandler.ServeHTTP(w, r)
}
