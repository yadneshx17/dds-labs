package main

import (
	"context"
	"dds-labs/internal/sidecar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ServerData struct {
	OrderId   string `json:"orderId"`
	Amount    string `json:"amount"`
	PaymentId string `json:"paymentId"`
}

func main() {
	sidecarClient := sidecar.NewClient("http://sidecar:8888")

	logs := make(chan sidecar.LogEvent, 100)

	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		id := i

		wg.Go(func() {
			worker(id, logs, sidecarClient)
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("POST /charge", handleCharge(logs))

	srv := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	go func() {
		fmt.Print("Payment Service is running on :8081\n\n")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server Listen error: %v\n", err)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdownChan
	log.Printf("Received signal: %v. Initiating graceful shutdown...\n\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}
	close(logs)

	wg.Wait()

	cleanupResources()

	log.Println("Server stopped cleanly.")
}

func cleanupResources() {
	time.Sleep(5 * time.Second)
	fmt.Print("Resource Cleaned up successfully\n\n")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[+] %s %s\n\n", r.Method, r.URL.Path)

	w.Write([]byte("Payment Service"))
}

func worker(
	WorkerId int,
	logs <-chan sidecar.LogEvent,
	sc *sidecar.Client,
) {
	for event := range logs {
		fmt.Printf(
			"Worker %d processing: %+v\n\n",
			WorkerId,
			event.Service,
		)

		event.WorkerId = WorkerId

		if err := sc.SendLog(event); err != nil {
			log.Printf(
				"sidecar logging failed: %v\n",
				err,
			)
		}
	}

	fmt.Printf("Worker %d stopped\n", WorkerId)
}

func handleCharge(
	logs chan<- sidecar.LogEvent,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[+] %s %s\n", r.Method, r.URL.Path)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		time.Sleep(5 * time.Second)

		var result ServerData

		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		status := http.StatusOK

		fmt.Printf("[-] Order ID: %s, Amount: %s, Payment ID: %s\n\n", result.OrderId, result.Amount, result.PaymentId)

		event := sidecar.LogEvent{
			Service:    "payment-service",
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
		}

		logs <- event

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Payment processed successfully\n"))
	}
}
