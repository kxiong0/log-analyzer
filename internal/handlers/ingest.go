package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	common "log-analyzer/internal/common"
	p "log-analyzer/internal/parser"
)

const (
	outputFile = "access.log"
)

func NewServer() *Server {
	s := Server{
		lp: *p.NewLogParser(),
	}
	return &s
}

type Server struct {
	lp p.LogParser
}

func (s *Server) Ingest(w http.ResponseWriter, req *http.Request) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var logs []common.LogEvent
	err = json.Unmarshal(bodyBytes, &logs)
	if err != nil {
		errString := fmt.Sprintf("Unable to parse log as a JSON string: %s", err)
		http.Error(w, errString, http.StatusBadRequest)
		return
	}

	// write output to a file for now
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "failed to write to file", http.StatusBadRequest)
		return
	}

	for _, le := range logs {
		tokens := s.lp.ParseLog(le.Log)
		lstr := fmt.Sprintln("raw log", le.Log)
		lstr += fmt.Sprintln("tokens: ", tokens)
		if _, err := f.Write([]byte(lstr + "\n")); err != nil {
			http.Error(w, "failed to write to file", http.StatusBadRequest)
			return
		}
	}
	if err := f.Close(); err != nil {
		http.Error(w, "failed to close file", http.StatusBadRequest)
	}
}
