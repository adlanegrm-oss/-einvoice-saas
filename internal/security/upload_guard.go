package security

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

var (
	ErrUploadEmpty    = errors.New("security: empty upload")
	ErrUploadTooLarge = errors.New("security: upload too large")
	ErrUploadXXE      = errors.New("security: XXE / DTD interdits")
	ErrUploadType     = errors.New("security: media type not allowed")
)

const DefaultMaxUploadBytes = 5 << 20 // 5 MiB

// ValidateXMLUpload contrÃ´les basiques avant parse XML.
func ValidateXMLUpload(data []byte, maxBytes int) error {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxUploadBytes
	}
	if len(data) == 0 {
		return ErrUploadEmpty
	}
	if len(data) > maxBytes {
		return ErrUploadTooLarge
	}
	// Refuse DOCTYPE / ENTITY (XXE)
	lower := bytes.ToLower(data)
	if bytes.Contains(lower, []byte("<!doctype")) || bytes.Contains(lower, []byte("<!entity")) {
		return ErrUploadXXE
	}
	if !utf8.Valid(data) {
		return errors.New("security: invalid utf-8")
	}
	trim := bytes.TrimSpace(data)
	if !bytes.HasPrefix(trim, []byte("<")) {
		return ErrUploadType
	}
	return nil
}

// AllowedFilename empÃªche path traversal dans les noms de piÃ¨ces jointes.
func AllowedFilename(name string) bool {
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return false
	}
	return true
}

// API de validation des fichiers de facture.
var (
	ErrPathTraversal = errors.New("security: path traversal interdit")
	ErrFileTooLarge  = errors.New("security: fichier trop volumineux")
	ErrXXEDetected   = errors.New("security: marqueur XXE dÃ©tectÃ©")
)

const MaxInvoiceFileSize = 10 << 20 // 10 MiB

// ValidateUploadPayload lit le fichier avec une limite stricte de taille.
func ValidateUploadPayload(filename string, reader io.Reader) ([]byte, error) {
	if !AllowedFilename(filename) {
		return nil, ErrPathTraversal
	}
	if reader == nil {
		return nil, errors.New("security: reader nil")
	}

	data, err := io.ReadAll(io.LimitReader(reader, int64(MaxInvoiceFileSize)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxInvoiceFileSize {
		return nil, ErrFileTooLarge
	}

	if strings.EqualFold(strings.TrimSpace(filename[strings.LastIndex(filename, ".")+1:]), "xml") &&
		strings.Contains(filename, ".") {
		lower := bytes.ToLower(bytes.TrimSpace(data))
		if bytes.Contains(lower, []byte("<!doctype")) ||
			bytes.Contains(lower, []byte("<!entity")) ||
			bytes.Contains(lower, []byte("system")) {
			return nil, ErrXXEDetected
		}
	}

	return data, nil
}
