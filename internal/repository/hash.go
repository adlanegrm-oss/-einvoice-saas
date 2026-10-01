package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// GenerateInvoiceHash calcule un hash SHA-256 unique et déterministe pour une facture
func GenerateInvoiceHash(inv Invoice) (string, error) {
	// Sérialisation de la structure de la facture en JSON
	data, err := json.Marshal(inv)
	if err != nil {
		return "", fmt.Errorf("erreur de sérialisation pour le hachage : %w", err)
	}

	// Calcul du hash SHA-256
	hashBytes := sha256.Sum256(data)
	
	// Conversion en chaîne hexadécimale lisible
	return hex.EncodeToString(hashBytes[:]), nil
}
