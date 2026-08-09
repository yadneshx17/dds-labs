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
	ProductId string `json:"productId"`
	Quantity  string `json:"quantity"`
}

func main() {
	sidecarClient := sidecar.NewClient("http://127.0.0.1:8888")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("POST /stock", handleStock(sidecarClient))

	fmt.Print("Inventory Service is running on :8082\n\n")

	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[+] %s %s\n\n", r.Method, r.URL.Path)

	w.Write([]byte("Inventory Service"))
}

func handleStock(sc *sidecar.Client) http.HandlerFunc {
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

		// Bussiness Logic
		fmt.Printf("[-] Order ID: %s, Product ID: %s, Quantity: %s\n\n", result.OrderId, result.ProductId, result.Quantity)

		// TODO:
		// Replace one-goroutine-per-request
		// with a worker pool if logging throughput becomes a bottleneck.

		status := http.StatusOK
		event := sidecar.LogEvent{
			Service:    "inventory-service",
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
		}

		go func() {
			if err := sc.SendLog(event); err != nil {
				log.Printf("sidecar logging failed: %v", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Inventory updated successfully\n"))
	}
}
