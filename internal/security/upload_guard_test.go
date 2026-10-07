package security

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

func TestValidateUploadPayload_PathTraversal(t *testing.T) {
	for _, name := range []string{
		"../etc/passwd",
		"a/b.xml",
		`a\b.xml`,
		"..",
		"dossier/../x.xml",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateUploadPayload(
				name,
				strings.NewReader("<Invoice/>"),
			)

			if !errors.Is(err, ErrPathTraversal) {
				t.Fatalf(
					"ErrPathTraversal attendu, obtenu %v",
					err,
				)
			}
		})
	}
}

func TestValidateUploadPayload_Size(t *testing.T) {
	t.Run("exactement la limite", func(t *testing.T) {
		data := bytes.Repeat(
			[]byte("a"),
			MaxInvoiceFileSize,
		)

		got, err := ValidateUploadPayload(
			"f.bin",
			bytes.NewReader(data),
		)

		if err != nil {
			t.Fatalf("erreur inattendue : %v", err)
		}

		if len(got) != MaxInvoiceFileSize {
			t.Fatalf(
				"taille lue = %d",
				len(got),
			)
		}
	})

	t.Run("limite + 1", func(t *testing.T) {
		data := bytes.Repeat(
			[]byte("a"),
			MaxInvoiceFileSize+1,
		)

		_, err := ValidateUploadPayload(
			"f.bin",
			bytes.NewReader(data),
		)

		if !errors.Is(err, ErrFileTooLarge) {
			t.Fatalf(
				"ErrFileTooLarge attendu, obtenu %v",
				err,
			)
		}
	})
}

func TestValidateUploadPayload_ReaderError(t *testing.T) {
	boom := errors.New("lecture impossible")

	_, err := ValidateUploadPayload(
		"f.xml",
		iotest.ErrReader(boom),
	)

	if !errors.Is(err, boom) {
		t.Fatalf(
			"l'erreur du reader doit etre propagee, obtenu %v",
			err,
		)
	}
}

func TestValidateUploadPayload_XXE(t *testing.T) {
	tests := map[string]string{
		"doctype":         `<?xml version="1.0"?><!DOCTYPE foo [<!ELEMENT foo ANY>]><foo/>`,
		"entity":          `<?xml version="1.0"?><!ENTITY xxe "x"><foo/>`,
		"system":          `<foo xmlns:x="SYSTEM">a</foo>`,
		"casse mixte":     `<?xml version="1.0"?><!DocType foo><foo/>`,
		"espaces en tete": "   \n\t<?xml version=\"1.0\"?><!DOCTYPE foo><foo/>",
		"entite externe":  `<!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><foo>&xxe;</foo>`,
	}

	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateUploadPayload(
				"f.xml",
				strings.NewReader(payload),
			)

			if !errors.Is(err, ErrXXEDetected) {
				t.Fatalf(
					"ErrXXEDetected attendu, obtenu %v",
					err,
				)
			}
		})
	}
}

func TestValidateUploadPayload_ValidPayloads(t *testing.T) {
	tests := map[string][]byte{
		"xml simple": []byte(
			`<?xml version="1.0"?><Invoice><ID>1</ID></Invoice>`,
		),
		"xml court": []byte(`<a/>`),
		"pdf":       []byte("%PDF-1.7\n%....\n"),
		"vide":      {},
	}

	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ValidateUploadPayload(
				"facture.xml",
				bytes.NewReader(payload),
			)

			if err != nil {
				t.Fatalf(
					"erreur inattendue : %v",
					err,
				)
			}

			if !bytes.Equal(got, payload) {
				t.Fatal(
					"les donnees retournees different de l'entree",
				)
			}
		})
	}
}

func TestValidateUploadPayload_NonXMLIsNotInspected(t *testing.T) {
	payload := []byte(
		"texte contenant <!DOCTYPE et SYSTEM",
	)

	if _, err := ValidateUploadPayload(
		"f.txt",
		bytes.NewReader(payload),
	); err != nil {
		t.Fatalf(
			"erreur inattendue : %v",
			err,
		)
	}
}

func TestMin(t *testing.T) {
	if min(1, 2) != 1 ||
		min(2, 1) != 1 ||
		min(3, 3) != 3 ||
		min(-1, 0) != -1 {
		t.Fatal("min() retourne une valeur incorrecte")
	}
}
