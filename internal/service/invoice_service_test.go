package service

import (
	"context"
	"einvoice-saas/internal/model"
	"einvoice-saas/internal/repository"
	"einvoice-saas/internal/validator"
	"errors"
	"sync"
	"testing"
)

type memoryInvoiceRepo struct {
	mu               sync.Mutex
	invoices         map[string]*repository.InvoiceRecord
	failCreate       bool
	failUpdateStatus bool
}

func newMemoryInvoiceRepo() *memoryInvoiceRepo {
	return &memoryInvoiceRepo{invoices: make(map[string]*repository.InvoiceRecord)}
}

func (m *memoryInvoiceRepo) Create(ctx context.Context, inv *repository.InvoiceRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failCreate {
		return errors.New("db_create_invoice_failed")
	}
	key := inv.TenantID + ":" + inv.ID
	if _, ok := m.invoices[key]; ok {
		return repository.ErrDuplicateBusiness
	}
	m.invoices[key] = inv
	return nil
}

func (m *memoryInvoiceRepo) GetByID(ctx context.Context, tenantID, id string) (*repository.InvoiceRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.invoices[tenantID+":"+id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return inv, nil
}

func (m *memoryInvoiceRepo) GetBySHA256(ctx context.Context, tenantID, sha256Hash string) (*repository.InvoiceRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, inv := range m.invoices {
		if inv.TenantID == tenantID && inv.DocumentSHA256 == sha256Hash {
			return inv, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *memoryInvoiceRepo) UpdateStatus(ctx context.Context, tenantID, id string, target repository.InvoiceStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failUpdateStatus {
		return errors.New("db_update_status_failed")
	}
	inv, ok := m.invoices[tenantID+":"+id]
	if !ok {
		return repository.ErrNotFound
	}
	inv.Status = target
	return nil
}

type memoryEventRepo struct {
	mu        sync.Mutex
	events    map[string][]repository.InvoiceEventRecord
	failEvent bool
}

func newMemoryEventRepo() *memoryEventRepo {
	return &memoryEventRepo{events: make(map[string][]repository.InvoiceEventRecord)}
}

func (m *memoryEventRepo) Append(ctx context.Context, ev *repository.InvoiceEventRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failEvent {
		return errors.New("db_append_event_failed")
	}
	key := ev.TenantID + ":" + ev.InvoiceID
	m.events[key] = append(m.events[key], *ev)
	return nil
}

func (m *memoryEventRepo) GetHistory(ctx context.Context, tenantID, invoiceID string) ([]repository.InvoiceEventRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events[tenantID+":"+invoiceID], nil
}

func (m *memoryEventRepo) GetLatestSequence(ctx context.Context, tenantID, invoiceID string) (int, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.events[tenantID+":"+invoiceID]
	if len(list) == 0 {
		return 0, "", nil
	}
	last := list[len(list)-1]
	return last.Sequence, last.CurrentHash, nil
}

type memoryIdemRepo struct {
	mu          sync.Mutex
	records     map[string]*repository.IdempotencyRecord
	shouldError bool
}

func newMemoryIdemRepo() *memoryIdemRepo {
	return &memoryIdemRepo{records: make(map[string]*repository.IdempotencyRecord)}
}

func (m *memoryIdemRepo) Save(ctx context.Context, rec *repository.IdempotencyRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shouldError {
		return errors.New("idempotency_store_failure")
	}
	m.records[rec.TenantID+":"+rec.Key] = rec
	return nil
}

func (m *memoryIdemRepo) Get(ctx context.Context, tenantID, key string) (*repository.IdempotencyRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.records[tenantID+":"+key]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return rec, nil
}

// 1. Cas nominal + replay idempotent
func TestInvoiceService_Ingest_Success_And_Idempotency(t *testing.T) {
	invRepo := newMemoryInvoiceRepo()
	evtRepo := newMemoryEventRepo()
	idemRepo := newMemoryIdemRepo()

	dummyValFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		return "UBL-2.1", &model.CanonicalInvoice{
			ID:        "INV-2026-001",
			IssueDate: "2026-10-08",
		}, true, map[string]string{"mock": "valid"}, nil
	}

	svc := NewInvoiceService(invRepo, evtRepo, idemRepo, dummyValFn)

	cmd := IngestionCommand{
		TenantID:       "tenant_test_1",
		IdempotencyKey: "idem_key_abc_123",
		Profile:        validator.ProfileEN16931,
		RawXML:         []byte("<Invoice>valid</Invoice>"),
		Actor:          "api_gateway",
	}

	// Premier appel
	res1, err := svc.IngestInvoice(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Premier appel Ã©chouÃ©: %v", err)
	}
	if res1.IsIdempotentReplay {
		t.Errorf("Le premier appel ne doit pas Ãªtre un rejeu")
	}
	if res1.Status != "accepted" {
		t.Errorf("Statut attendu 'accepted', obtenu: %s", res1.Status)
	}
	if res1.AuditHash == "" {
		t.Errorf("AuditHash manquant")
	}

	// Rejeu avec la mÃªme clÃ© et le mÃªme XML
	res2, err := svc.IngestInvoice(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Rejeu idempotent Ã©chouÃ©: %v", err)
	}
	if !res2.IsIdempotentReplay {
		t.Errorf("Le second appel aurait dÃ» Ãªtre dÃ©tectÃ© comme rejeu idempotent")
	}
	if res2.InvoiceID != res1.InvoiceID || res2.AuditHash != res1.AuditHash {
		t.Errorf("La rÃ©ponse rejouÃ©e ne correspond pas Ã  l'originale")
	}
}

// 2. Conflit d'idempotence : mÃªme clÃ©, XML diffÃ©rent
func TestInvoiceService_Ingest_IdempotencyConflict(t *testing.T) {
	invRepo := newMemoryInvoiceRepo()
	evtRepo := newMemoryEventRepo()
	idemRepo := newMemoryIdemRepo()

	dummyValFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		return "UBL-2.1", &model.CanonicalInvoice{
			ID:        "INV-2026-002",
			IssueDate: "2026-10-08",
		}, true, nil, nil
	}

	svc := NewInvoiceService(invRepo, evtRepo, idemRepo, dummyValFn)

	cmd := IngestionCommand{
		TenantID:       "tenant_test_1",
		IdempotencyKey: "shared_key_123",
		Profile:        validator.ProfileEN16931,
		RawXML:         []byte("<Invoice>initial</Invoice>"),
		Actor:          "api_gateway",
	}

	_, err := svc.IngestInvoice(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Premier appel Ã©chouÃ©: %v", err)
	}

	cmdConflict := cmd
	cmdConflict.RawXML = []byte("<Invoice>tampered_or_different</Invoice>")
	_, errConflict := svc.IngestInvoice(context.Background(), cmdConflict)
	if !errors.Is(errConflict, ErrIdempotencyConflict) {
		t.Fatalf("Attendu ErrIdempotencyConflict, obtenu: %v", errConflict)
	}
}

// 3. DÃ©duplication par empreinte documentaire SHA-256
func TestInvoiceService_Ingest_DuplicateDocument(t *testing.T) {
	invRepo := newMemoryInvoiceRepo()
	evtRepo := newMemoryEventRepo()
	idemRepo := newMemoryIdemRepo()

	dummyValFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		return "UBL-2.1", &model.CanonicalInvoice{
			ID:        "INV-2026-003",
			IssueDate: "2026-10-08",
		}, true, nil, nil
	}

	svc := NewInvoiceService(invRepo, evtRepo, idemRepo, dummyValFn)

	rawXML := []byte("<Invoice>identical_document_content</Invoice>")

	cmd1 := IngestionCommand{
		TenantID:       "tenant_test_1",
		IdempotencyKey: "key_first_call",
		Profile:        validator.ProfileEN16931,
		RawXML:         rawXML,
		Actor:          "api_gateway",
	}
	if _, err := svc.IngestInvoice(context.Background(), cmd1); err != nil {
		t.Fatalf("Ingestion initiale Ã©chouÃ©e: %v", err)
	}

	// MÃªme XML, mais clÃ© d'idempotence diffÃ©rente (tentative de soumission d'un doublon)
	cmd2 := IngestionCommand{
		TenantID:       "tenant_test_1",
		IdempotencyKey: "key_second_call",
		Profile:        validator.ProfileEN16931,
		RawXML:         rawXML,
		Actor:          "api_gateway",
	}
	_, errDup := svc.IngestInvoice(context.Background(), cmd2)
	if !errors.Is(errDup, ErrDuplicateDocument) {
		t.Fatalf("Attendu ErrDuplicateDocument, obtenu: %v", errDup)
	}
}

