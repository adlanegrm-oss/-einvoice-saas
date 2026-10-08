package evidence

import (
	"testing"
)

func TestAuditStore_AppendAndChain(t *testing.T) {
	s := NewStore()
	e1 := s.Append("t1", "INV-1", EventReceived, "INFO", "received", []byte("<xml/>"), nil)
	e2 := s.Append("t1", "INV-1", EventValidated, "INFO", "ok", nil, map[string]string{"profile": "EN16931"})
	if e1.ChainHash == "" || e2.PrevHash != e1.ChainHash {
		t.Fatalf("chain link broken: %+v -> %+v", e1, e2)
	}
	if err := s.VerifyChain(); err != nil {
		t.Fatal(err)
	}
	list := s.ListByInvoice("t1", "INV-1")
	if len(list) != 2 {
		t.Fatalf("want 2 events, got %d", len(list))
	}
	raw, err := s.ExportJSON("t1", "INV-1")
	if err != nil || len(raw) < 10 {
		t.Fatalf("export: %v %s", err, raw)
	}
}

func TestAuditStore_TenantIsolation(t *testing.T) {
	s := NewStore()
	s.Append("t1", "INV-1", EventReceived, "INFO", "a", nil, nil)
	s.Append("t2", "INV-1", EventReceived, "INFO", "b", nil, nil)
	if len(s.ListByInvoice("t1", "INV-1")) != 1 {
		t.Fatal("tenant isolation failed")
	}
}
