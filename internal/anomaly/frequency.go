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

	z := 0.0
	if stddev != 0 {
		z = math.Abs((float64(count) - mean) / stddev)
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
