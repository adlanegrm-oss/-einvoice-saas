package service_test

import (
"context"
"database/sql"
"errors"
"testing"
"time"

_ "modernc.org/sqlite"

"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/rulesets"
"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
"github.com/adlanegrm-oss/einvoice-saas/internal/service"
"github.com/adlanegrm-oss/einvoice-saas/pkg/canonical"
)

func setupTestPipelineDB(t *testing.T) *sql.DB {
t.Helper()
db, err := sql.Open("sqlite", ":memory:")
if err != nil {
t.Fatalf("échec ouverture base mémoire: %v", err)
}
return db
}

func createSampleValidInvoice() invoice.Invoice {
return invoice.Invoice{
ID:        "INV-TEST-001",
Number:    "FA-2026-0001",
Currency:  invoice.EUR,
IssueDate: time.Now(),
Seller: invoice.Party{
Name:  "Fournisseur Test SAS",
SIRET: "12345678901234",
VATID: "FR12345678901",
},
Customer: invoice.Party{
Name:  "Client Destinataire SAS",
SIRET: "98765432109876",
VATID: "FR98765432109",
},
Items: []invoice.InvoiceItem{
{
Description: "Prestation d'intégration SaaS",
Quantity:    1,
UnitPrice:   invoice.NewMoneyFromFloat(1000.0, 2, invoice.EUR),
VATRate:     invoice.NewMoneyFromFloat(20.0, 2, invoice.EUR),
TotalHT:     invoice.NewMoneyFromFloat(1000.0, 2, invoice.EUR),
},
},
TotalHT:  invoice.NewMoneyFromFloat(1000.0, 2, invoice.EUR),
TotalVAT: invoice.NewMoneyFromFloat(200.0, 2, invoice.EUR),
TotalTTC: invoice.NewMoneyFromFloat(1200.0, 2, invoice.EUR),
}
}

type customMockDir struct{}
func (m *customMockDir) Lookup(ctx context.Context, participantID string) (*dispatcher.TargetEndpoint, error) {
return &dispatcher.TargetEndpoint{ReceiverID: participantID, PlatformName: "PDP Mock", AS4Endpoint: "https://mock.pdp"}, nil
}

type customMockAS4 struct{}
func (m *customMockAS4) SendPayload(ctx context.Context, ep *dispatcher.TargetEndpoint, p []byte) (string, error) {
return "ACK-MOCK-OK", nil
}

func TestPipeline_CustomValidator_Integration(t *testing.T) {
db := setupTestPipelineDB(t)
defer db.Close()

repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		t.Fatalf("échec création repo: %v", err)
	}
val := en16931.NewValidator()
sm := status.NewStateMachine()
disp := dispatcher.NewDispatcher(&customMockDir{}, &customMockAS4{})

reg := rulesets.NewRegistryValidator()
// Règle spécifique Client M : Le numéro de facture doit être FA-CLIENTM-2026
reg.RegisterRule("CLIENT_M", func(inv *canonical.CanonicalInvoice) error {
if inv.InvoiceNumber == "" || inv.InvoiceNumber != "FA-CLIENTM-2026" {
return errors.New("BR-CLIENT-M-99: La facture doit impérativement porter le numéro FA-CLIENTM-2026")
}
return nil
})

pipeline := service.NewInvoicePipeline(repo, val, sm, disp).WithCustomValidator(reg)
ctx := context.Background()

// CAS 1 : Facture soumise par un client standard -> DOIT PASSER (règles isolées)
stdInv := createSampleValidInvoice()
stdInv.ID = "INV-STD-01"
stdInv.Number = "FA-STD-0001"
resStd, err := pipeline.ProcessAndEmit(ctx, "STANDARD_TENANT", stdInv)
if err != nil {
t.Fatalf("erreur inattendue pour tenant standard : %v", err)
}
if resStd.Status != status.StateTransmitted {
t.Fatalf("attendu StateTransmitted pour standard, reçu %s (erreur: %s)", resStd.Status, resStd.Error)
}

// CAS 2 : Facture non conforme aux exigences du CLIENT_M -> DOIT ÉCHOUER (rejet et audit sans émission)
mInvInvalid := createSampleValidInvoice()
mInvInvalid.ID = "INV-M-FAIL-01"
mInvInvalid.Number = "FA-STD-0001"
resM, err := pipeline.ProcessAndEmit(ctx, "CLIENT_M", mInvInvalid)
if err != nil {
t.Fatalf("l'échec de règle ne doit pas faire paniquer le pipeline : %v", err)
}
if resM.Status != status.StateRejected {
t.Fatalf("attendu StateRejected pour CLIENT_M invalide, reçu %s", resM.Status)
}
if resM.Error != "BR-CLIENT-M-99: La facture doit impérativement porter le numéro FA-CLIENTM-2026" {
t.Fatalf("message d'erreur inattendu pour CLIENT_M: %s", resM.Error)
}

// CAS 3 : Facture conforme aux exigences du CLIENT_M -> DOIT PASSER
mInvValid := createSampleValidInvoice()
mInvValid.ID = "INV-M-OK-01"
mInvValid.Number = "FA-CLIENTM-2026"
resMOk, err := pipeline.ProcessAndEmit(ctx, "CLIENT_M", mInvValid)
if err != nil {
t.Fatalf("erreur inattendue pour CLIENT_M conforme : %v", err)
}
if resMOk.Status != status.StateTransmitted {
t.Fatalf("attendu StateTransmitted pour CLIENT_M valide, reçu %s", resMOk.Status)
}
}




