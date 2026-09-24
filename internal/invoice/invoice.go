package invoice

import (
	"errors"
	"time"
)

// Invoice représente la structure d'une facture électronique basique
type Invoice struct {
	ID          string    `json:"id"`
	Number      string    `json:"number"`
	Customer    string    `json:"customer"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	IssueDate   time.Time `json:"issue_date"`
	IsValidated bool      `json:"is_validated"`
}

// Validate vérifie la conformité minimale de la facture
func (i *Invoice) Validate() error {
	if i.Number == "" {
		return errors.New("le numéro de facture est obligatoire")
	}
	if i.Customer == "" {
		return errors.New("le nom du client est obligatoire")
	}
	if i.Amount <= 0 {
		return errors.New("le montant doit être supérieur à zéro")
	}
	return nil
}
