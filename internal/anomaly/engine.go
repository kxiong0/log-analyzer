package anomaly

import (
	"fmt"
	db "log-analyzer/internal/db"
	"log/slog"
	"sync"
)

const (
	alertThreshold = SeverityMedium
	databaseFile   = "data.db"
)

type AnomalyEngine struct {
	tdb       *db.TemplateDB
	detectors []AnomalyDetector

	prevTid   string
	prevTidMu sync.Mutex
}

func NewAnomalyEngine() (*AnomalyEngine, error) {
	ae := AnomalyEngine{}

	// Init DB
	tdb, err := db.NewTemplateDB(databaseFile)
	if err != nil {
		return nil, err
	}
	ae.tdb = tdb

	ae.AddAnomalyDetector(FrequencyDetector{})
	return &ae, nil
}

func (ae *AnomalyEngine) AddAnomalyDetector(ad AnomalyDetector) {
	ae.detectors = append(ae.detectors, ad)
}

func (ae *AnomalyEngine) ProcessTemplate(tid string) {
	// Iterate through all detectors
	for _, d := range ae.detectors {
		a, err := d.Check(ae.tdb, tid)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to process template %s: %s", tid, err))
			continue
		}

		if a.Severity >= alertThreshold {
			// TODO: send alert
			slog.Warn(fmt.Sprintf("Triggered alerts - %s", a.Description))
		}
	}

	ae.updateTemplateStats(tid)
}

// Update template stats and increase count
func (ae *AnomalyEngine) updateTemplateStats(tid string) error {
	if err := ae.tdb.CountTemplate(tid); err != nil {
		slog.Error("Failed to count template stat:",
			"error", err)
	}
	if err := ae.tdb.CountTemplateHourly(tid); err != nil {
		slog.Error("Failed to count template hourly stat:",
			"error", err)
	}

	ae.prevTidMu.Lock()
	if err := ae.tdb.CountTransition(ae.prevTid, tid); err != nil {
		slog.Error("Failed to count template transition:",
			"error", err)
	}
	ae.prevTid = tid
	ae.prevTidMu.Unlock()

	return nil
}
