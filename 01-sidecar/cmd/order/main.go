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
	OrderId     string `json:"orderId"`
	ProductName string `json:"productName"`
}

func main() {
	sidecarClient := sidecar.NewClient("http://127.0.0.1:8888")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("POST /order", handleOrder(sidecarClient))

	fmt.Print("Order Service is running on :8080\n\n")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[+] %s %s\n\n", r.Method, r.URL.Path)

	w.Write([]byte("Order Service"))
}

func handleOrder(sc *sidecar.Client) http.HandlerFunc {
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

		fmt.Printf("[-] Order ID: %s, Product Name: %s\n\n", result.OrderId, result.ProductName)

		// TODO:
		// Replace one-goroutine-per-request
		// with a worker pool if logging throughput becomes a bottleneck.

		event := sidecar.LogEvent{
			Service:    "order-service",
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusOK,
		}

		go func() {
			if err := sc.SendLog(event); err != nil {
				http.Error(w, "failed to send log to sidecar", http.StatusInternalServerError)
			}
		}()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Body processed successfully\n"))
	}
}
