python -c "
code = '''package repository

import (
	\"context\"
	\"database/sql\"
	\"encoding/json\"
	\"errors\"
	\"fmt\"
	\"strings\"
	\"time\"

	\"github.com/adlanegrm-oss/einvoice-saas/internal/invoice\"
	\"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status\"
	_ \"modernc.org/sqlite\"
)

var (
	ErrDuplicate = errors.New(\"facture deja enregistree (identifiant ou numero en double)\")
	ErrConflict  = errors.New(\"conflit d idempotence : ce numero de facture existe deja avec une empreinte differente\")
	ErrNotFound  = errors.New(\"facture introuvable\")
)

const dayLayout = \"2006-01-02\"

type DailyReport struct {
	Date          string  `json:\"date\"`
	TotalInvoices int     `json:\"total_invoices\"`
	TotalHT       float64 `json:\"total_ht\"`
	TotalVAT      float64 `json:\"total_vat\"`
	TotalTTC      float64 `json:\"total_ttc\"`
}

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
		total_ht INTEGER NOT NULL,
		total_vat INTEGER NOT NULL,
		total_ttc INTEGER NOT NULL,
		is_validated INTEGER NOT NULL,
		owner TEXT NOT NULL DEFAULT '',
		issue_day TEXT NOT NULL DEFAULT '',
		payload_hash TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS invoice_status_history (
		id TEXT PRIMARY KEY,
		invoice_id TEXT NOT NULL,
		from_state TEXT NOT NULL,
		to_state TEXT NOT NULL,
		actor TEXT NOT NULL,
		reason TEXT,
		created_at DATETIME NOT NULL,
		payload_hash TEXT NOT NULL,
		FOREIGN KEY(invoice_id) REFERENCES invoices(id)
	);

	CREATE INDEX IF NOT EXISTS idx_status_history_invoice ON invoice_status_history(invoice_id);
	`
	if _, err := r.db.Exec(create); err != nil {
		return err
	}

	for _, alter := range []string{
		`ALTER TABLE invoices ADD COLUMN owner TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE invoices ADD COLUMN issue_day TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE invoices ADD COLUMN payload_hash TEXT NOT NULL DEFAULT ''`,
	} {
		if _, err := r.db.Exec(alter); err != nil && !strings.Contains(err.Error(), \"duplicate column\") {
			return fmt.Errorf(\"migration : %w\", err)
		}
	}
	if _, err := r.db.Exec(`UPDATE invoices SET issue_day = substr(issue_date, 1, 10) WHERE issue_day = ''`); err != nil {
		return fmt.Errorf(\"migration issue_day : %w\", err)
	}

	if _, err := r.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_owner_number ON invoices(owner, number)`); err != nil {
		return fmt.Errorf(\"index d unicite des numeros : %w\", err)
	}
	return nil
}

func (r *SQLiteInvoiceRepository) Ping(ctx context.Context) error { return r.db.PingContext(ctx) }

func (r *SQLiteInvoiceRepository) Save(inv invoice.Invoice) error { return r.SaveFor(\"\", inv) }

func (r *SQLiteInvoiceRepository) SaveFor(owner string, inv invoice.Invoice) error {
	_, _, err := r.SaveForWithHash(owner, inv, \"\")
	return err
}

func (r *SQLiteInvoiceRepository) SaveForWithHash(owner string, inv invoice.Invoice, payloadHash string) (*invoice.Invoice, bool, error) {
	if inv.Number != \"\" {
		existing, err := r.GetByNumberFor(owner, inv.Number)
		if err == nil && existing != nil {
			var existingHash string
			_ = r.db.QueryRow(`SELECT payload_hash FROM invoices WHERE id = ?`, existing.ID).Scan(&existingHash)
			if payloadHash == \"\" || existingHash == \"\" || existingHash == payloadHash {
				return existing, true, nil
			}
			return nil, false, ErrConflict
		}
	}

	itemsBytes, err := json.Marshal(inv.Items)
	if err != nil {
		return nil, false, err
	}

	issue := inv.IssueDate
	if issue.IsZero() {
		issue = time.Now()
	}

	isValidated := 0
	if inv.IsValidated {
		isValidated = 1
	}

	customerName := inv.Customer.Name

	_, err = r.db.Exec(`
		INSERT INTO invoices (id, number, customer, issue_date, items_json, total_ht, total_vat, total_ttc, is_validated, owner, issue_day, payload_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inv.ID, inv.Number, customerName, issue.UTC(), string(itemsBytes),
		inv.TotalHT.Amount, inv.TotalVAT.Amount, inv.TotalTTC.Amount, isValidated, owner, issue.UTC().Format(dayLayout), payloadHash,
	)
	if err != nil && strings.Contains(err.Error(), \"UNIQUE constraint failed\") {
		return nil, false, ErrDuplicate
	}
	if err != nil {
		return nil, false, err
	}
	return &inv, false, nil
}

func (r *SQLiteInvoiceRepository) GetAll() ([]invoice.Invoice, error) { return r.ListFor(\"\") }

const selectCols = `SELECT id, number, customer, issue_date, items_json, total_ht, total_vat, total_ttc, is_validated FROM invoices`

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

func (r *SQLiteInvoiceRepository) GetByNumberFor(owner, number string) (*invoice.Invoice, error) {
	row := r.db.QueryRow(selectCols+` WHERE number = ? AND (? = '' OR owner = ?)`, number, owner, owner)
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
	var customerName string
	var itemsJSON string
	var isValidated int
	var issueDate sql.NullTime
	var htCents, vatCents, ttcCents int64

	if err := s.Scan(&inv.ID, &inv.Number, &customerName, &issueDate,
		&itemsJSON, &htCents, &vatCents, &ttcCents, &isValidated); err != nil {
		return inv, err
	}

	inv.Customer.Name = customerName
	inv.IssueDate = issueDate.Time
	inv.IsValidated = isValidated == 1
	inv.Currency = invoice.CurrencyEUR
	inv.TotalHT = invoice.NewMoneyFromCents(htCents, invoice.CurrencyEUR)
	inv.TotalVAT = invoice.NewMoneyFromCents(vatCents, invoice.CurrencyEUR)
	inv.TotalTTC = invoice.NewMoneyFromCents(ttcCents, invoice.CurrencyEUR)

	if err := json.Unmarshal([]byte(itemsJSON), &inv.Items); err != nil {
		return inv, err
	}
	return inv, nil
}

func (r *SQLiteInvoiceRepository) GetDailyReport(dateStr string) (*DailyReport, error) {
	return r.DailyReportFor(\"\", dateStr)
}

func (r *SQLiteInvoiceRepository) DailyReportFor(owner, dateStr string) (*DailyReport, error) {
	day, err := time.Parse(dayLayout, dateStr)
	if err != nil {
		return nil, fmt.Errorf(\"date invalide (format AAAA-MM-JJ attendu) : %w\", err)
	}

	report := &DailyReport{Date: day.Format(dayLayout)}
	var sumHT, sumVAT, sumTTC int64
	err = r.db.QueryRow(`
		SELECT COUNT(*),
		       COALESCE(SUM(total_ht), 0),
		       COALESCE(SUM(total_vat), 0),
		       COALESCE(SUM(total_ttc), 0)
		FROM invoices
		WHERE issue_day = ? AND (? = '' OR owner = ?)`,
		report.Date, owner, owner,
	).Scan(&report.TotalInvoices, &sumHT, &sumVAT, &sumTTC)

	if err != nil {
		return nil, err
	}

	report.TotalHT = float64(sumHT) / 100.0
	report.TotalVAT = float64(sumVAT) / 100.0
	report.TotalTTC = float64(sumTTC) / 100.0
	return report, nil
}

func (r *SQLiteInvoiceRepository) AppendStatusEvent(ctx context.Context, ev status.Event) error {
	const insert = `
	INSERT INTO invoice_status_history (id, invoice_id, from_state, to_state, actor, reason, created_at, payload_hash)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, insert,
		ev.ID, ev.InvoiceID, string(ev.From), string(ev.To),
		ev.Actor, ev.Reason, ev.CreatedAt.UTC(), ev.PayloadHash,
	)
	return err
}

func (r *SQLiteInvoiceRepository) GetStatusHistory(ctx context.Context, invoiceID string) ([]status.Event, error) {
	const query = `
	SELECT id, invoice_id, from_state, to_state, actor, reason, created_at, payload_hash
	FROM invoice_status_history
	WHERE invoice_id = ?
	ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []status.Event
	for rows.Next() {
		var ev status.Event
		var fromStr, toStr string
		var reason sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&ev.ID, &ev.InvoiceID, &fromStr, &toStr, &ev.Actor, &reason, &createdAt, &ev.PayloadHash); err != nil {
			return nil, err
		}
		ev.From = status.State(fromStr)
		ev.To = status.State(toStr)
		ev.CreatedAt = createdAt
		if reason.Valid {
			ev.Reason = reason.String
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}
'''
with open('internal/repository/sqlite.go', 'w', encoding='utf-8') as f:
    f.write(code)
"