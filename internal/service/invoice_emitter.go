package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/audit"
	"github.com/adlanegrm-oss/einvoice-saas/internal/currency"
)

type InvoiceDocument struct {
	ID            string
	TenantID      string
	InvoiceNumber string
	CustomerSIRET string
	NetAmount     currency.Amount
	TaxAmount     currency.Amount
	TotalAmount   currency.Amount
	RawXML        []byte // Résultat compilé UBL 2.1 ou CII opposable
}

type InvoiceEmitterService struct {
	db *sql.DB
}

func NewInvoiceEmitterService(db *sql.DB) *InvoiceEmitterService {
	return &InvoiceEmitterService{db: db}
}

// EmitInvoiceTransaction garantit l'atomicité totale de l'émission
func (s *InvoiceEmitterService) EmitInvoiceTransaction(
	ctx context.Context,
	idempotencyKey string,
	doc InvoiceDocument,
) error {
	// Calcul de l'empreinte fiscale opposable sur le XML canonique (et NON sur le JSON)
	h := sha256.New()
	h.Write(doc.RawXML)
	canonicalHash := hex.EncodeToString(h.Sum(nil))

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("début de transaction échoué: %w", err)
	}
	defer tx.Rollback()

	// 1. Verrou d'idempotence multi-instances
	now := time.Now().UTC()
	var status string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO idempotency_keys (tenant_id, idempotency_key, request_hash, status, expires_at)
		VALUES ($1, $2, $3, 'PENDING', $4)
		ON CONFLICT (tenant_id, idempotency_key) DO UPDATE
		SET tenant_id = EXCLUDED.tenant_id
		RETURNING status;
	`, doc.TenantID, idempotencyKey, canonicalHash, now.Add(24*time.Hour)).Scan(&status)

	if err != nil {
		return fmt.Errorf("concurrence/idempotence: %w", err)
	}
	if status == "PROCESSED" {
		return nil // Facture déjà émise avec succès
	}

	// 2. Persistance de la Facture avec son document opposable
	_, err = tx.ExecContext(ctx, `
		INSERT INTO invoices (
			id, tenant_id, invoice_number, status, document_format,
			payload_hash, raw_payload, net_amount, tax_amount, total_amount, created_at
		) VALUES ($1, $2, $3, 'ISSUED', 'UBL_2_1', $4, $5, $6, $7, $8, $9);
	`, doc.ID, doc.TenantID, doc.InvoiceNumber, canonicalHash, doc.RawXML,
		doc.NetAmount.StringFixed(2), doc.TaxAmount.StringFixed(2), doc.TotalAmount.StringFixed(2), now)
	if err != nil {
		return fmt.Errorf("insertion facture: %w", err)
	}

	// 3. Premier bloc d'audit : Événement ISSUED
	auditEvt := audit.CreateNextEvent(doc.TenantID, doc.ID, "INVOICE_ISSUED", canonicalHash, nil)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_events (
			id, tenant_id, invoice_id, sequence_num, event_type,
			payload_hash, prev_event_hash, event_hash, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`, auditEvt.ID, auditEvt.TenantID, auditEvt.InvoiceID, auditEvt.SequenceNum,
		auditEvt.EventType, auditEvt.PayloadHash, auditEvt.PrevEventHash, auditEvt.EventHash, auditEvt.CreatedAt)
	if err != nil {
		return fmt.Errorf("insertion audit: %w", err)
	}

	// 4. Message Transactional Outbox (AS4 Dispatcher)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactional_outbox (
			tenant_id, invoice_id, destination_id, transport_type, payload_ref, status, next_retry_at
		) VALUES ($1, $2, $3, 'AS4', $4, 'PENDING', $5);
	`, doc.TenantID, doc.ID, doc.CustomerSIRET, canonicalHash, now)
	if err != nil {
		return fmt.Errorf("insertion outbox: %w", err)
	}

	// 5. Clôture de l'idempotence
	_, err = tx.ExecContext(ctx, `
		UPDATE idempotency_keys
		SET status = 'PROCESSED', response_code = 201
		WHERE tenant_id = $1 AND idempotency_key = $2;
	`, doc.TenantID, idempotencyKey)
	if err != nil {
		return fmt.Errorf("mise à jour statut idempotence: %w", err)
	}

	return tx.Commit()
}
