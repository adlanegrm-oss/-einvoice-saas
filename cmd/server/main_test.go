package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
	"github.com/adlanegrm-oss/einvoice-saas/internal/config"
	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
)

const (
	adminEmail   = "admin@example.com"
	adminPass    = "admin-password-123"
	clientEmail  = "client@example.com"
	clientPass   = "client-password-123"
	client2Email = "autre@example.com"
	client2Pass  = "autre-password-123"
)

const validUBL = `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
         xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
         xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:ID>F-2026-001</cbc:ID>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cac:AccountingSupplierParty><cac:Party><cac:PartyName><cbc:Name>Vendeur SA</cbc:Name></cac:PartyName></cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party><cac:PartyName><cbc:Name>Acheteur SARL</cbc:Name></cac:PartyName></cac:Party></cac:AccountingCustomerParty>
</Invoice>`

type captureNotifier struct {
	mu    sync.Mutex
	links []string
}

func (c *captureNotifier) SendResetLink(email, link string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.links = append(c.links, link)
	return nil
}

func (c *captureNotifier) last() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.links) == 0 {
		return ""
	}
	return c.links[len(c.links)-1]
}

type env struct {
	app      *app
	srv      *httptest.Server
	notifier *captureNotifier
	archives string
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dir := t.TempDir()
	web := filepath.Join(dir, "web")
	if err := os.MkdirAll(web, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(web, "index.html"), []byte("<html>ok</html>"), 0o644)

	cfg := &config.Config{
		AppEnv: config.EnvDev, Port: "0", DataDir: dir,
		DBPath:     filepath.Join(dir, "test.db"),
		ArchiveDir: filepath.Join(dir, "archives"),
		WebDir:     web, PublicBaseURL: "http://test.local",
		JWTSecret: strings.Repeat("s", 40), TokenTTL: time.Hour,
		AdminEmail: adminEmail, AdminPassword: adminPass,
		ClientEmail: clientEmail, ClientPassword: clientPass,
	}
	n := &captureNotifier{}
	a, err := newApp(cfg, n)
	if err != nil {
		t.Fatalf("newApp : %v", err)
	}
	if _, err := a.store.AddUser(client2Email, client2Pass, auth.RoleClient); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(a.Handler)
	t.Cleanup(func() { srv.Close(); a.Close() })
	return &env{app: a, srv: srv, notifier: n, archives: cfg.ArchiveDir}
}

func (e *env) do(t *testing.T, method, path, token, contentType string, body io.Reader) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

func (e *env) postJSON(t *testing.T, path, token string, payload any) (int, []byte) {
	t.Helper()
	b, _ := json.Marshal(payload)
	return e.do(t, "POST", path, token, "application/json", bytes.NewReader(b))
}

func (e *env) login(t *testing.T, email, password string) string {
	t.Helper()
	code, body := e.postJSON(t, "/api/v1/auth/login", "", map[string]string{"email": email, "password": password})
	if code != 200 {
		t.Fatalf("connexion de %s : code %d, corps %s", email, code, body)
	}
	var out struct{ Token string }
	json.Unmarshal(body, &out)
	if out.Token == "" {
		t.Fatal("jeton vide")
	}
	return out.Token
}

func multipartBody(t *testing.T, target string, files map[string]string) (string, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("target", target)
	for name, content := range files {
		fw, _ := w.CreateFormFile("files", name)
		fw.Write([]byte(content))
	}
	w.Close()
	return w.FormDataContentType(), &buf
}

func TestHealthAndStatic(t *testing.T) {
	e := newEnv(t)
	if code, _ := e.do(t, "GET", "/health", "", "", nil); code != 200 {
		t.Errorf("/health : %d", code)
	}
	if code, body := e.do(t, "GET", "/", "", "", nil); code != 200 || !strings.Contains(string(body), "ok") {
		t.Errorf("page statique : %d %s", code, body)
	}
}

