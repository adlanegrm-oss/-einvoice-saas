package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/auth"
	"github.com/adlanegrm-oss/einvoice-saas/internal/middleware"
	"github.com/adlanegrm-oss/einvoice-saas/internal/validator"
)

const (
	FolderReady = "factures"
	FolderDraft = "temporaire"

	StatusReady = "VALIDE_PRET_A_ENVOYER"
	StatusDraft = "BROUILLON_TEMPORAIRE_72H"

	// DraftTTL : durée de conservation des brouillons.
	DraftTTL = 72 * time.Hour

	maxFilesPerDeposit = 50
	maxFileSize        = 10 << 20 // 10 Mo par fichier
	maxRequestSize     = 60 << 20 // 60 Mo par dépôt
)

var (
	tenantRe   = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)
	unsafeName = regexp.MustCompile(`[^A-Za-z0-9._ -]+`)
	storedRe   = regexp.MustCompile(`^\d{8}-\d{6}-[0-9a-f]{8}-`)
	allowedExt = map[string]bool{".pdf": true, ".xml": true, ".edi": true, ".edifact": true}
)

// ArchiveHandler gère le dépôt, la liste et le téléchargement des documents.
// Chaque client ne voit que <racine>/<son tenant>/ ; l'administrateur voit tout.
type ArchiveHandler struct {
	Root      string
	Validator *validator.FileValidator
}

func NewArchiveHandler(root string) (*ArchiveHandler, error) {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, err
	}
	return &ArchiveHandler{Root: root, Validator: validator.NewFileValidator()}, nil
}

// ArchiveEntry est une ligne du registre d'archives.
type ArchiveEntry struct {
	Filename    string `json:"filename"`     // nom stocké (sert au téléchargement)
	DisplayName string `json:"display_name"` // nom d'origine, sans préfixe technique
	Folder      string `json:"folder"`       // factures | temporaire
	Date        string `json:"date"`
	Status      string `json:"status"`
	Tenant      string `json:"tenant"`
	Size        int64  `json:"size"`
	modTime     time.Time
}

func (h *ArchiveHandler) allTenants() []string {
	entries, err := os.ReadDir(h.Root)
	if err != nil {
		return nil
	}
	var tenants []string
	for _, e := range entries {
		if e.IsDir() && tenantRe.MatchString(e.Name()) {
			tenants = append(tenants, e.Name())
		}
	}
	return tenants
}

// List renvoie le registre des archives visibles par l'utilisateur.
func (h *ArchiveHandler) List(w http.ResponseWriter, r *http.Request) {
	c, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	tenants := []string{c.Tenant}
	if c.Role == auth.RoleAdmin {
		tenants = h.allTenants()
	}

	folders := []struct{ name, status string }{{FolderReady, StatusReady}, {FolderDraft, StatusDraft}}
	list := make([]ArchiveEntry, 0)
	for _, tenant := range tenants {
		if !tenantRe.MatchString(tenant) {
			continue
		}
		for _, f := range folders {
			entries, err := os.ReadDir(filepath.Join(h.Root, tenant, f.name))
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !e.Type().IsRegular() {
					continue
				}
				info, err := e.Info()
				if err != nil {
					continue
				}
				list = append(list, ArchiveEntry{
					Filename:    e.Name(),
					DisplayName: storedRe.ReplaceAllString(e.Name(), ""),
					Folder:      f.name,
					Date:        info.ModTime().Format("2006-01-02 15:04"),
					Status:      f.status,
					Tenant:      tenant,
					Size:        info.Size(),
					modTime:     info.ModTime(),
				})
			}
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].modTime.After(list[j].modTime) })
	writeJSON(w, http.StatusOK, list)
}

type depositResult struct {
	Filename string               `json:"filename"`
	StoredAs string               `json:"stored_as,omitempty"`
	Status   string               `json:"status"` // ACCEPTE | BROUILLON | REJETE
	Format   validator.FormatType `json:"format,omitempty"`
	SHA256   string               `json:"sha256,omitempty"`
	Errors   []string             `json:"errors,omitempty"`
	Warnings []string             `json:"warnings,omitempty"`
}

