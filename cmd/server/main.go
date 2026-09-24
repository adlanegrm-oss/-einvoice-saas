package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adlanegrm-oss/einvoice-saas/internal/service"
)

func main() {
	invoiceService := service.NewInvoiceService()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "eInvoice SaaS OK",
		})
	})

	http.HandleFunc("/invoices/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var invoice struct {
			InvoiceNumber string `json:"invoice_number"`
			Currency      string `json:"currency"`
			Seller        struct {
				Name string `json:"name"`
			} `json:"seller"`
			Buyer struct {
				Name string `json:"name"`
			} `json:"buyer"`
			Amounts struct {
				Net   float64 `json:"net"`
				Tax   float64 `json:"tax"`
				Total float64 `json:"total"`
			} `json:"amounts"`
		}

		if err := json.NewDecoder(r.Body).Decode(&invoice); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		result := fmt.Sprintf(
			"Invoice %s received",
			invoice.InvoiceNumber,
		)

		fmt.Println(result)

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(map[string]string{
			"status":  "received",
			"message": result,
		})

		_ = invoiceService
	})

	fmt.Println("Server listening on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}