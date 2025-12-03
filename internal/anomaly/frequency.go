package anomaly

import (
	"fmt"
	"log-analyzer/internal/db"
	"log/slog"
	"math"
	"time"
)

type FrequencyDetector struct{}

func (fd FrequencyDetector) Check(tdb *db.TemplateDB, tid string) ([]Anomaly, error) {
	mean, stddev, err := tdb.GetHourlyStats(tid)
	slog.Info("hourly stats",
		"mean", mean,
		"stddev", stddev,
		"err", err,
	)

	count, err := tdb.GetCurrHourlyCount(tid)
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

	anomalies := []Anomaly{}
	anomaly := Anomaly{TemplateID: tid, Type: "frequency", Timestamp: time.Now()}
	if z >= 4 {
		anomaly.Severity = SeverityHigh
		anomaly.Description = fmt.Sprintf("Z score for template %s is larger than 4", tid)
		anomalies = append(anomalies, anomaly)
	} else if z >= 3 {
		anomaly.Severity = SeverityMedium
		anomaly.Description = fmt.Sprintf("Z score for template %s is larger than 3", tid)
		anomalies = append(anomalies, anomaly)
	}

	return anomalies,
		nil
}