// 4. Facture non conforme : transition REJECTED + Audit scellÃ©
func TestInvoiceService_Ingest_Rejection_And_Audit(t *testing.T) {
	invRepo := newMemoryInvoiceRepo()
	evtRepo := newMemoryEventRepo()
	idemRepo := newMemoryIdemRepo()

	dummyValFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		return "UBL-2.1", &model.CanonicalInvoice{
			ID:        "INV-FAIL-001",
			IssueDate: "2026-10-08",
		}, false, map[string]string{"rule": "BR-01 failed"}, nil
	}

	svc := NewInvoiceService(invRepo, evtRepo, idemRepo, dummyValFn)

	cmd := IngestionCommand{
		TenantID:       "tenant_test_1",
		IdempotencyKey: "idem_rejected_invoice",
		Profile:        validator.ProfileCIUSFR,
		RawXML:         []byte("<Invoice>invalid_tax</Invoice>"),
		Actor:          "api_gateway",
	}

	res, err := svc.IngestInvoice(context.Background(), cmd)
	if err != nil {
		t.Fatalf("L'ingestion d'une facture non conforme ne doit pas renvoyer d'erreur systÃ¨me: %v", err)
	}

	if res.Status != "rejected" {
		t.Errorf("Statut attendu 'rejected', obtenu: %s", res.Status)
	}

	// VÃ©rifier la persistance du statut final dans le repository
	inv, err := invRepo.GetByID(context.Background(), "tenant_test_1", "INV-FAIL-001")
	if err != nil {
		t.Fatalf("Facture introuvable en base: %v", err)
	}
	if inv.Status != repository.StatusRejected {
		t.Errorf("Statut en base attendu REJECTED, obtenu: %s", inv.Status)
	}

	// VÃ©rifier qu'un Ã©vÃ©nement d'audit REJECTED a Ã©tÃ© consignÃ©
	events, err := evtRepo.GetHistory(context.Background(), "tenant_test_1", "INV-FAIL-001")
	if err != nil || len(events) != 1 {
		t.Fatalf("Attendu 1 Ã©vÃ©nement d'audit, obtenu: %d", len(events))
	}
	if events[0].EventType != string(repository.StatusRejected) {
		t.Errorf("Type d'Ã©vÃ©nement d'audit attendu REJECTED, obtenu: %s", events[0].EventType)
	}
}

