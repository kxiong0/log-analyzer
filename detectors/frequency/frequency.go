package frequency

import (
	"fmt"
	"log-analyzer/internal/anomaly"
	"log-analyzer/internal/common"
	"log-analyzer/internal/db"
	"log/slog"
	"math"
	"time"
)

const (
	sweepInterval = 5 * time.Second
)

type FrequencyDetector struct {
	tdb *db.TemplateDB
}

func (fd *FrequencyDetector) Init(tdb *db.TemplateDB) error {
	fd.tdb = tdb
	return nil
}

func (fd FrequencyDetector) Start(done <-chan bool) error {
	ticker := time.NewTicker(sweepInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := fd.sweep()
				if err != nil {
					slog.Error(fmt.Sprintf("hourly stats check failed: %s", err))
				}
			case <-done:
				fmt.Println("Stopping sweep scheduler...")
				return
			}
		}
	}()
	return nil
}

func (fd FrequencyDetector) Check(tmpl common.Template) ([]anomaly.Anomaly, error) {
	mean, stddev, err := fd.tdb.GetHourlyStats(tmpl.ID)
	if err != nil {
		return nil, err
	}

	count, err := fd.tdb.GetCurrHourlyCount(tmpl.ID)
	if err != nil {
		return nil, err
	}

	// Scaled-hour variance (simple heuristic)
	minutes_elapsed := float64(time.Now().Minute()) / 60
	expected_partial := mean * minutes_elapsed
	var_partial := math.Pow(stddev, 2) * minutes_elapsed
	std_partial := math.Sqrt(var_partial)

	z := 0.0
	if std_partial != 0 {
		z = (float64(count) - expected_partial) / std_partial
	}

	slog.Debug(fmt.Sprintf("Template: %s | Frequency Z score: %f", tmpl.ID, z))

	sev := anomaly.SeverityFromZScore(z)
	a := anomaly.Anomaly{TemplateID: tmpl.ID, Type: anomaly.AnomalyTypeFrequency, Severity: sev, Timestamp: time.Now()}
	if sev > anomaly.SeverityInfo {
		a.Description = fmt.Sprintf(
			"abnormal frequency spike detected for template %s: Frequency deviates significantly from baseline (Z = %f)",
			tmpl.ID, z,
		)
	}

	return []anomaly.Anomaly{a}, nil
}

func (fd FrequencyDetector) sweep() error {
	allTemplates, err := fd.tdb.GetAllTemplates()
	if err != nil {
		return fmt.Errorf("failed to get all templates: %s", err)
	}

	slog.Debug("Updating hourly stats for all templates")

	for _, c := range allTemplates {
		for _, tmpl := range c {
			err := fd.tdb.InsertHourlyRow(tmpl.ID)
			if err != nil {
				slog.Warn(fmt.Sprintf("Could not update hourly stats for template %s", tmpl.ID))
			}
		}
	}
	slog.Debug("Hourly stats for all templates updated")

	return nil

}
