package parser

import (
"strings"
"testing"
)

func TestHardenedXMLReader_ValidXML(t *testing.T) {
validXML := `<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"><ID>INV-2026-001</ID></Invoice>`
res, err := HardenedXMLReader(strings.NewReader(validXML))
if err != nil {
t.Fatalf("expected valid XML to pass, got: %v", err)
}
if len(res) == 0 {
t.Fatal("expected non-empty output")
}
}

func TestHardenedXMLReader_RejectXXE(t *testing.T) {
maliciousPayloads := []string{
`<!DOCTYPE foo [ <!ENTITY xxe SYSTEM "file:///etc/passwd"> ]><Invoice><ID>&xxe;</ID></Invoice>`,
`<!DOCTYPE foo [ <!ENTITY xxe SYSTEM "http://malicious.local/evil"> ]><Invoice><ID>&xxe;</ID></Invoice>`,
`<!DOCTYPE lolz [ <!ENTITY lol "lol"><!ELEMENT lolz (#PCDATA)><!ENTITY lol1 "&lol;&lol;"> ]><lolz>&lol1;</lolz>`,
}

for _, payload := range maliciousPayloads {
_, err := HardenedXMLReader(strings.NewReader(payload))
if err == nil {
t.Errorf("expected security rejection for payload: %s, but got nil", payload)
}
}
}

func TestHardenedXMLReader_RejectOversized(t *testing.T) {
hugeData := strings.Repeat("A", MaxInvoiceXMLSize+1024)
_, err := HardenedXMLReader(strings.NewReader(hugeData))
if err != ErrPayloadTooLarge {
t.Fatalf("expected ErrPayloadTooLarge, got: %v", err)
}
}
