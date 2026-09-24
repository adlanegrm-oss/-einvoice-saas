package models

import "time"

type Invoice struct {
	ID            string    `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	IssueDate     time.Time `json:"issue_date"`

	Seller   Party `json:"seller"`
	Buyer    Party `json:"buyer"`
	Amounts  Amounts `json:"amounts"`
	Currency string   `json:"currency"`

	Status string `json:"status"`
}

type Party struct {
	Name       string `json:"name"`
	TaxID      string `json:"tax_id"`
	Country    string `json:"country"`
	Address    string `json:"address"`
}

type Amounts struct {
	Net   float64 `json:"net"`
	Tax   float64 `json:"tax"`
	Total float64 `json:"total"`
}