func TestAPIRequiresAuthentication(t *testing.T) {
	e := newEnv(t)
	routes := [][2]string{
		{"GET", "/api/v1/invoices/list"}, {"GET", "/api/v1/invoices"},
		{"GET", "/api/v1/invoices/download?file=a.pdf&folder=factures"},
		{"GET", "/api/v1/reports/daily"}, {"GET", "/api/v1/auth/me"},
		{"POST", "/api/v1/invoices/deposit"}, {"POST", "/api/v1/validate"}, {"POST", "/api/v1/invoices"},
	}
	for _, r := range routes {
		if code, _ := e.do(t, r[0], r[1], "", "", nil); code != 401 {
			t.Errorf("%s %s sans jeton : attendu 401, reÃ§u %d", r[0], r[1], code)
		}
	}

	// Ancienne faille : l'en-tÃªte X-User-Role fourni par le client donnait l'accÃ¨s.
	req, _ := http.NewRequest("GET", e.srv.URL+"/api/v1/invoices/list", nil)
	req.Header.Set("X-User-Role", "EXPLOITATION")
	req.Header.Set("Authorization", "ADMIN")
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("en-tÃªtes de rÃ´le forgÃ©s : attendu 401, reÃ§u %d", resp.StatusCode)
	}

	// route d'administration refusÃ©e Ã  un client
	client := e.login(t, clientEmail, clientPass)
	if code, _ := e.do(t, "POST", "/api/v1/jobs/daily-report", client, "", nil); code != 403 {
		t.Errorf("route admin avec jeton client : attendu 403, reÃ§u %d", code)
	}
	admin := e.login(t, adminEmail, adminPass)
	if code, _ := e.do(t, "POST", "/api/v1/jobs/daily-report", admin, "", nil); code != 202 {
		t.Errorf("route admin avec jeton admin : attendu 202, reÃ§u %d", code)
	}
}

func TestLogin(t *testing.T) {
	e := newEnv(t)
	if code, _ := e.postJSON(t, "/api/v1/auth/login", "", map[string]string{"email": adminEmail, "password": "faux"}); code != 401 {
		t.Errorf("mauvais mot de passe : attendu 401, reÃ§u %d", code)
	}
	// les anciens identifiants codÃ©s en dur ne doivent plus fonctionner
	if code, _ := e.postJSON(t, "/api/v1/auth/login", "", map[string]string{"email": "admin.super@einvoice.int", "password": "hadahowana"}); code != 401 {
		t.Errorf("ancien identifiant : attendu 401, reÃ§u %d", code)
	}
	tok := e.login(t, clientEmail, clientPass)
	code, body := e.do(t, "GET", "/api/v1/auth/me", tok, "", nil)
	if code != 200 || !strings.Contains(string(body), "CLIENT_DO") {
		t.Errorf("/auth/me : %d %s", code, body)
	}
}

func TestPasswordResetFlow(t *testing.T) {
	e := newEnv(t)

	code, known := e.postJSON(t, "/api/v1/auth/forgot-password", "", map[string]string{"email": clientEmail})
	if code != 200 {
		t.Fatalf("forgot-password : %d", code)
	}
	link := e.notifier.last()
	if link == "" {
		t.Fatal("aucun lien transmis au canal de notification")
	}
	u, err := url.Parse(link)
	if err != nil || u.Query().Get("token") == "" {
		t.Fatalf("lien de rÃ©initialisation invalide : %q", link)
	}
	token := u.Query().Get("token")
	if strings.Contains(string(known), token) {
		t.Fatal("FAILLE : le jeton est renvoyÃ© dans la rÃ©ponse HTTP")
	}

	// un compte inconnu reÃ§oit exactement la mÃªme rÃ©ponse (pas d'Ã©numÃ©ration)
	_, unknown := e.postJSON(t, "/api/v1/auth/forgot-password", "", map[string]string{"email": "inconnu@example.com"})
	if string(known) != string(unknown) {
		t.Errorf("rÃ©ponses diffÃ©rentes selon l'existence du compte :\n%s\n%s", known, unknown)
	}

	// mot de passe faible refusÃ©, puis rÃ©initialisation valide
	if code, _ := e.postJSON(t, "/api/v1/auth/reset-password", "", map[string]string{"token": token, "password": "court"}); code != 400 {
		t.Errorf("mot de passe faible : attendu 400, reÃ§u %d", code)
	}
	if code, _ := e.postJSON(t, "/api/v1/auth/reset-password", "", map[string]string{"token": token, "password": "nouveau-mot-de-passe"}); code != 200 {
		t.Fatalf("rÃ©initialisation valide : %d", code)
	}
	e.login(t, clientEmail, "nouveau-mot-de-passe")
	if code, _ := e.postJSON(t, "/api/v1/auth/login", "", map[string]string{"email": clientEmail, "password": clientPass}); code != 401 {
		t.Error("l'ancien mot de passe fonctionne encore")
	}
	// jeton Ã  usage unique
	if code, _ := e.postJSON(t, "/api/v1/auth/reset-password", "", map[string]string{"token": token, "password": "encore-un-autre-mdp"}); code != 400 {
		t.Errorf("rÃ©utilisation du jeton : attendu 400, reÃ§u %d", code)
	}
	// impossible de crÃ©er un compte via un faux jeton
	if code, _ := e.postJSON(t, "/api/v1/auth/reset-password", "", map[string]string{"token": "n-importe-quoi", "password": "mot-de-passe-ok-123"}); code != 400 {
		t.Errorf("faux jeton : attendu 400, reÃ§u %d", code)
	}
}

