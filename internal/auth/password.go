// Package auth regroupe l'authentification : hachage des mots de passe,
// jetons signés (HS256), utilisateurs, réinitialisation et limitation de débit.
// Il n'utilise que la bibliothèque standard (Go 1.24+).
package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	pbkdf2Prefix = "pbkdf2-sha256"
	pbkdf2Iter   = 600_000 // recommandation OWASP pour PBKDF2-HMAC-SHA256
	saltLen      = 16
	keyLen       = 32

	// MinPasswordLength est la longueur minimale acceptée pour un mot de passe.
	MinPasswordLength = 10
	maxPasswordLength = 256
)

// ErrWeakPassword est renvoyée quand un mot de passe ne respecte pas la politique.
var ErrWeakPassword = fmt.Errorf("le mot de passe doit contenir entre %d et %d caractères", MinPasswordLength, maxPasswordLength)

// ValidatePasswordPolicy contrôle la longueur du mot de passe.
func ValidatePasswordPolicy(pw string) error {
	n := len([]rune(pw))
	if n < MinPasswordLength || n > maxPasswordLength {
		return ErrWeakPassword
	}
	return nil
}

// HashPassword renvoie "pbkdf2-sha256$iterations$sel$hash" (base64 sans padding).
func HashPassword(pw string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, pw, salt, pbkdf2Iter, keyLen)
	if err != nil {
		return "", err
	}
	enc := base64.RawStdEncoding
	return fmt.Sprintf("%s$%d$%s$%s", pbkdf2Prefix, pbkdf2Iter, enc.EncodeToString(salt), enc.EncodeToString(key)), nil
}

// VerifyPassword compare un mot de passe à son empreinte en temps constant.
func VerifyPassword(pw, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != pbkdf2Prefix {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter < 1 || iter > 10_000_000 {
		return false
	}
	enc := base64.RawStdEncoding
	salt, err := enc.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := enc.DecodeString(parts[3])
	if err != nil || len(want) == 0 {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pw, salt, iter, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

// randomToken renvoie n octets aléatoires encodés en base64url.
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("génération aléatoire impossible")
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RandomSecret génère un secret aléatoire (usage : mot de passe / clé de dev).
func RandomSecret(n int) (string, error) { return randomToken(n) }
