package status

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

type InvoiceState string

const (
	// Ingestion & Validation amont
	StateIngested         InvoiceState = "INGESTED"
	StateValidating       InvoiceState = "VALIDATING"
	StateValidationFailed InvoiceState = "VALIDATION_FAILED"

	// Statuts émetteur réglementaires (socle DGFIP)
	StateDeposited   InvoiceState = "DEPOSITED"
	StateIssued      InvoiceState = "ISSUED"
	StateTransmitted InvoiceState = "TRANSMITTED"
	StateRejected    InvoiceState = "REJECTED"

	// Clearance technique (PPF / PDP / KSeF)
	StateClearancePending  InvoiceState = "CLEARANCE_PENDING"
	StateCleared           InvoiceState = "CLEARED"
	StateClearanceRejected InvoiceState = "CLEARANCE_REJECTED"

	// Routage eDelivery AS4 / PEPPOL
	StateTransportPending InvoiceState = "TRANSPORT_PENDING"
	StateAS4Sent          InvoiceState = "AS4_SENT"
	StateDeliveredNRR     InvoiceState = "DELIVERED_NRR"
	StateTransportFailed  InvoiceState = "TRANSPORT_FAILED"

	// Statuts récepteur / acheteur (cycle PDP étendu)
	StateReceived        InvoiceState = "RECEIVED"
	StateSuspended       InvoiceState = "SUSPENDED"
	StateApproved        InvoiceState = "APPROVED"
	StateRejectedByBuyer InvoiceState = "REJECTED_BY_BUYER"
	StatePaid            InvoiceState = "PAID"

	// Échecs terminaux
	StateTerminalFailed InvoiceState = "TERMINAL_FAILED"

	// Alias rétrocompatibilité
	StatusDeposited       = StateDeposited
	StatusIssued          = StateIssued
	StatusTransmitted     = StateTransmitted
	StatusRejected        = StateRejected
	StatusReceived        = StateReceived
	StatusSuspended       = StateSuspended
	StatusApproved        = StateApproved
	StatusRejectedByBuyer = StateRejectedByBuyer
	StatusPaid            = StatePaid
)

func (s InvoiceState) String() string {
	return string(s)
}

func (s InvoiceState) IsTerminal() bool {
	return s == StateDeliveredNRR || s == StatePaid || s == StateTerminalFailed || s == StateRejected
}

// StatusEvent représente une transition d'état auditée avec intégrité cryptographique (PAF)
type StatusEvent struct {
	ID          string       `json:"id,omitempty"`
	InvoiceID   string       `json:"invoice_id"`
	FromState   InvoiceState `json:"from_state"`
	ToState     InvoiceState `json:"to_state"`
	Actor       string       `json:"actor"`
	Reason      string       `json:"reason,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
	PayloadHash string       `json:"payload_hash,omitempty"`
	PrevHash    string       `json:"prev_hash,omitempty"`
	Hash        string       `json:"hash,omitempty"`
	Signature   string       `json:"signature,omitempty"`
}

// ComputeHash calcule le condensat SHA-256 scellant l'entrée du journal d'audit
func (e *StatusEvent) ComputeHash() string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s",
		e.InvoiceID,
		e.FromState,
		e.ToState,
		e.Actor,
		e.Reason,
		e.Timestamp.UTC().Format(time.RFC3339Nano),
		e.PayloadHash,
		e.PrevHash,
	)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

var (
	ErrInvalidTransition = errors.New("transition de statut PDP interdite")
	ErrTerminalState     = errors.New("le document est dans un état terminal immuable")
)

type StateMachine struct {
	mu                 sync.RWMutex
	allowedTransitions map[InvoiceState]map[InvoiceState]bool
}

func NewStateMachine() *StateMachine {
	sm := &StateMachine{
		allowedTransitions: make(map[InvoiceState]map[InvoiceState]bool),
	}

	// 1. Ingestion et validation
	sm.allow(StateIngested, StateValidating)
	sm.allow(StateValidating, StateDeposited)
	sm.allow(StateValidating, StateValidationFailed)
	sm.allow(StateValidationFailed, StateTerminalFailed)

	// 2. Cycle émetteur & Clearance
	sm.allow(StateDeposited, StateIssued)
	sm.allow(StateDeposited, StateClearancePending)
	sm.allow(StateDeposited, StateRejected)

	sm.allow(StateClearancePending, StateCleared)
	sm.allow(StateClearancePending, StateClearanceRejected)
	sm.allow(StateClearanceRejected, StateClearancePending) // retry
	sm.allow(StateClearanceRejected, StateTerminalFailed)

	sm.allow(StateCleared, StateIssued)
	sm.allow(StateCleared, StateTransportPending)

	sm.allow(StateIssued, StateTransmitted)
	sm.allow(StateIssued, StateTransportPending)
	sm.allow(StateIssued, StateRejected)

	// 3. Transport AS4 / PEPPOL
	sm.allow(StateTransportPending, StateAS4Sent)
	sm.allow(StateTransportPending, StateTransportFailed)
	sm.allow(StateTransportFailed, StateTransportPending) // retry
	sm.allow(StateTransportFailed, StateTerminalFailed)

	sm.allow(StateAS4Sent, StateDeliveredNRR)
	sm.allow(StateAS4Sent, StateTransmitted)
	sm.allow(StateAS4Sent, StateTransportFailed)

	// 4. Cycle récepteur PDP
	sm.allow(StateTransmitted, StateReceived)
	sm.allow(StateTransmitted, StateRejected)
	sm.allow(StateDeliveredNRR, StateReceived)

	sm.allow(StateReceived, StateApproved)
	sm.allow(StateReceived, StateSuspended)
	sm.allow(StateReceived, StateRejectedByBuyer)

	// 5. Déblocage & Clôture
	sm.allow(StateSuspended, StateApproved)
	sm.allow(StateSuspended, StateRejectedByBuyer)
	sm.allow(StateApproved, StatePaid)

	return sm
}

func (sm *StateMachine) allow(from, to InvoiceState) {
	if _, ok := sm.allowedTransitions[from]; !ok {
		sm.allowedTransitions[from] = make(map[InvoiceState]bool)
	}
	sm.allowedTransitions[from][to] = true
}

func (sm *StateMachine) CanTransition(from, to InvoiceState) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Idempotence : rester dans le même état est accepté
	if from == to {
		return true
	}
	targets, ok := sm.allowedTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}

// Transition valide la transition et produit un StatusEvent chaîné (PAF)
func (sm *StateMachine) Transition(invoiceID string, from, to InvoiceState, actor, reason, payloadHash, prevHash string) (*StatusEvent, error) {
	if !sm.CanTransition(from, to) {
		if from.IsTerminal() && from != to {
			return nil, fmt.Errorf("%w : %s est un état terminal", ErrTerminalState, from)
		}
		return nil, fmt.Errorf("%w : impossible de passer de %s à %s", ErrInvalidTransition, from, to)
	}

	event := &StatusEvent{
		InvoiceID:   invoiceID,
		FromState:   from,
		ToState:     to,
		Actor:       actor,
		Reason:      reason,
		Timestamp:   time.Now().UTC(),
		PayloadHash: payloadHash,
		PrevHash:    prevHash,
	}
	event.Hash = event.ComputeHash()
	return event, nil
}

func (sm *StateMachine) ValidateTransition(from, to InvoiceState) error {
	if !sm.CanTransition(from, to) {
		return fmt.Errorf("%w : impossible de passer de %s à %s", ErrInvalidTransition, from, to)
	}
	return nil
}