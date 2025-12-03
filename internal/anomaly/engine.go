package anomaly

import (
	"fmt"
	db "log-analyzer/internal/db"
	"log/slog"
)

const (
	databaseFile = "data.db"
)

type AnomalyEngine struct {
	tdb       *db.TemplateDB
	detectors []AnomalyDetector
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
	ae.AddAnomalyDetector(SequenceDetector{})
	ae.AddAnomalyDetector(TimingDetector{})
	return &ae, nil
}

func (ae *AnomalyEngine) AddAnomalyDetector(ad AnomalyDetector) {
	ae.detectors = append(ae.detectors, ad)
}

// Process template through detectors to detect anomalies and update
// template statistics
// Return slice of anomalies detected
func (ae *AnomalyEngine) ProcessTemplate(tid string) []Anomaly {
	anomalies := []Anomaly{}
	// Iterate through all detectors
	for _, d := range ae.detectors {
		as, err := d.Check(ae.tdb, tid)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to detect anomalies in template %s: %s", tid, err))
			continue
		}
		anomalies = append(anomalies, as...)
	}

	ae.updateTemplateStats(tid)

	return anomalies
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

	if err := ae.tdb.CountTransition(tid); err != nil {
		slog.Error("Failed to count template transition:",
			"error", err)
	}

	return nil
}
