package main

import (
	"dds-labs/internal/sidecar"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// To Do:
// prevent sLOW DISk i/o from blocking sidecar request handlers.

const LogPath = "logs/2026/app.log"

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /log", handleLogs)

	fmt.Print("LoggingService is running on :8888\n\n")

	if err := http.ListenAndServe(":8888", mux); err != nil {
		fmt.Println(err)
	}
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if err := os.MkdirAll(filepath.Dir(LogPath), 0700); err != nil {
		http.Error(w, "Failed to create directories", http.StatusInternalServerError)
	}
	file, err := os.OpenFile(LogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)

	}

	defer file.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	var result sidecar.LogEvent
	if err := json.Unmarshal(body, &result); err != nil {
		// imrpovement in error ?
		http.Error(w, "failed to parse log event", http.StatusInternalServerError)
	}

	currentTime := time.Now()

	layout := "2006/01/02 15:04:05"
	formattedTimeString := currentTime.Format(layout)

	logstring := fmt.Sprintf("%s workerId: %d %s %s %d %s\n",
		formattedTimeString,
		result.WorkerId,
		result.Service,
		result.Method,
		result.StatusCode,
		result.Path,
	)

	_, err = file.WriteString(logstring)
	if err != nil {
		http.Error(w, "Failed to write data", http.StatusInternalServerError)
	}

	fmt.Println("Data successfully logged!")
}
