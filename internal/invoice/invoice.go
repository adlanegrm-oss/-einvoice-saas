package invoice

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	maxNumberLen   = 64
	maxCustomerLen = 200
	maxItems       = 1000
	maxItemDescLen = 500
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

// round2 arrondit au centime.
func round2(v float64) float64 { return math.Round(v*100) / 100 }

// CalculateTotals calcule automatiquement les montants HT, TVA et TTC.
// Chaque ligne est arrondie au centime avant d'être cumulée, ce qui évite les
// dérives de virgule flottante (0,1 + 0,2 != 0,3).
func (i *Invoice) CalculateTotals() {
	var totalHT, totalVAT float64
	for _, item := range i.Items {
		itemHT := round2(float64(item.Quantity) * item.UnitPrice)
		itemVAT := round2(itemHT * (item.VATRate / 100.0))
		totalHT += itemHT
		totalVAT += itemVAT
	}
	i.TotalHT = round2(totalHT)
	i.TotalVAT = round2(totalVAT)
	i.TotalTTC = round2(i.TotalHT + i.TotalVAT)
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Validate vérifie la conformité globale de la facture et calcule les totaux
func (i *Invoice) Validate() error {
	i.Number = strings.TrimSpace(i.Number)
	i.Customer = strings.TrimSpace(i.Customer)

	if i.Number == "" {
		return errors.New("le numéro de facture est obligatoire")
	}
	if len(i.Number) > maxNumberLen {
		return fmt.Errorf("le numéro de facture dépasse %d caractères", maxNumberLen)
	}
	if i.Customer == "" {
		return errors.New("le nom du client est obligatoire")
	}
	if len(i.Customer) > maxCustomerLen {
		return fmt.Errorf("le nom du client dépasse %d caractères", maxCustomerLen)
	}
	if len(i.Items) == 0 {
		return errors.New("la facture doit contenir au moins un article")
	}
	if len(i.Items) > maxItems {
		return fmt.Errorf("une facture ne peut pas dépasser %d lignes", maxItems)
	}
	for n, item := range i.Items {
		line := n + 1
		if len(item.Description) > maxItemDescLen {
			return fmt.Errorf("ligne %d : la description dépasse %d caractères", line, maxItemDescLen)
		}
		if item.Quantity <= 0 {
			return errors.New("la quantité d'un article doit être supérieure à zéro")
		}
		if !finite(item.UnitPrice) || item.UnitPrice < 0 {
			return errors.New("le prix unitaire ne peut pas être négatif")
		}
		if !finite(item.VATRate) || item.VATRate < 0 || item.VATRate > 100 {
			return fmt.Errorf("ligne %d : le taux de TVA doit être compris entre 0 et 100", line)
		}
	}

	i.CalculateTotals()
	if !finite(i.TotalTTC) {
		return errors.New("les montants de la facture sont invalides")
	}
	i.IsValidated = true
	return nil
}
