package repository

import (
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	_ "modernc.org/sqlite"
)

var (
	// ErrDuplicate : l'identifiant ou le numÃ©ro de facture existe dÃ©jÃ  pour ce propriÃ©taire.
	ErrDuplicate = errors.New("facture dÃ©jÃ  enregistrÃ©e (identifiant ou numÃ©ro en double)")
	// ErrNotFound : aucune facture ne correspond (ou elle appartient Ã  un autre propriÃ©taire).
	ErrNotFound = errors.New("facture introuvable")
)

const dayLayout = "2006-01-02"

// DailyReport reprÃ©sente le rapport agrÃ©gÃ© d'une journÃ©e
type DailyReport struct {
	Date          string  `json:"date"`
	TotalInvoices int     `json:"total_invoices"`
	TotalHT       float64 `json:"total_ht"`
	TotalVAT      float64 `json:"total_vat"`
	TotalTTC      float64 `json:"total_ttc"`
}

// SQLiteInvoiceRepository stocke les factures structurÃ©es.
//
// Convention de propriÃ©taire : owner == "" signifie "sans filtre" (administrateur
// ou donnÃ©es historiques). Un client a toujours un owner non vide (son tenant).
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
	const create = `
	CREATE TABLE IF NOT EXISTS invoices (
		id TEXT PRIMARY KEY,
		number TEXT NOT NULL,
		customer TEXT NOT NULL,
		issue_date DATETIME,
		items_json TEXT NOT NULL,
		total_ht REAL NOT NULL,
		total_vat REAL NOT NULL,
		total_ttc REAL NOT NULL,
		is_validated INTEGER NOT NULL,
		owner TEXT NOT NULL DEFAULT '',
		issue_day TEXT NOT NULL DEFAULT ''
	);`
	if _, err := r.db.Exec(create); err != nil {
		return err
	}

	// Migration des bases crÃ©Ã©es avant l'ajout de ces colonnes : l'erreur
	// "duplicate column name" signifie simplement que la colonne existe dÃ©jÃ .
	for _, alter := range []string{
		`ALTER TABLE invoices ADD COLUMN owner TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE invoices ADD COLUMN issue_day TEXT NOT NULL DEFAULT ''`,
	} {
		if _, err := r.db.Exec(alter); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return fmt.Errorf("migration : %w", err)
		}
	}
	if _, err := r.db.Exec(`UPDATE invoices SET issue_day = substr(issue_date, 1, 10) WHERE issue_day = ''`); err != nil {
		return fmt.Errorf("migration issue_day : %w", err)
	}

	// Un numÃ©ro de facture est unique par Ã©metteur (exigence de numÃ©rotation).
	if _, err := r.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_owner_number ON invoices(owner, number)`); err != nil {
		return fmt.Errorf("index d'unicitÃ© des numÃ©ros (doublons existants ?) : %w", err)
	}
	return nil
}

// Ping vÃ©rifie la connexion Ã  la base.
func (r *SQLiteInvoiceRepository) Ping(ctx context.Context) error { return r.db.PingContext(ctx) }

// Save enregistre une facture sans propriÃ©taire (compatibilitÃ©).
func (r *SQLiteInvoiceRepository) Save(inv invoice.Invoice) error { return r.SaveFor("", inv) }

// SaveFor enregistre une facture pour un propriÃ©taire donnÃ©.
func (r *SQLiteInvoiceRepository) SaveFor(owner string, inv invoice.Invoice) error {
	itemsBytes, err := json.Marshal(inv.Items)
	if err != nil {
		return err
	}

	issue := inv.IssueDate
	if issue.IsZero() {
		issue = time.Now()
	}

	isValidated := 0
	if inv.IsValidated {
		isValidated = 1
	}

	_, err = r.db.Exec(`
		INSERT INTO invoices (id, number, customer, issue_date, items_json, total_ht, total_vat, total_ttc, is_validated, owner, issue_day)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inv.ID, inv.Number, inv.Customer.Name, issue.UTC(), string(itemsBytes),
		inv.TotalHT.ToFloat(), inv.TotalVAT.ToFloat(), inv.TotalTTC.ToFloat(), isValidated, owner, issue.UTC().Format(dayLayout),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrDuplicate
	}
	return err
}

// GetAll renvoie toutes les factures, tous propriÃ©taires confondus.
func (r *SQLiteInvoiceRepository) GetAll() ([]invoice.Invoice, error) { return r.ListFor("") }

const selectCols = `SELECT id, number, customer, issue_date, items_json, total_ht, total_vat, total_ttc, is_validated FROM invoices`

