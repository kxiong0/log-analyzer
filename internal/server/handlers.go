package server

import (
	"encoding/json"
	"fmt"
	"io"
	common "log-analyzer/internal/common"
	"net/http"
	"os"
)

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
		tid := s.lp.ParseLog(le.Log)
		s.ae.ProcessTemplate(tid)

		lstr := fmt.Sprintln("raw log", le.Log)
		lstr += fmt.Sprintln("template id: ", tid)
		if _, err := f.Write([]byte(lstr + "\n")); err != nil {
			http.Error(w, "failed to write to file", http.StatusBadRequest)
			return
		}
	}
	if err := f.Close(); err != nil {
		http.Error(w, "failed to close file", http.StatusBadRequest)
	}
}
