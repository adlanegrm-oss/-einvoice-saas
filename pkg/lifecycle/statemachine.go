package lifecycle

import (
	"errors"
	"fmt"
)

type InvoiceStatus string

const (
	StatusDraft            InvoiceStatus = "DRAFT"
	StatusValidating       InvoiceStatus = "VALIDATING"
	StatusValidated        InvoiceStatus = "VALIDATED"
	StatusQueued           InvoiceStatus = "QUEUED"
	StatusSubmitted        InvoiceStatus = "SUBMITTED"
	StatusAccepted         InvoiceStatus = "ACCEPTED"
	StatusIssued           InvoiceStatus = "ISSUED"
	StatusRejected         InvoiceStatus = "REJECTED"
	StatusTechnicalError   InvoiceStatus = "TECHNICAL_ERROR"
	StatusBusinessRejected InvoiceStatus = "BUSINESS_REJECTED"
)

var validTransitions = map[InvoiceStatus][]InvoiceStatus{
	StatusDraft:            {StatusValidating},
	StatusValidating:       {StatusValidated, StatusRejected},
	StatusValidated:        {StatusQueued},
	StatusQueued:           {StatusSubmitted},
	StatusSubmitted:        {StatusAccepted, StatusTechnicalError, StatusBusinessRejected},
	StatusTechnicalError:   {StatusQueued, StatusRejected},
	StatusAccepted:         {StatusIssued},
	StatusIssued:           {},
	StatusRejected:         {},
	StatusBusinessRejected: {},
}

var (
	ErrInvalidTransition = errors.New("transition de statut interdite")
	ErrInvoiceImmutable  = errors.New("document emis scelle : modification interdite")
)

func TransitionStatus(current, target InvoiceStatus) (InvoiceStatus, error) {
	if current == StatusIssued {
		return current, ErrInvoiceImmutable
	}

	allowed, ok := validTransitions[current]
	if !ok {
		return current, fmt.Errorf("%w: statut source inconnu %s", ErrInvalidTransition, current)
	}

	for _, s := range allowed {
		if s == target {
			return target, nil
		}
	}

	return current, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current, target)
}