func TestDepositListDownloadIsolation(t *testing.T) {
	e := newEnv(t)
	c1 := e.login(t, clientEmail, clientPass)
	c2 := e.login(t, client2Email, client2Pass)
	admin := e.login(t, adminEmail, adminPass)

	// dÃ©pÃ´t valide
	ct, body := multipartBody(t, "ready", map[string]string{"facture.xml": validUBL})
	code, resp := e.do(t, "POST", "/api/v1/invoices/deposit", c1, ct, body)
	if code != 200 {
		t.Fatalf("dÃ©pÃ´t valide : %d %s", code, resp)
	}
	var dep struct {
		Results []struct {
			StoredAs string `json:"stored_as"`
			SHA256   string `json:"sha256"`
		}
	}
	json.Unmarshal(resp, &dep)
	if len(dep.Results) != 1 || dep.Results[0].StoredAs == "" || len(dep.Results[0].SHA256) != 64 {
		t.Fatalf("rÃ©sultat de dÃ©pÃ´t inattendu : %s", resp)
	}
	stored := dep.Results[0].StoredAs

	// XML invalide en mode "ready" : rejetÃ© et non archivÃ©
	ct, body = multipartBody(t, "ready", map[string]string{"casse.xml": "<Invoice><a></Invoice>"})
	if code, _ := e.do(t, "POST", "/api/v1/invoices/deposit", c1, ct, body); code != 422 {
		t.Errorf("XML invalide : attendu 422, reÃ§u %d", code)
	}
	// extension interdite
	ct, body = multipartBody(t, "ready", map[string]string{"shell.html": "<script>alert(1)</script>"})
	if code, _ := e.do(t, "POST", "/api/v1/invoices/deposit", c1, ct, body); code != 422 {
		t.Errorf("extension interdite : attendu 422, reÃ§u %d", code)
	}
	// le mÃªme XML invalide est acceptÃ© en brouillon
	ct, body = multipartBody(t, "draft", map[string]string{"brouillon.xml": "<Invoice><a></Invoice>"})
	if code, _ := e.do(t, "POST", "/api/v1/invoices/deposit", c1, ct, body); code != 200 {
		t.Errorf("brouillon : attendu 200, reÃ§u %d", code)
	}

	// listes : c1 voit ses 2 documents, c2 rien, l'admin tout
	count := func(token string) int {
		_, b := e.do(t, "GET", "/api/v1/invoices/list", token, "", nil)
		var list []map[string]any
		json.Unmarshal(b, &list)
		return len(list)
	}
	if n := count(c1); n != 2 {
		t.Errorf("client 1 : 2 documents attendus, %d reÃ§us", n)
	}
	if n := count(c2); n != 0 {
		t.Errorf("client 2 ne doit rien voir, %d reÃ§us", n)
	}
	if n := count(admin); n != 2 {
		t.Errorf("admin : 2 documents attendus, %d reÃ§us", n)
	}

	// tÃ©lÃ©chargement : propriÃ©taire OK, autre client refusÃ©, y compris avec ?tenant=
	dl := "/api/v1/invoices/download?folder=factures&file=" + url.QueryEscape(stored)
	code, data := e.do(t, "GET", dl, c1, "", nil)
	if code != 200 || string(data) != validUBL {
		t.Errorf("tÃ©lÃ©chargement par le propriÃ©taire : %d", code)
	}
	if code, _ := e.do(t, "GET", dl, c2, "", nil); code != 404 {
		t.Errorf("tÃ©lÃ©chargement par un autre client : attendu 404, reÃ§u %d", code)
	}
	entries, _ := os.ReadDir(e.archives)
	for _, d := range entries {
		if code, _ := e.do(t, "GET", dl+"&tenant="+d.Name(), c2, "", nil); code != 404 {
			t.Errorf("paramÃ¨tre tenant exploitÃ© par un client : attendu 404, reÃ§u %d", code)
		}
	}

	// traversÃ©e de rÃ©pertoire
	for _, bad := range []string{"../../etc/passwd", "..%2f..%2fetc%2fpasswd", "..", "a/b.pdf", `..\..\x`} {
		if code, _ := e.do(t, "GET", "/api/v1/invoices/download?folder=factures&file="+url.QueryEscape(bad), c1, "", nil); code == 200 {
			t.Errorf("traversÃ©e acceptÃ©e pour %q", bad)
		}
	}
	if code, _ := e.do(t, "GET", "/api/v1/invoices/download?folder=../x&file="+url.QueryEscape(stored), c1, "", nil); code != 400 {
		t.Errorf("dossier invalide : attendu 400, reÃ§u %d", code)
	}
}

