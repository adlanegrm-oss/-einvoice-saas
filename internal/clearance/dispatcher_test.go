package clearance

import (
	"context"
	"testing"

	"einvoice-saas/internal/model"
)

func TestResolveChannel(t *testing.T) {
	fr := &model.CanonicalInvoice{Seller: model.Party{CountryCode: "FR"}}
	if ResolveChannel(fr, ChannelNone) != ChannelPPF {
		t.Fatalf("FR → PPF")
	}
	pl := &model.CanonicalInvoice{Seller: model.Party{CountryCode: "PL"}}
	if ResolveChannel(pl, ChannelNone) != ChannelKSeF {
		t.Fatalf("PL → KSEF")
	}
	ma := &model.CanonicalInvoice{Seller: model.Party{CountryCode: "MA"}}
	if ResolveChannel(ma, ChannelNone) != ChannelAS4 {
		t.Fatalf("MA → AS4")
	}
	if ResolveChannel(fr, ChannelAS4) != ChannelAS4 {
		t.Fatalf("explicit channel must win")
	}
}

func TestDispatcher_PPF_DryRun(t *testing.T) {
	d := NewDispatcher(NewPPFConnector("", "", "", true))
	req := TransmissionRequest{
		InvoiceID:  "INV-1",
		TenantID:   "t1",
		Channel:    ChannelPPF,
		Canonical:  &model.CanonicalInvoice{ID: "INV-1", Seller: model.Party{CountryCode: "FR"}},
		XMLPayload: []byte(`<Invoice/>`),
	}
	res, err := d.Transmit(context.Background(), req)
	if err != nil {
		t.Fatalf("Transmit: %v", err)
	}
	if !res.Accepted || res.RemoteStatus != "DEPOSEE" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestDispatcher_UnknownChannel(t *testing.T) {
	d := NewDispatcher() // aucun connecteur
	_, err := d.Transmit(context.Background(), TransmissionRequest{
		Channel:    ChannelPPF,
		XMLPayload: []byte(`<Invoice/>`),
	})
	if err == nil {
		t.Fatal("expected error for missing connector")
	}
}
