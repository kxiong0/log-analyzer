package main

import (
	"log"
	server "log-analyzer/internal/server"
	"net/http"
)

func main() {
	s, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	http.HandleFunc("/ingest", s.Ingest)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
