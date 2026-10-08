package security

import (
	"context"
	"testing"
	"time"
)

func TestJWT_RoundTrip(t *testing.T) {
	secret := "test-secret-32-bytes-minimum!!"
	v := NewJWTValidator(secret, "einvoice", "api")
	now := time.Now()
	c := Claims{
		Subject:   "user-1",
		Issuer:    "einvoice",
		Audience:  "api",
		TenantID:  "tenant-1",
		Roles:     []string{"operator"},
		IssuedAt:  now.Unix(),
		NotBefore: now.Add(-time.Minute).Unix(),
		ExpiresAt: now.Add(time.Hour).Unix(),
	}
	tok, err := SignHS256([]byte(secret), c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := v.ParseAndValidate(tok)
	if err != nil {
		t.Fatal(err)
	}
	if got.TenantID != "tenant-1" || got.Subject != "user-1" {
		t.Fatalf("claims: %+v", got)
	}
}

func TestJWT_Expired(t *testing.T) {
	secret := "test-secret-32-bytes-minimum!!"
	v := NewJWTValidator(secret, "einvoice", "api")
	c := Claims{
		Subject: "u", Issuer: "einvoice", Audience: "api", TenantID: "t1",
		ExpiresAt: time.Now().Add(-time.Hour).Unix(),
	}
	tok, _ := SignHS256([]byte(secret), c)
	if _, err := v.ParseAndValidate(tok); err != ErrTokenExpired {
		t.Fatalf("want expired, got %v", err)
	}
}

func TestRBAC(t *testing.T) {
	if !HasPermission([]string{"operator"}, PermInvoiceWrite) {
		t.Fatal("operator should write")
	}
	if HasPermission([]string{"auditor"}, PermInvoiceTransmit) {
		t.Fatal("auditor must not transmit")
	}
	if !HasPermission([]string{"admin"}, PermAuditRead) {
		t.Fatal("admin should read audit")
	}
}

func TestTenantAssert(t *testing.T) {
	ctx := ContextWithClaims(context.Background(), &Claims{TenantID: "t1", Subject: "u", Roles: []string{"admin"}})
	if err := AssertTenant(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	if err := AssertTenant(ctx, "t2"); err != ErrTenantMismatch {
		t.Fatalf("want mismatch, got %v", err)
	}
	if err := RequirePermission(ctx, PermInvoiceRead); err != nil {
		t.Fatal(err)
	}
}

func TestUploadGuard(t *testing.T) {
	if err := ValidateXMLUpload([]byte(`<Invoice/>`), 1024); err != nil {
		t.Fatal(err)
	}
	if err := ValidateXMLUpload([]byte(`<!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><Invoice/>`), 1024); err != ErrUploadXXE {
		t.Fatalf("want XXE, got %v", err)
	}
	if AllowedFilename("../etc/passwd") {
		t.Fatal("path traversal should be rejected")
	}
}
