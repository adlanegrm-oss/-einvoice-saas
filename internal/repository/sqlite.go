package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	_ "modernc.org/sqlite"
)

type SQLiteInvoiceRepository struct {
	db *sql.DB
}

func NewSQLiteInvoiceRepository(db *sql.DB) (*SQLiteInvoiceRepository, error) {
	repo := &SQLiteInvoiceRepository{db: db}
	if err := repo.initTable(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteInvoiceRepository) initTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS invoices (
		id TEXT PRIMARY KEY,
		number TEXT NOT NULL,
		customer TEXT NOT NULL,
		issue_date DATETIME,
		items_json TEXT NOT NULL,
		total_ht REAL NOT NULL,
		total_vat REAL NOT NULL,
		total_ttc REAL NOT NULL,
		is_validated INTEGER NOT NULL
	);`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteInvoiceRepository) Save(inv invoice.Invoice) error {
	itemsBytes, err := json.Marshal(inv.Items)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO invoices (id, number, customer, issue_date, items_json, total_ht, total_vat, total_ttc, is_validated)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	isValidated := 0
	if inv.IsValidated {
		isValidated = 1
	}

	_, err = r.db.Exec(query,
		inv.ID, inv.Number, inv.Customer, inv.IssueDate, string(itemsBytes),
		inv.TotalHT, inv.TotalVAT, inv.TotalTTC, isValidated,
	)
	return err
}

func (r *SQLiteInvoiceRepository) GetAll() ([]invoice.Invoice, error) {
	query := `SELECT id, number, customer, issue_date, items_json, total_ht, total_vat, total_ttc, is_validated FROM invoices`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []invoice.Invoice
	for rows.Next() {
		var inv invoice.Invoice
		var itemsJSON string
		var isValidated int
		var issueDate time.Time

		err := rows.Scan(
			&inv.ID, &inv.Number, &inv.Customer, &issueDate,
			&itemsJSON, &inv.TotalHT, &inv.TotalVAT, &inv.TotalTTC, &isValidated,
		)
		if err != nil {
			return nil, err
		}

		inv.IssueDate = issueDate
		inv.IsValidated = (isValidated == 1)

		if err := json.Unmarshal([]byte(itemsJSON), &inv.Items); err != nil {
			return nil, err
		}

		invoices = append(invoices, inv)
	}

	return invoices, nil
}
