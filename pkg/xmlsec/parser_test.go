package xmlsec

import (
"errors"
"strings"
"testing"
)

func TestBlockXXE(t *testing.T) {
xxePayload := `<?xml version="1.0"?><!DOCTYPE root [<!ENTITY test SYSTEM "file:///etc/passwd">]><root>&test;</root>`
err := ValidateSecureXML(strings.NewReader(xxePayload))
if !errors.Is(err, ErrDTDForbidden) {
t.Fatalf("attendu ErrDTDForbidden, obtenu %v", err)
}
}

func TestValidXML(t *testing.T) {
cleanXML := `<Invoice><ID>INV-001</ID><Amount>1000</Amount></Invoice>`
err := ValidateSecureXML(strings.NewReader(cleanXML))
if err != nil {
t.Fatalf("attendu nil pour XML propre, obtenu %v", err)
}
}