func TestStructuredInvoices(t *testing.T) {
	e := newEnv(t)
	c1 := e.login(t, clientEmail, clientPass)
	c2 := e.login(t, client2Email, client2Pass)

	inv := map[string]any{
		"id": "choisi-par-le-client", "number": "FAC-2026-001", "issue_date": "2026-10-03T10:00:00Z", "customer": map[string]any{"name": "ACME"},
		"items": []map[string]any{{"description": "Prestation", "quantity": 2, "unit_price": 100.0, "vat_rate": 20.0}},
	}
	code, body := e.postJSON(t, "/api/v1/invoices", c1, inv)
	if code != 201 {
		t.Fatalf("crÃ©ation : %d %s", code, body)
	}
	var created struct {
		Invoice struct {
			ID       string  `json:"id"`
			TotalTTC float64 `json:"total_ttc"`
		}
	}
	json.Unmarshal(body, &created)
	if created.Invoice.ID == "choisi-par-le-client" || created.Invoice.ID == "" {
		t.Errorf("l'identifiant doit Ãªtre attribuÃ© par le serveur : %q", created.Invoice.ID)
	}
	if created.Invoice.TotalTTC != 240 {
		t.Errorf("TTC attendu 240, reÃ§u %v", created.Invoice.TotalTTC)
	}

	if code, _ := e.postJSON(t, "/api/v1/invoices", c1, inv); code != 409 {
		t.Errorf("numÃ©ro en double : attendu 409, reÃ§u %d", code)
	}
	if code, _ := e.postJSON(t, "/api/v1/invoices", c1, map[string]any{"number": "X", "customer": "Y", "items": []any{}}); code != 422 {
		t.Errorf("facture sans ligne : attendu 422, reÃ§u %d", code)
	}

	list := func(token string) int {
		_, b := e.do(t, "GET", "/api/v1/invoices", token, "", nil)
		var l []map[string]any
		json.Unmarshal(b, &l)
		return len(l)
	}
	if list(c1) != 1 || list(c2) != 0 {
		t.Error("l'isolation des factures entre clients est brisÃ©e")
	}

	exp := "/api/v1/invoices/export?id=" + url.QueryEscape(created.Invoice.ID)
	if code, b := e.do(t, "GET", exp, c1, "", nil); code != 200 || !strings.Contains(string(b), "CrossIndustryInvoice") {
		t.Errorf("export par le propriÃ©taire : %d", code)
	}
	if code, _ := e.do(t, "GET", exp, c2, "", nil); code != 404 {
		t.Errorf("export par un autre client : attendu 404, reÃ§u %d", code)
	}

	if code, _ := e.do(t, "GET", "/api/v1/reports/daily?date=pas-une-date", c1, "", nil); code != 400 {
		t.Errorf("date invalide : attendu 400, reÃ§u %d", code)
	}
	today := time.Now().UTC().Format("2006-01-02")
	code, b := e.do(t, "GET", "/api/v1/reports/daily?date="+today, c1, "", nil)
	var rep struct {
		TotalInvoices int `json:"total_invoices"`
	}
	json.Unmarshal(b, &rep)
	if code != 200 || rep.TotalInvoices != 1 {
		t.Errorf("rapport du jour : %d %s", code, b)
	}
}

func TestPurgeExpiredDrafts(t *testing.T) {
	e := newEnv(t)
	c1 := e.login(t, clientEmail, clientPass)
	ct, body := multipartBody(t, "draft", map[string]string{"vieux.xml": "<a/>"})
	if code, _ := e.do(t, "POST", "/api/v1/invoices/deposit", c1, ct, body); code != 200 {
		t.Fatalf("dÃ©pÃ´t du brouillon : %d", code)
	}
	ct, body = multipartBody(t, "draft", map[string]string{"recent.xml": "<a/>"})
	if code, _ := e.do(t, "POST", "/api/v1/invoices/deposit", c1, ct, body); code != 200 {
		t.Fatalf("dÃ©pÃ´t du brouillon rÃ©cent : %d", code)
	}

	find := func(suffix string) string {
		var found string
		filepath.Walk(e.archives, func(p string, info os.FileInfo, err error) error {
			if err == nil && info.Mode().IsRegular() && strings.HasSuffix(p, suffix) {
				found = p
			}
			return nil
		})
		return found
	}
	oldPath, newPath := find("vieux.xml"), find("recent.xml")
	if oldPath == "" || newPath == "" {
		t.Fatal("brouillons introuvables")
	}
	old := time.Now().Add(-73 * time.Hour)
	if err := os.Chtimes(oldPath, old, old); err != nil {
		t.Fatal(err)
	}

	if n := e.app.archive.PurgeExpiredDrafts(time.Now(), handler.DraftTTL); n != 1 {
		t.Errorf("1 brouillon expirÃ© attendu, %d supprimÃ©(s)", n)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("le brouillon de plus de 72 h doit Ãªtre supprimÃ©")
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Error("le brouillon rÃ©cent doit Ãªtre conservÃ©")
	}
}
