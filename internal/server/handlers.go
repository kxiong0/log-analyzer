package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log-analyzer/internal/anomaly"
	"log-analyzer/internal/common"
	"log/slog"
	"net/http"
)

const (
	alertThreshold = anomaly.SeverityMedium
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

	for _, le := range logs {
		tmpl, newTemplate := s.lp.ParseLog(le.Log)
		if newTemplate {
			slog.Debug(fmt.Sprintf("New template detected: %s", tmpl.ID))
			// TODO mark AnomalyTypeNewTemplate as pending to be sent out
			// TODO send alert for new Template
		}

		anomalies := s.ae.ProcessTemplate(tmpl)
		for _, a := range anomalies {
			if a.Severity >= alertThreshold {
				slog.Error(fmt.Sprintf("Alert triggered: %s", a.Description))
				// TODO mark Anomaly Type as pending to be sent out
			}
		}
	}
}
