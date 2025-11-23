package main

import (
	handlers "log-analyzer/internal/handlers"
	"net/http"
)

func main() {
	s := handlers.NewServer()
	http.HandleFunc("/ingest", s.Ingest)
	http.ListenAndServe(":8080", nil)
}
