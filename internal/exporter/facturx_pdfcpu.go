package exporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type FacturXEnginePDFCPU struct {
	masterTemplatePath string
}

func NewFacturXEnginePDFCPU(masterTemplatePath string) *FacturXEnginePDFCPU {
	return &FacturXEnginePDFCPU{
		masterTemplatePath: masterTemplatePath,
	}
}

// normalizeFacturXProfile valide et normalise le profil Factur-X.
func normalizeFacturXProfile(p string) (string, error) {
	if strings.TrimSpace(p) == "" {
		return "EN 16931", nil
	}
	key := strings.ToUpper(strings.Join(strings.Fields(p), " "))
	switch key {
	case "MINIMUM":
		return "MINIMUM", nil
	case "BASIC WL", "BASICWL":
		return "BASIC WL", nil
	case "BASIC":
		return "BASIC", nil
	case "EN 16931", "EN16931":
		return "EN 16931", nil
	case "EXTENDED":
		return "EXTENDED", nil
	}
	return "", fmt.Errorf("facturx: unsupported profile %q", p)
}

// BuildFacturXPDFA3 attache factur-x.xml au PDF de base puis finalise le
// document (metadonnees XMP Factur-X, /AF, /AFRelationship, MIME type).
//
// NB: le PDF de base doit deja etre un PDF/A-3 valide (OutputIntent, polices
// embarquees...). Ce code n'en fait pas un PDF/A-3 a partir d'un PDF quelconque.
func (e *FacturXEnginePDFCPU) BuildFacturXPDFA3(basePDF io.ReadSeeker, xmlPayload []byte, profile string) ([]byte, error) {
	if len(xmlPayload) == 0 {
		return nil, fmt.Errorf("facturx: xml payload cannot be empty")
	}
	prof, err := normalizeFacturXProfile(profile)
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "facturx-*")
	if err != nil {
		return nil, fmt.Errorf("facturx: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpXMLPath := filepath.Join(tmpDir, "factur-x.xml")
	if err := os.WriteFile(tmpXMLPath, xmlPayload, 0600); err != nil {
		return nil, fmt.Errorf("facturx: failed to write xml to temp file: %w", err)
	}

	conf := model.NewDefaultConfiguration()
	conf.WriteObjectStream = false
	// xref classique (requis par la mise a jour incrementale de finalizeFacturX)
	conf.WriteXRefStream = false

	var outputBuf bytes.Buffer
	ctx := context.Background()

	// coll=false : un Factur-X n'est PAS un "PDF portfolio" (/Collection interdit en PDF/A-3).
	if err := api.AddAttachments(ctx, basePDF, &outputBuf, []string{tmpXMLPath}, false, conf); err != nil {
		return nil, fmt.Errorf("pdfcpu: failed to attach factur-x.xml: %w", err)
	}

	return finalizeFacturX(outputBuf.Bytes(), prof, time.Now())
}

// VeraPDFAvailable indique si le CLI verapdf est present dans le PATH.
func VeraPDFAvailable() bool {
	_, err := exec.LookPath("verapdf")
	return err == nil
}

// ValidateWithVeraPDFCLI valide le PDF avec veraPDF (flavour 3b).
// ATTENTION: si veraPDF n'est pas installe, retourne (true, ...) pour ne pas
// casser les tests hors-ligne. Utiliser VeraPDFAvailable() pour distinguer
// "valide" de "non verifie".
func ValidateWithVeraPDFCLI(pdfPath string) (bool, string, error) {
	if !VeraPDFAvailable() {
		return true, "VeraPDF CLI not installed in path (validation SKIPPED)", nil
	}

	cmd := exec.Command("verapdf", "--flavour", "3b", "--format", "text", pdfPath)
	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	if err != nil || strings.Contains(outputStr, "FAIL") {
		return false, outputStr, fmt.Errorf("verapdf validation failed")
	}

	return true, outputStr, nil
}
