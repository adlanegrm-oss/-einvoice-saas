package repository

import (
	"sync"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
)

// InvoiceRepository définit le contrat pour le stockage des factures
type InvoiceRepository interface {
	Save(inv invoice.Invoice) error
	GetAll() ([]invoice.Invoice, error)
}

// MemoryInvoiceRepository implémente InvoiceRepository en mémoire
type MemoryInvoiceRepository struct {
	store []invoice.Invoice
	mu    sync.RWMutex
}

// NewMemoryInvoiceRepository initialise un nouveau stockage mémoire
func NewMemoryInvoiceRepository() *MemoryInvoiceRepository {
	return &MemoryInvoiceRepository{
		store: make([]invoice.Invoice, 0),
	}
}

// Save ajoute une facture valide au stockage
func (r *MemoryInvoiceRepository) Save(inv invoice.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store = append(r.store, inv)
	return nil
}

// GetAll renvoie l'ensemble des factures
func (r *MemoryInvoiceRepository) GetAll() ([]invoice.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.store, nil
}
