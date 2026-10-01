package repository

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // Driver SQLite pur Go
)

type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type Invoice struct {
	ID            int64         `json:"id"`
	InvoiceNumber string        `json:"invoice_number"`
	Items         []InvoiceItem `json:"items"`
}

type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository initialise la connexion SQLite, active le mode WAL et crée les tables
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de la base de données : %w", err)
	}

	// Activer le mode WAL pour de hautes performances sous forte charge
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("erreur d'activation du mode WAL : %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS registry_invoices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invoice_number TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS registry_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invoice_id INTEGER,
		description TEXT,
		quantity INTEGER,
		unit_price REAL,
		FOREIGN KEY (invoice_id) REFERENCES registry_invoices(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS invoice_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		invoice_number TEXT NOT NULL,
		status TEXT NOT NULL,
		invoice_hash TEXT NOT NULL,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("erreur de création des tables : %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// Create insère une nouvelle facture et ses articles en base de données de manière transactionnelle
func (r *SQLiteRepository) Create(inv Invoice) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("erreur de démarrage de transaction : %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO registry_invoices (invoice_number) VALUES (?)", inv.InvoiceNumber)
	if err != nil {
		return fmt.Errorf("erreur d'insertion de la facture : %w", err)
	}

	invoiceID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur de récupération de l'ID de facture : %w", err)
	}

	for _, item := range inv.Items {
		_, err = tx.Exec(
			"INSERT INTO registry_items (invoice_id, description, quantity, unit_price) VALUES (?, ?, ?, ?)",
			invoiceID, item.Description, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return fmt.Errorf("erreur d'insertion de l'article : %w", err)
		}
	}

	return tx.Commit()
}

// LogInvoiceProcessing enregistre le hash et le statut dans le journal journalier
func (r *SQLiteRepository) LogInvoiceProcessing(invoiceNumber, status, invoiceHash string) error {
	query := `INSERT INTO invoice_logs (invoice_number, status, invoice_hash) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, invoiceNumber, status, invoiceHash)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture dans le journal de traitement : %w", err)
	}
	return nil
}
