package anomaly

import (
	"fmt"
	"log-analyzer/internal/common"
	db "log-analyzer/internal/db"
	"log/slog"
)

type AnomalyEngine struct {
	tdb       *db.TemplateDB
	detectors []AnomalyDetector
}

func NewAnomalyEngine(tdb *db.TemplateDB) (*AnomalyEngine, error) {
	ae := AnomalyEngine{}
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
func (ae *AnomalyEngine) ProcessTemplate(tmpl common.Template) []Anomaly {
	anomalies := []Anomaly{}
	// Iterate through all detectors
	for _, d := range ae.detectors {
		as, err := d.Check(ae.tdb, tmpl)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to detect anomalies in template %s: %s", tmpl.ID, err))
			continue
		}
		anomalies = append(anomalies, as...)
	}

	ae.updateTemplateStats(tmpl)

	return anomalies
}

// Update template stats and increase count
func (ae *AnomalyEngine) updateTemplateStats(tmpl common.Template) error {
	if err := ae.tdb.CountTemplate(tmpl.ID); err != nil {
		slog.Error("Failed to count template stat:",
			"error", err)
	}
	if err := ae.tdb.CountTemplateHourly(tmpl.ID); err != nil {
		slog.Error("Failed to count template hourly stat:",
			"error", err)
	}

	if err := ae.tdb.CountTransition(tmpl.ID, tmpl.K8sMetadata.PodID); err != nil {
		slog.Error("Failed to count template transition:",
			"error", err)
	}

	return nil
}
