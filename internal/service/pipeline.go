package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
)

type InvoicePipeline struct {
	repo       *repository.SQLiteInvoiceRepository
	validator  *en16931.Validator
	sm         *status.StateMachine
	dispatcher *dispatcher.Dispatcher
}

func NewInvoicePipeline(
	repo *repository.SQLiteInvoiceRepository,
	val *en16931.Validator,
	sm *status.StateMachine,
	disp *dispatcher.Dispatcher,
) *InvoicePipeline {
	return &InvoicePipeline{
		repo:       repo,
		validator:  val,
		sm:         sm,
		dispatcher: disp,
	}
}

type PipelineResult struct {
	InvoiceID   string              `json:"invoice_id"`
	Status      status.InvoiceState `json:"status"`
	PayloadHash string              `json:"payload_hash"`
	Error       string              `json:"error,omitempty"`
}

// IngestionResult est un alias de rétrocompatibilité pour les tests
type IngestionResult = PipelineResult

func (p *InvoicePipeline) ProcessAndEmit(ctx context.Context, tenantID string, inv invoice.Invoice) (*PipelineResult, error) {
	rawBytes, err := json.Marshal(inv)
	if err != nil {
		return nil, fmt.Errorf("sérialisation facture : %w", err)
	}

	sum := sha256.Sum256(rawBytes)
	hash := hex.EncodeToString(sum[:])
	now := time.Now().UTC()

	// 1. Dépôt initial
	depEvent := &status.StatusEvent{
		ID:        fmt.Sprintf("dep_%s_%d", inv.ID, time.Now().UnixNano()),
		InvoiceID: inv.ID,
		FromState: status.StateDeposited,
		ToState:   status.StateDeposited,
		Actor:     tenantID,
		Reason:    "Dépôt initial de la facture",
		Timestamp: now,
		Signature: hash,
	}
	if err := p.repo.RecordStatusTransition(ctx, depEvent, inv.ID, hash); err != nil {
		return nil, fmt.Errorf("enregistrement dépôt : %w", err)
	}

	// 2. Validation réglementaire EN 16931
	valErrors := p.validator.ValidateInvoice(&inv)
	if len(valErrors) > 0 {
		var errMsgs []string
		for _, ve := range valErrors {
			errMsgs = append(errMsgs, ve.Error())
		}
		joinedErr := strings.Join(errMsgs, "; ")

		rejEvent, err := p.sm.Transition(inv.ID, status.StateDeposited, status.StateRejected, "VALIDATOR_EN16931", joinedErr, hash, "")
		if err != nil {
			return nil, fmt.Errorf("transition rejet validation : %w", err)
		}
		rejEvent.ID = fmt.Sprintf("rej_%s_%d", inv.ID, time.Now().UnixNano())
		rejEvent.Signature = hash
		_ = p.repo.RecordStatusTransition(ctx, rejEvent, inv.ID, hash)

		return &PipelineResult{
			InvoiceID:   inv.ID,
			Status:      status.StateRejected,
			PayloadHash: hash,
			Error:       joinedErr,
		}, nil
	}

	// 3. Émission réussie
	issEvent, err := p.sm.Transition(inv.ID, status.StateDeposited, status.StateIssued, "SYSTEM_PDP", "Conformité EN 16931 validée", hash, "")
	if err != nil {
		return nil, err
	}
	issEvent.ID = fmt.Sprintf("iss_%s_%d", inv.ID, time.Now().UnixNano())
	issEvent.Signature = hash
	if err := p.repo.RecordStatusTransition(ctx, issEvent, inv.ID, hash); err != nil {
		return nil, fmt.Errorf("enregistrement émission : %w", err)
	}

	// 4. Routage & Dispatch vers la PDP destinataire
	_, dispErr := p.dispatcher.RouteAndDispatch(ctx, inv.Customer.SIRET, rawBytes)
	if dispErr != nil {
		rejEvent, err := p.sm.Transition(inv.ID, status.StateIssued, status.StateRejected, "DISPATCHER", dispErr.Error(), hash, issEvent.Signature)
		if err != nil {
			return nil, fmt.Errorf("transition rejet dispatch : %w", err)
		}
		rejEvent.ID = fmt.Sprintf("rej_%s_%d", inv.ID, time.Now().UnixNano())
		rejEvent.Signature = hash
		_ = p.repo.RecordStatusTransition(ctx, rejEvent, inv.ID, hash)

		return &PipelineResult{
			InvoiceID:   inv.ID,
			Status:      status.StateRejected,
			PayloadHash: hash,
			Error:       dispErr.Error(),
		}, nil
	}

	transEvent, err := p.sm.Transition(inv.ID, status.StateIssued, status.StateTransmitted, "DISPATCHER", "Acheminé avec succès vers le destinataire", hash, issEvent.Signature)
	if err != nil {
		return nil, fmt.Errorf("transition transmission : %w", err)
	}
	transEvent.ID = fmt.Sprintf("tra_%s_%d", inv.ID, time.Now().UnixNano())
	transEvent.Signature = hash
	if err := p.repo.RecordStatusTransition(ctx, transEvent, inv.ID, hash); err != nil {
		return nil, fmt.Errorf("enregistrement transmission : %w", err)
	}

	return &PipelineResult{
		InvoiceID:   inv.ID,
		Status:      status.StateTransmitted,
		PayloadHash: hash,
	}, nil
}

// TransitionInvoiceStatus enregistre les statuts du cycle acheteur (SUSPENDED, APPROVED, PAID, etc.)
func (p *InvoicePipeline) TransitionInvoiceStatus(ctx context.Context, invoiceID string, targetStatus status.InvoiceState, actor, reason string) error {
	history, err := p.repo.GetStatusHistory(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("lecture audit trail : %w", err)
	}
	if len(history) == 0 {
		return fmt.Errorf("facture %s introuvable dans l'historique", invoiceID)
	}

	lastEvent := history[len(history)-1]
	currentStatus := lastEvent.ToState
	prevHash := lastEvent.Signature
	payloadHash := history[0].Signature

	event, err := p.sm.Transition(invoiceID, currentStatus, targetStatus, actor, reason, payloadHash, prevHash)
	if err != nil {
		return err
	}

	event.ID = fmt.Sprintf("tr_%s_%d", invoiceID, time.Now().UnixNano())
	event.Signature = payloadHash

	return p.repo.RecordStatusTransition(ctx, event, invoiceID, payloadHash)
}
