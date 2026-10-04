package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	_ "modernc.org/sqlite"
)

func newRepo(t *testing.T) *SQLiteInvoiceRepository {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Impossible d'ouvrir la base en mémoire : %v", err)
	}
	db.SetMaxOpenConns(1) // ":memory:" est propre à chaque connexion
	t.Cleanup(func() { db.Close() })

	repo, err := NewSQLiteInvoiceRepository(db)
	if err != nil {
		t.Fatalf("Impossible d'initialiser le repository : %v", err)
	}
	return repo
}

func sample(id, number string, day time.Time) invoice.Invoice {
	inv := invoice.Invoice{
		ID: id, Number: number, Customer: invoice.Party{Name: "Client Test"}, IssueDate: day,
		Items: []invoice.InvoiceItem{{Description: "Service Cloud", Quantity: 2, UnitPrice: invoice.NewMoneyFromFloat(150.0, 2, invoice.CurrencyEUR), VATRate: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR)}},
	}
	if err := inv.Validate(); err != nil {
		panic(err)
	}
	return inv
}

func TestSQLiteInvoiceRepository(t *testing.T) {
	repo := newRepo(t)
	inv := sample("TEST-1", "INV-2026-001", time.Now())

	if err := repo.Save(inv); err != nil {
		t.Fatalf("Erreur lors de la sauvegarde SQLite : %v", err)
	}
	invoices, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Erreur lors de la récupération SQLite : %v", err)
	}
	if len(invoices) != 1 {
		t.Fatalf("Nombre de factures incorrect : attendu 1, reçu %d", len(invoices))
	}
	if invoices[0].Number != "INV-2026-001" {
		t.Errorf("Numéro de facture incorrect : %s", invoices[0].Number)
	}
}

func TestOwnerIsolation(t *testing.T) {
	repo := newRepo(t)
	day := time.Date(2026, 9, 24, 10, 0, 0, 0, time.UTC)

	if err := repo.SaveFor("t-a", sample("A1", "F-1", day)); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveFor("t-b", sample("B1", "F-1", day)); err != nil {
		t.Fatalf("le même numéro doit être permis pour un autre propriétaire : %v", err)
	}

	listA, _ := repo.ListFor("t-a")
	if len(listA) != 1 || listA[0].ID != "A1" {
		t.Errorf("t-a ne doit voir que sa facture : %+v", listA)
	}
	if _, err := repo.GetFor("B1", "t-a"); !errors.Is(err, ErrNotFound) {
		t.Errorf("t-a ne doit pas lire la facture de t-b, erreur reçue : %v", err)
	}
	if inv, err := repo.GetFor("B1", ""); err != nil || inv.ID != "B1" {
		t.Errorf("l'administrateur (owner vide) doit tout lire : %v", err)
	}
	all, _ := repo.ListFor("")
	if len(all) != 2 {
		t.Errorf("attendu 2 factures pour l'administrateur, reçu %d", len(all))
	}
}

func TestDuplicates(t *testing.T) {
	repo := newRepo(t)
	day := time.Now()
	if err := repo.SaveFor("t-a", sample("A1", "F-1", day)); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveFor("t-a", sample("A1", "F-2", day)); !errors.Is(err, ErrDuplicate) {
		t.Errorf("identifiant en double : ErrDuplicate attendue, reçue %v", err)
	}
	if err := repo.SaveFor("t-a", sample("A2", "F-1", day)); !errors.Is(err, ErrDuplicate) {
		t.Errorf("numéro en double : ErrDuplicate attendue, reçue %v", err)
	}
}

func TestDailyReport(t *testing.T) {
	repo := newRepo(t)
	d1 := time.Date(2026, 9, 24, 8, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 9, 25, 8, 0, 0, 0, time.UTC)
	repo.SaveFor("t-a", sample("A1", "F-1", d1))
	repo.SaveFor("t-a", sample("A2", "F-2", d1))
	repo.SaveFor("t-b", sample("B1", "F-3", d1))
	repo.SaveFor("t-a", sample("A3", "F-4", d2))

	rep, err := repo.DailyReportFor("t-a", "2026-09-24")
	if err != nil {
		t.Fatal(err)
	}
	// chaque facture : 2 x 150 = 300 HT, 60 TVA, 360 TTC
	if rep.TotalInvoices != 2 || rep.TotalHT != 600 || rep.TotalVAT != 120 || rep.TotalTTC != 720 {
		t.Errorf("rapport t-a inattendu : %+v", rep)
	}
	all, _ := repo.GetDailyReport("2026-09-24")
	if all.TotalInvoices != 3 {
		t.Errorf("rapport global : attendu 3 factures, reçu %d", all.TotalInvoices)
	}
	if _, err := repo.DailyReportFor("", "24/09/2026"); err == nil {
		t.Error("une date au mauvais format doit être refusée")
	}
}
