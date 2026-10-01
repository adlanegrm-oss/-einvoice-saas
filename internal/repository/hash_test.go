package repository

import "testing"

func TestGenerateInvoiceHash(t *testing.T) {
	a := Invoice{InvoiceNumber: "F-1", Items: []InvoiceItem{{Description: "x", Quantity: 1, UnitPrice: 10}}}
	h1, err := GenerateInvoiceHash(a)
	if err != nil || len(h1) != 64 {
		t.Fatalf("empreinte SHA-256 attendue (64 hex) : %q, %v", h1, err)
	}
	if h2, _ := GenerateInvoiceHash(a); h1 != h2 {
		t.Error("l'empreinte doit être déterministe")
	}
	b := a
	b.Items = []InvoiceItem{{Description: "x", Quantity: 1, UnitPrice: 11}}
	if h3, _ := GenerateInvoiceHash(b); h1 == h3 {
		t.Error("une modification doit changer l'empreinte")
	}
}