// 5. Ã‰chec lors de UpdateStatus -> propagation d'erreur
func TestInvoiceService_Ingest_UpdateStatus_Failure(t *testing.T) {
	invRepo := newMemoryInvoiceRepo()
	invRepo.failUpdateStatus = true
	evtRepo := newMemoryEventRepo()
	idemRepo := newMemoryIdemRepo()

	dummyValFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		return "UBL-2.1", &model.CanonicalInvoice{
			ID:        "INV-FAIL-STATUS",
			IssueDate: "2026-10-08",
		}, true, nil, nil
	}

	svc := NewInvoiceService(invRepo, evtRepo, idemRepo, dummyValFn)

	cmd := IngestionCommand{
		TenantID: "tenant_test_1",
		Profile:  validator.ProfileEN16931,
		RawXML:   []byte("<Invoice>test</Invoice>"),
		Actor:    "api_gateway",
	}

	_, err := svc.IngestInvoice(context.Background(), cmd)
	if err == nil {
		t.Fatalf("Attendu une erreur lors de l'Ã©chec d'UpdateStatus, obtenu nil")
	}
}

// 6. Ã‰chec lors du Save de l'idempotence -> propagation d'erreur
func TestInvoiceService_Ingest_IdempotencySave_Failure(t *testing.T) {
	invRepo := newMemoryInvoiceRepo()
	evtRepo := newMemoryEventRepo()
	idemRepo := newMemoryIdemRepo()
	idemRepo.shouldError = true

	dummyValFn := func(xmlData []byte, profile validator.ValidationProfile) (string, *model.CanonicalInvoice, bool, interface{}, error) {
		return "UBL-2.1", &model.CanonicalInvoice{
			ID:        "INV-IDEM-FAIL",
			IssueDate: "2026-10-08",
		}, true, nil, nil
	}

	svc := NewInvoiceService(invRepo, evtRepo, idemRepo, dummyValFn)

	cmd := IngestionCommand{
		TenantID:       "tenant_test_1",
		IdempotencyKey: "key_save_fail",
		Profile:        validator.ProfileEN16931,
		RawXML:         []byte("<Invoice>test</Invoice>"),
		Actor:          "api_gateway",
	}

	_, err := svc.IngestInvoice(context.Background(), cmd)
	if err == nil {
		t.Fatalf("Attendu une erreur lors de l'Ã©chec de persistance de l'idempotence, obtenu nil")
	}
}
