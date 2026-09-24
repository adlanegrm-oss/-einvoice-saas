package invoice

import (
	"errors"
	"time"
)

// InvoiceItem représente une ligne de produit ou service sur la facture
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	VATRate     float64 `json:"vat_rate"` // Ex: 20.0 pour 20%
}

// Invoice représente la structure complète d'une facture électronique
type Invoice struct {
	ID          string        `json:"id"`
	Number      string        `json:"number"`
	Customer    string        `json:"customer"`
	IssueDate   time.Time     `json:"issue_date"`
	Items       []InvoiceItem `json:"items"`
	TotalHT     float64       `json:"total_ht"`
	TotalVAT    float64       `json:"total_vat"`
	TotalTTC    float64       `json:"total_ttc"`
	IsValidated bool          `json:"is_validated"`
}

// CalculateTotals calcule automatiquement les montants HT, TVA et TTC
func (i *Invoice) CalculateTotals() {
	var totalHT, totalVAT float64
	for _, item := range i.Items {
		itemHT := float64(item.Quantity) * item.UnitPrice
		itemVAT := itemHT * (item.VATRate / 100.0)
		totalHT += itemHT
		totalVAT += itemVAT
	}
	i.TotalHT = totalHT
	i.TotalVAT = totalVAT
	i.TotalTTC = totalHT + totalVAT
}

// Validate vérifie la conformité globale de la facture et calcule les totaux
func (i *Invoice) Validate() error {
	if i.Number == "" {
		return errors.New("le numéro de facture est obligatoire")
	}
	if i.Customer == "" {
		return errors.New("le nom du client est obligatoire")
	}
	if len(i.Items) == 0 {
		return errors.New("la facture doit contenir au moins un article")
	}
	for _, item := range i.Items {
		if item.Quantity <= 0 {
			return errors.New("la quantité d'un article doit être supérieure à zéro")
		}
		if item.UnitPrice < 0 {
			return errors.New("le prix unitaire ne peut pas être négatif")
		}
	}

	i.CalculateTotals()
	i.IsValidated = true
	return nil
}
