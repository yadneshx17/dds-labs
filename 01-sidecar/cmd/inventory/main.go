package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ServerData struct {
	OrderId   string `json:"orderId"`
	ProductId string `json:"productId"`
	Quantity  string `json:"quantity"`
}

type LogEvent struct {
	Service    string
	Method     string
	Path       string
	StatusCode int
	LatencyMS  int64
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("POST /stock", handleStock)

	fmt.Print("Server is running on :8082\n\n")

	if err := http.ListenAndServe(":8082", mux); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[+] %s %s\n\n", r.Method, r.URL.Path)

	w.Write([]byte("Inventory Service"))
}

func handleStock(w http.ResponseWriter, r *http.Request) {
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

	start := time.Now()
	status := http.StatusOK

	fmt.Printf("[-] Order ID: %s, Product ID: %s, Quantity: %s\n\n", result.OrderId, result.ProductId, result.Quantity)

	// TODO:
	// Replace one-goroutine-per-request
	// with a worker pool if logging throughput becomes a bottleneck.

	latency := time.Since(start).Milliseconds()

	event := LogEvent{
		Service:    "inventory-service",
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: status,
		LatencyMS:  latency,
	}

	go func(e LogEvent) {
		if err := doRequest("POST", "http://127.0.0.1:8883/log", e); err != nil {
			log.Printf("failed to send log to sidecar: %v", err)
		}
	}(event)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Inventory updated successfully\n"))
}

func doRequest(method, url string, event LogEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sidecar returned %d", resp.StatusCode)
	}

	return nil
}
