package main

import (
	handlers "log-analyzer/internal/handlers"
	"net/http"
)

func main() {
	// mux := http.NewServeMux()
	http.HandleFunc("/ingest", handlers.IngestHandler)
	http.ListenAndServe(":8080", nil)
}
