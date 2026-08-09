package main

import (
	"dds-labs/internal/sidecar"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ServerData struct {
	OrderId   string `json:"orderId"`
	Amount    string `json:"amount"`
	PaymentId string `json:"paymentId"`
}

func main() {
	sidecarClient := sidecar.NewClient("http://127.0.0.1:8888")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("POST /charge", handleCharge(sidecarClient))

	fmt.Print("Payment Service is running on :8081\n\n")

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[+] %s %s\n\n", r.Method, r.URL.Path)

	w.Write([]byte("Payment Service"))
}

func handleCharge(sc *sidecar.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[+] %s %s\n", r.Method, r.URL.Path)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		var result ServerData

		// encoding/json matches fieldnames case-insensitively
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		status := http.StatusOK

		fmt.Printf("[-] Order ID: %s, Amount: %s, Payment ID: %s\n\n", result.OrderId, result.Amount, result.PaymentId)

		// TODO:
		// Replace one-goroutine-per-request
		// with a worker pool if logging throughput becomes a bottleneck.

		event := sidecar.LogEvent{
			Service:    "payment-service",
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
		}

		go func(e sidecar.LogEvent) {
			if err := sc.SendLog(e); err != nil {
				log.Printf("failed to send log to sidecar: %v", err)
			}
		}(event)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Payment processed successfully\n"))
	}
}