// Deposit valide puis archive les fichiers reçus (multipart, champ "files").
// target=draft : conservation 72 h sans exiger la validité ; sinon le fichier
// doit être valide pour être archivé.
func (h *ArchiveHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	c, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Requête invalide ou trop volumineuse (60 Mo maximum)")
		return
	}
	defer r.MultipartForm.RemoveAll()

	draft := false
	switch r.FormValue("target") {
	case "draft":
		draft = true
	case "", "ready":
	default:
		writeError(w, http.StatusBadRequest, "Paramètre target invalide (ready ou draft)")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "Aucun fichier reçu")
		return
	}
	if len(files) > maxFilesPerDeposit {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Dépôt limité à %d documents maximum", maxFilesPerDeposit))
		return
	}

	results := make([]depositResult, 0, len(files))
	accepted := 0
	for _, fh := range files {
		res := h.processFile(c.Tenant, fh, draft)
		if res.Status != "REJETE" {
			accepted++
		}
		results = append(results, res)
	}

	status, code, msg := "success", http.StatusOK, ""
	switch {
	case accepted == 0:
		status, code, msg = "rejected", http.StatusUnprocessableEntity, "Aucun document n'a été archivé : consultez les erreurs de contrôle."
	case accepted < len(results):
		status, msg = "partial", fmt.Sprintf("%d document(s) archivé(s), %d rejeté(s).", accepted, len(results)-accepted)
	case draft:
		msg = "Facture(s) conservée(s) dans l'archive temporaire pour un délai de 72h."
	default:
		msg = "Contrôle effectué et factures archivées avec succès."
	}
	writeJSON(w, code, map[string]any{"status": status, "message": msg, "results": results})
}

func (h *ArchiveHandler) processFile(tenant string, fh *multipart.FileHeader, draft bool) depositResult {
	name := sanitizeName(fh.Filename)
	res := depositResult{Filename: name}
	reject := func(msg string) depositResult {
		res.Status = "REJETE"
		res.Errors = append(res.Errors, msg)
		return res
	}

	if !allowedExt[strings.ToLower(filepath.Ext(name))] {
		return reject("Extension non autorisée (formats acceptés : .pdf, .xml, .edi)")
	}
	if fh.Size > maxFileSize {
		return reject("Fichier trop volumineux (10 Mo maximum)")
	}
	f, err := fh.Open()
	if err != nil {
		return reject("Lecture du fichier impossible")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxFileSize+1))
	if err != nil || len(data) > maxFileSize {
		return reject("Lecture du fichier impossible ou fichier trop volumineux")
	}

	vr := h.Validator.ValidateFileContent(data)
	sum := sha256.Sum256(data)
	res.Format, res.SHA256, res.Warnings = vr.Format, hex.EncodeToString(sum[:]), vr.Warnings

	folder, status, label := FolderReady, StatusReady, "ACCEPTE"
	if draft {
		folder, status, label = FolderDraft, StatusDraft, "BROUILLON"
		res.Warnings = append(res.Warnings, vr.Errors...) // un brouillon peut être incomplet
	} else if !vr.IsValid {
		res.Errors = vr.Errors
		res.Status = "REJETE"
		return res
	}

	dir := filepath.Join(h.Root, tenant, folder)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		slog.Error("création du dossier d'archive", "error", err)
		return reject("Archivage impossible (erreur serveur)")
	}
	stored := uniqueName(name)
	if err := writeNewFile(filepath.Join(dir, stored), data); err != nil {
		slog.Error("écriture de l'archive", "error", err)
		return reject("Archivage impossible (erreur serveur)")
	}

	res.Status, res.StoredAs = label, stored
	h.journal(tenant, fmt.Sprintf("[%s] Fichier: %q | Stocké: %q | Statut: %s | Format: %s | SHA256: %s\n",
		time.Now().Format("2006-01-02 15:04:05"), name, stored, status, vr.Format, res.SHA256))
	slog.Info("document archivé", "tenant", tenant, "stored_as", stored, "status", status, "format", vr.Format)
	return res
}

