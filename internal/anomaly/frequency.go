package anomaly

import (
	"fmt"
	"log-analyzer/internal/common"
	"log-analyzer/internal/db"
	"log/slog"
	"math"
	"time"
)

type FrequencyDetector struct{}

func (fd FrequencyDetector) Check(tdb *db.TemplateDB, tmpl common.Template) ([]Anomaly, error) {
	mean, stddev, err := tdb.GetHourlyStats(tmpl.ID)
	if err != nil {
		return nil, err
	}

	count, err := tdb.GetCurrHourlyCount(tmpl.ID)
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

	sev := SeverityFromZScore(z)
	anomaly := Anomaly{TemplateID: tmpl.ID, Type: AnomalyTypeFrequency, Severity: sev, Timestamp: time.Now()}
	if sev > SeverityInfo {
		anomaly.Description = fmt.Sprintf(
			"abnormal frequency spike detected for template %s: Frequency deviates significantly from baseline (Z = %f)",
			tmpl.ID, z,
		)
	}

	return []Anomaly{anomaly}, nil
}
