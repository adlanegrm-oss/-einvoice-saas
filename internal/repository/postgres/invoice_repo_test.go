package postgres

import (
	"testing"

	"einvoice-saas/internal/repository"
)

func TestRepositoriesImplementInterfaces(t *testing.T) {
	var _ repository.InvoiceRepository = (*InvoiceRepo)(nil)
	var _ repository.IdempotencyRepository = (*IdempotencyRepo)(nil)
	var _ repository.EventRepository = (*EventRepo)(nil)
}
