package main

import (
	"bytes"
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

type LogEvent struct {
	Service    string
	Method     string
	Path       string
	StatusCode int
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("POST /order", handleOrder)

	fmt.Print("Server is running on :8080\n\n")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[+] %s %s\n\n", r.Method, r.URL.Path)

	w.Write([]byte("Order Service"))
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
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

	event := LogEvent{
		Service:    "order-service",
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusOK,
	}

	go func(e LogEvent) {
		if err := doRequest("POST", "http://127.0.0.1:8888/log", e); err != nil {
			http.Error(w, "failed to send log to sidecar", http.StatusInternalServerError)
		}
	}(event)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Body processed successfully\n"))
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
