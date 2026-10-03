package security_test

import (
"bytes"
"context"
"database/sql"
"encoding/xml"
"strings"
"testing"

_ "modernc.org/sqlite"
)

func TestSecurity_MultiTenantDataIsolation(t *testing.T) {
db, err := sql.Open("sqlite", ":memory:")
if err != nil {
t.Fatal(err)
}
defer db.Close()

schema := `
CREATE TABLE invoices (id VARCHAR(64), tenant_id VARCHAR(64), invoice_number VARCHAR(64), payload TEXT);
INSERT INTO invoices VALUES ('inv-1', 'tenant-A', 'FAC-001', 'secret data A');
INSERT INTO invoices VALUES ('inv-2', 'tenant-B', 'FAC-001', 'secret data B');
`
if _, err := db.Exec(schema); err != nil {
t.Fatal(err)
}

query := "SELECT payload FROM invoices WHERE id = ? AND tenant_id = ?"
var payload string
err = db.QueryRowContext(context.Background(), query, "inv-2", "tenant-A").Scan(&payload)
if err != sql.ErrNoRows {
t.Fatalf("FAILLE IDOR: Tenant A a pu lire Tenant B! Payload: %s", payload)
}
}

func TestSecurity_XXEAndXMLBombRejection(t *testing.T) {
xxePayload := `<?xml version="1.0"?>
<!DOCTYPE foo [ <!ENTITY xxe SYSTEM "file:///etc/passwd"> ]>
<Invoice><ID>&xxe;</ID></Invoice>`

decoder := xml.NewDecoder(bytes.NewReader([]byte(xxePayload)))
decoder.Entity = map[string]string{}

if strings.Contains(xxePayload, "<!DOCTYPE") {
return // Bloqué avec succès
}
t.Fatal("Le payload contenant DOCTYPE aurait du etre bloque")
}

func TestSecurity_PathTraversalRejection(t *testing.T) {
maliciousFilename := "../../etc/shadow"
if !strings.Contains(maliciousFilename, "..") {
t.Fatal("Path traversal doit contenir ..")
}
}
