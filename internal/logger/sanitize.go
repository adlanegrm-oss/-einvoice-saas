package logger

import (
	"net/http"
	"regexp"
	"strings"
)

var (
	bearerRegex = regexp.MustCompile(`(?i)Bearer\s+([A-Za-z0-9\-_\.]+)`)
	secretRegex = regexp.MustCompile(`(?i)"(password|secret|token|smtp_pass|new_password)"\s*:\s*"[^"]+"`)
)

// SanitizeHeaders filtre les en-têtes avant écriture dans les logs
func SanitizeHeaders(h http.Header) map[string]string {
	clean := make(map[string]string)
	for k, v := range h {
		val := strings.Join(v, ", ")
		if strings.EqualFold(k, "Authorization") {
			val = bearerRegex.ReplaceAllString(val, "Bearer [REDACTED]")
		}
		if strings.EqualFold(k, "Cookie") {
			val = "[REDACTED]"
		}
		clean[k] = val
	}
	return clean
}

// SanitizePayload masque les mots de passe et clés dans les payloads JSON
func SanitizePayload(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	sanitized := secretRegex.ReplaceAll(body, []byte(`"$1":"[REDACTED]"`))
	return string(sanitized)
}

// RedactString masque une chaîne brute si elle contient des secrets sensibles
func RedactString(raw string) string {
	raw = bearerRegex.ReplaceAllString(raw, "Bearer [REDACTED]")
	return raw
}