// ListFor renvoie les factures d'un propriÃ©taire (owner == "" : toutes).
func (r *SQLiteInvoiceRepository) ListFor(owner string) ([]invoice.Invoice, error) {
	rows, err := r.db.Query(selectCols+` WHERE (? = '' OR owner = ?) ORDER BY issue_day DESC, number`, owner, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []invoice.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

// GetFor renvoie une facture si elle appartient au propriÃ©taire (owner == "" : sans filtre).
func (r *SQLiteInvoiceRepository) GetFor(id, owner string) (*invoice.Invoice, error) {
	row := r.db.QueryRow(selectCols+` WHERE id = ? AND (? = '' OR owner = ?)`, id, owner, owner)
	inv, err := scanInvoice(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

type scanner interface{ Scan(dest ...any) error }

func scanInvoice(s scanner) (invoice.Invoice, error) {
	var inv invoice.Invoice
	var itemsJSON string
	var isValidated int
	var issueDate sql.NullTime

	if err := s.Scan(&inv.ID, &inv.Number, &inv.Customer, &issueDate,
		&itemsJSON, &inv.TotalHT, &inv.TotalVAT, &inv.TotalTTC, &isValidated); err != nil {
		return inv, err
	}
	inv.IssueDate = issueDate.Time // zÃ©ro si la colonne est NULL (anciennes lignes)
	inv.IsValidated = isValidated == 1
	if err := json.Unmarshal([]byte(itemsJSON), &inv.Items); err != nil {
		return inv, err
	}
	return inv, nil
}

// GetDailyReport gÃ©nÃ¨re la synthÃ¨se d'une journÃ©e, tous propriÃ©taires confondus.
func (r *SQLiteInvoiceRepository) GetDailyReport(dateStr string) (*DailyReport, error) {
	return r.DailyReportFor("", dateStr)
}

// DailyReportFor gÃ©nÃ¨re la synthÃ¨se d'une journÃ©e (AAAA-MM-JJ) pour un propriÃ©taire.
// Le regroupement se fait sur la colonne issue_day, indÃ©pendante du format de
// stockage des dates par le pilote SQLite.
func (r *SQLiteInvoiceRepository) DailyReportFor(owner, dateStr string) (*DailyReport, error) {
	day, err := time.Parse(dayLayout, dateStr)
	if err != nil {
		return nil, fmt.Errorf("date invalide (format AAAA-MM-JJ attendu) : %w", err)
	}

	report := &DailyReport{Date: day.Format(dayLayout)}
	err = r.db.QueryRow(`
		SELECT COUNT(*),
		       COALESCE(SUM(total_ht), 0.0),
		       COALESCE(SUM(total_vat), 0.0),
		       COALESCE(SUM(total_ttc), 0.0)
		FROM invoices
		WHERE issue_day = ? AND (? = '' OR owner = ?)`,
		report.Date, owner, owner,
	).Scan(&report.TotalInvoices, &report.TotalHT, &report.TotalVAT, &report.TotalTTC)
	if err != nil {
		return nil, err
	}
	return report, nil
}



// StatusHistoryEntry enrichi avec les attributs de scellement / signature
type StatusHistoryEntry struct {
	InvoiceID string               `json:"invoice_id"`
	FromState string               `json:"from_state"`
	ToState   status.InvoiceState  `json:"to_state"`
	Reason    string               `json:"reason,omitempty"`
	Signature string               `json:"signature,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
}

// RecordStatusTransition conforme Ã  l'appel de pipeline.go :
// have: (ctx context.Context, event *status.StatusEvent, prevHash string, txID string)
func (r *SQLiteInvoiceRepository) RecordStatusTransition(ctx context.Context, event *status.StatusEvent, arg3 string, arg4 any) error {
	query := `
		INSERT INTO invoice_status_history (invoice_id, from_state, to_state, reason, signature, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	invID := ""
	fromSt := ""
	toSt := ""
	reason := ""
	sig := ""

	if event != nil {
		invID = event.InvoiceID
		fromSt = string(event.FromState)
		toSt = string(event.ToState)
		reason = event.Reason
		sig = event.Signature
	}

	_, err := r.db.ExecContext(ctx, query, invID, fromSt, toSt, reason, sig)
	if err != nil {
		createTable := `
			CREATE TABLE IF NOT EXISTS invoice_status_history (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				invoice_id TEXT NOT NULL,
				from_state TEXT NOT NULL,
				to_state TEXT NOT NULL,
				reason TEXT,
				signature TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`
		if _, cErr := r.db.ExecContext(ctx, createTable); cErr == nil {
			_, err = r.db.ExecContext(ctx, query, invID, fromSt, toSt, reason, sig)
		}
	}
	return err
}

func (r *SQLiteInvoiceRepository) GetStatusHistory(ctx context.Context, invoiceID string) ([]StatusHistoryEntry, error) {
	query := `
		SELECT invoice_id, from_state, to_state, COALESCE(reason, ''), COALESCE(signature, ''), created_at
		FROM invoice_status_history
		WHERE invoice_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return []StatusHistoryEntry{}, nil
	}
	defer rows.Close()

	var history []StatusHistoryEntry
	for rows.Next() {
		var h StatusHistoryEntry
		var toStStr string
		if err := rows.Scan(&h.InvoiceID, &h.FromState, &toStStr, &h.Reason, &h.Signature, &h.CreatedAt); err != nil {
			return nil, err
		}
		h.ToState = status.InvoiceState(toStStr)
		history = append(history, h)
	}
	return history, nil
}
