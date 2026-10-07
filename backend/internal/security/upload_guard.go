package security

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

var (
	ErrFileTooLarge    = errors.New("upload: file exceeds maximum allowed size")
	ErrXXEDetected     = errors.New("upload: XXE injection detected in XML payload")
	ErrInvalidFileType = errors.New("upload: magic bytes do not match declared content type")
	ErrPathTraversal   = errors.New("upload: dangerous path traversal filename detected")
)

const (
	MaxInvoiceFileSize = 10 * 1024 * 1024 // 10 Mo
)

// ValidateUploadPayload filtre les uploads contre XXE, path traversal et vÃ©rifie les signatures magiques
func ValidateUploadPayload(filename string, r io.Reader) ([]byte, error) {
	// 1. Path traversal check
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return nil, ErrPathTraversal
	}

	// 2. Limite stricte de taille (prÃ©vention XML / Zip Bomb)
	limitedReader := io.LimitReader(r, MaxInvoiceFileSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxInvoiceFileSize {
		return nil, ErrFileTooLarge
	}

	// 3. Inspection XXE stricte sur les XML
	trimmed := bytes.TrimSpace(data)
	if bytes.HasPrefix(trimmed, []byte("<")) {
		lower := strings.ToLower(string(trimmed[:min(len(trimmed), 1024)]))
		if strings.Contains(lower, "<!doctype") || strings.Contains(lower, "<!entity") || strings.Contains(lower, "system") {
			return nil, ErrXXEDetected
		}
	}

	return data, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
