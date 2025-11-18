package main

import (
	ingest "log-analyzer/internal/ingest"

	"net/http"
)

func main() {
	// mux := http.NewServeMux()
	http.HandleFunc("/ingest", ingest.IngestHandler)
	http.ListenAndServe(":8080", nil)
}
