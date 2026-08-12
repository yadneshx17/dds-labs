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
	ProductId string `json:"productId"`
	Quantity  string `json:"quantity"`
}

func main() {
	sidecarClient := sidecar.NewClient("http://inventory-sidecar:8888")

	logs := make(chan sidecar.LogEvent, 100)

	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		id := i

		// Async Logging.
		wg.Go(func() {
			worker(id, logs, sidecarClient)
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("POST /stock", handleStock(logs))

	srv := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	// Put blocking components in goroutines so the main goroutine can orchestrate the process lifecycle.
	go func() {
		fmt.Print("Inventory Service is running on :8082\n\n")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server Listen error: %v\n", err)
		}
	}()

	// catches OS termination signals
	shutdownChan := make(chan os.Signal, 1)
	// SIGTERM = Kubernetes or system termination request
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// block main exectuion thread until a signal is received
	sig := <-shutdownChan
	log.Printf("Received signal: %v. Initiating graceful shutdown...\n\n", sig)

	// Create a context with a timeout (e.g., 5 seconds) for completion
	// shutdown deadline context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Trigger the graceful shutdown
	// stops accepting new requests and waits for ongoing request to finish
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}

	// No more HTTP handlers can produce log events.
	close(logs)

	// Wait for workers to finish all queued events.
	wg.Wait()

	// Now background logging is finished.
	cleanupResources()

	log.Println("Server stopped cleanly.")
}

func cleanupResources() {
	time.Sleep(5 * time.Second)
	fmt.Print("Resource Cleaned up successfully\n\n")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[+] %s %s\n\n", r.Method, r.URL.Path)

	w.Write([]byte("Inventory Service"))
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

func handleStock(
	logs chan<- sidecar.LogEvent,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[+] %s %s\n", r.Method, r.URL.Path)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		// simulates long running request -> try terminating the server
		// 10 sec for Shutdown context to expire and shutdown ( dont wait longer than this context allows )
		// time.Sleep(10 * time.Second)

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

		logs <- event

		// if the logging goroutine hasn't finished by the time the process exits -> log potentially lost
		// go func() {
		// 	if err := sc.SendLog(event); err != nil {
		// 		log.Printf("sidecar logging failed: %v\n", err)
		// 	}
		// }()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Inventory updated successfully\n"))
	}
}

/*
----------- Normal Operation

Request
   │
   ▼
Handler
   │
   ▼
logs channel
   │
   ├──── W1 ──► Sidecar
   ├──── W2 ──► Sidecar
   └──── W3 ──► Sidecar

----------- Shutdown Lifecycle

SIGINT
  │
  ▼
srv.Shutdown(ctx)
  │
  ▼
No new HTTP requests
  │
  ▼
Existing handlers finish
  │
  ▼
close(logs)
  │
  ▼
Workers drain remaining events
  │
  ▼
wg.Wait()
  │
  ▼
cleanup
  │
  ▼
main returns
  │
  ▼
process exits
*/
