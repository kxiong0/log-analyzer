package anomaly

import (
	"fmt"
	"log-analyzer/internal/db"
	"log/slog"
	"math"
	"time"
)

type FrequencyDetector struct{}

func (fd FrequencyDetector) Check(tdb *db.TemplateDB, tid string) (Anomaly, error) {
	mean, stddev, err := tdb.GetHourlyStats(tid)
	slog.Info("hourly stats",
		"mean", mean,
		"stddev", stddev,
		"err", err,
	)

	count, err := tdb.GetCurrHourlyCount(tid)
	if err != nil {
		return Anomaly{}, err
	}

	// Scaled-hour variance (simple heuristic)
	minutes_elapsed := (time.Now().Minute()) / 60
	expected_partial := mean * float64(minutes_elapsed)
	var_partial := math.Pow(stddev, 2) * float64(minutes_elapsed)
	std_partial := math.Sqrt(var_partial)

	z := 0.0
	if std_partial != 0 {
		z = (float64(count) - expected_partial) / std_partial
	}

	sev := SeverityInfo
	description := ""
	if z >= 4 {
		sev = SeverityCritical
		description = fmt.Sprintf("Z score for template %s is larger than 4", tid)
	} else if z >= 3 {
		sev = SeverityHigh
		description = fmt.Sprintf("Z score for template %s is larger than 3", tid)
	}

	slog.Debug(fmt.Sprintf("z score: %f", z))
	return Anomaly{
			TemplateID:  tid,
			Type:        "frequency",
			Description: description,
			Severity:    sev,
			Timestamp:   time.Now(),
		},
		nil
}
