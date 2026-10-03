package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adlanegrm-oss/einvoice-saas/internal/compliance/en16931"
	"github.com/adlanegrm-oss/einvoice-saas/internal/handler"
	"github.com/adlanegrm-oss/einvoice-saas/internal/invoice"
	"github.com/adlanegrm-oss/einvoice-saas/internal/lifecycle/status"
	"github.com/adlanegrm-oss/einvoice-saas/internal/repository"
	"github.com/adlanegrm-oss/einvoice-saas/internal/routing/dispatcher"
	"github.com/adlanegrm-oss/einvoice-saas/internal/service"
	_ "modernc.org/sqlite"
)

type mockDir struct{}

func (m *mockDir) Lookup(ctx context.Context, id string) (*dispatcher.TargetEndpoint, error) {
	return &dispatcher.TargetEndpoint{
		ReceiverID:   id,
		PlatformName: "PDP Recette",
		AS4Endpoint:  "https://as4.test",
	}, nil
}

type mockAS4 struct{}

func (m *mockAS4) SendPayload(ctx context.Context, ep *dispatcher.TargetEndpoint, p []byte) (string, error) {
	return "RECEIPT-QUALIF", nil
}

type Stage struct {
	Name        string
	TotalReqs   int
	Concurrency int
}

func main() {
	fmt.Println("================================================================================")
	fmt.Println(" CAMPAGNE DE QUALIFICATION DE CHARGE & RESILIENCE SQLITE WAL (EN 16931)")
	fmt.Println("================================================================================")

	tempDir, err := os.MkdirTemp("", "einvoice-load-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "qualification.db")
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repo, err := repository.NewSQLiteInvoiceRepository(db)
	if err != nil {
		panic(err)
	}

	val := en16931.NewValidator()
	sm := status.NewStateMachine()
	disp := dispatcher.NewDispatcher(&mockDir{}, &mockAS4{})
	pipeline := service.NewInvoicePipeline(repo, val, sm, disp)
	pipelineHandler := handler.NewPipelineHandler(pipeline, repo)

	stages := []Stage{
		{Name: "Palier A (Amorçage)", TotalReqs: 100, Concurrency: 10},
		{Name: "Palier B (Montée)", TotalReqs: 500, Concurrency: 25},
		{Name: "Palier C (Charge)", TotalReqs: 1000, Concurrency: 50},
		{Name: "Palier D (Stress Max)", TotalReqs: 2500, Concurrency: 50},
	}

	for _, stage := range stages {
		runStage(stage, pipelineHandler)
	}

	fmt.Println("\n>> QUALIFICATION TERMINEE AVEC SUCCES.")
}

func runStage(stage Stage, h *handler.PipelineHandler) {
	fmt.Printf("\n--- Exécution : %s [%d requêtes / %d workers concurrents] ---\n",
		stage.Name, stage.TotalReqs, stage.Concurrency)

	jobs := make(chan int, stage.TotalReqs)
	latencies := make([]time.Duration, stage.TotalReqs)
	var latMu sync.Mutex

	var (
		c200 int64
		c422 int64
		c5xx int64
	)

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < stage.Concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for id := range jobs {
				// 10% d'injections volontairement invalides (calcul TTC faux)
				isInvalid := (id % 10) == 0

				ttc := 120.0
				if isInvalid {
					ttc = 999.0 // Doit être rejeté proprement par le validateur -> 422
				}

				inv := invoice.Invoice{
					ID:        fmt.Sprintf("INV-%d-%d", workerID, id),
					Number:    fmt.Sprintf("FA-2026-%07d", id),
					Currency:  "EUR",
					IssueDate: time.Now(),
					Seller: invoice.Party{
						Name:  "E-Invoice SaaS Tech",
						SIRET: "12345678901234",
						Address: invoice.PostalAddress{
							StreetName:  "10 Rue de la Paix",
							PostalZone:  "75001",
							CityName:    "Paris",
							CountryCode: "FR",
						},
					},
					Customer: invoice.Party{
						Name:  fmt.Sprintf("Client Tenant %d", id%5),
						SIRET: "98765432109876",
					},
					Items: []invoice.InvoiceItem{
						{Description: "Abonnement EDI Cloud", Quantity: 1, UnitPrice: invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR), VATRate: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR)},
					},
					TotalHT:  invoice.NewMoneyFromFloat(100.0, 2, invoice.CurrencyEUR),
					TotalVAT: invoice.NewMoneyFromFloat(20.0, 2, invoice.CurrencyEUR),
					TotalTTC: invoice.NewMoneyFromFloat(ttc, 2, invoice.CurrencyEUR),
				}

				body, _ := json.Marshal(inv)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices/emit", bytes.NewReader(body))
				req.Header.Set("X-Tenant-ID", fmt.Sprintf("tenant-%d", id%5))
				rec := httptest.NewRecorder()

				t0 := time.Now()
				h.EmitInvoice(rec, req)
				dur := time.Since(t0)

				latMu.Lock()
				latencies[id] = dur
				latMu.Unlock()

				switch rec.Code {
				case http.StatusOK:
					atomic.AddInt64(&c200, 1)
				case http.StatusUnprocessableEntity:
					atomic.AddInt64(&c422, 1)
				default:
					atomic.AddInt64(&c5xx, 1)
				}
			}
		}(w)
	}

	for i := 0; i < stage.TotalReqs; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	totalDuration := time.Since(start)

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := latencies[int(float64(len(latencies))*0.50)]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]
	rps := float64(stage.TotalReqs) / totalDuration.Seconds()

	fmt.Printf("Durée totale : %v | Débit : %.1f req/s\n", totalDuration.Round(time.Millisecond), rps)
	fmt.Printf("Statuts HTTP : 200 OK: %d | 422 Rejets attendus: %d | 5xx Erreurs: %d\n", c200, c422, c5xx)
	fmt.Printf("Latences     : p50: %v | p95: %v | p99: %v\n", p50, p95, p99)

	if c5xx > 0 {
		fmt.Printf("ALERTE : %d erreurs serveur détectées (verrou ou blocage SQLite)\n", c5xx)
	}
}
