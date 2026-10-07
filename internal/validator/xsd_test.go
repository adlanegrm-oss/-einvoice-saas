package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestXSDValidator_Validate(t *testing.T) {
	tmpDir := t.TempDir()
	dummySchema := filepath.Join(tmpDir, "schema.xsd")
	if err := os.WriteFile(dummySchema, []byte("<xs:schema xmlns:xs=\"http://www.w3.org/2001/XMLSchema\"/>"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewXSDValidator(dummySchema)

	// 1. Rejet XML vide
	if err := v.Validate([]byte("")); err == nil {
		t.Fatal("attendu: erreur sur XML vide")
	}

	// 2. Erreur sur schéma inexistant
	vMissing := NewXSDValidator(filepath.Join(tmpDir, "absent.xsd"))
	if err := vMissing.Validate([]byte("<test/>")); err == nil {
		t.Fatal("attendu: erreur schéma introuvable")
	}

	// 3. Succès sur fichier présent et payload non vide
	if err := v.Validate([]byte("<test/>")); err != nil {
		t.Fatalf("validation inattendue en échec: %v", err)
	}
}