func (h *ArchiveHandler) journal(tenant, line string) {
	f, err := os.OpenFile(filepath.Join(h.Root, tenant, "traitement.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		slog.Error("journal de traitement", "error", err)
		return
	}
	defer f.Close()
	f.WriteString(line)
}

// Download renvoie un fichier archivé. Le client ne peut lire que son propre tenant ;
// l'administrateur peut préciser ?tenant=.
func (h *ArchiveHandler) Download(w http.ResponseWriter, r *http.Request) {
	c, ok := middleware.ClaimsFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	q := r.URL.Query()
	name, folder := q.Get("file"), q.Get("folder")

	if folder != FolderReady && folder != FolderDraft {
		writeError(w, http.StatusBadRequest, "Dossier invalide")
		return
	}
	tenant := c.Tenant
	if c.Role == auth.RoleAdmin && q.Get("tenant") != "" {
		tenant = q.Get("tenant")
	}
	if !tenantRe.MatchString(tenant) {
		writeError(w, http.StatusBadRequest, "Tenant invalide")
		return
	}
	if name == "" || name == "." || name == ".." || name != filepath.Base(name) || strings.ContainsAny(name, `/\`) {
		writeError(w, http.StatusBadRequest, "Nom de fichier invalide")
		return
	}

	f, err := os.Open(filepath.Join(h.Root, tenant, folder, name))
	if err != nil {
		writeError(w, http.StatusNotFound, "Document introuvable")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		writeError(w, http.StatusNotFound, "Document introuvable")
		return
	}

	ctype := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf":
		ctype = "application/pdf"
	case ".xml":
		ctype = "application/xml"
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": storedRe.ReplaceAllString(name, "")}))
	http.ServeContent(w, r, name, info.ModTime(), f)
}

// PurgeExpiredDrafts supprime les brouillons plus vieux que maxAge et renvoie leur nombre.
func (h *ArchiveHandler) PurgeExpiredDrafts(now time.Time, maxAge time.Duration) int {
	removed := 0
	for _, tenant := range h.allTenants() {
		dir := filepath.Join(h.Root, tenant, FolderDraft)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.Type().IsRegular() {
				continue
			}
			info, err := e.Info()
			if err != nil || now.Sub(info.ModTime()) <= maxAge {
				continue
			}
			if err := os.Remove(filepath.Join(dir, e.Name())); err == nil {
				removed++
			}
		}
	}
	return removed
}

// RunPurger purge les brouillons expirés au démarrage puis toutes les heures.
func (h *ArchiveHandler) RunPurger(ctx context.Context) {
	tick := time.NewTicker(time.Hour)
	defer tick.Stop()
	for {
		if n := h.PurgeExpiredDrafts(time.Now(), DraftTTL); n > 0 {
			slog.Info("brouillons expirés supprimés", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

// sanitizeName ne garde que le nom de base et des caractères sûrs.
func sanitizeName(name string) string {
	base := filepath.Base(strings.ReplaceAll(name, `\`, "/"))
	base = unsafeName.ReplaceAllString(base, "_")
	base = strings.TrimLeft(base, ". ")
	if base == "" {
		return "document"
	}
	if len(base) > 100 {
		ext := filepath.Ext(base)
		if len(ext) > 10 {
			ext = ""
		}
		base = base[:100-len(ext)] + ext
	}
	return base
}

// uniqueName préfixe le nom (horodatage + aléa) pour ne jamais écraser un document.
func uniqueName(name string) string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%s-%s-%s", time.Now().UTC().Format("20060102-150405"), hex.EncodeToString(b), name)
}

// writeNewFile crée le fichier en exclusivité : échoue s'il existe déjà.
func writeNewFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return err
	}
	return nil
}
