package anomaly

import (
	"database/sql"
	"errors"
	"fmt"
	"log-analyzer/internal/common"
	db "log-analyzer/internal/db"
	"log/slog"
	"time"
)

type TimingDetector struct{}

func (td TimingDetector) Check(tdb *db.TemplateDB, tmpl common.Template) ([]Anomaly, error) {
	_, mean, stddev, ts, err := tdb.GetIATStats(tmpl.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Anomaly{}, nil
		}
		return []Anomaly{}, fmt.Errorf("failed to get IAT Stats: %s", err)
	}

	// Not enough data to calculate IAT z score yet, or that its meaningless
	fmt.Println("mean", mean, "stddev", stddev)
	if stddev == 0 {
		return []Anomaly{}, nil
	}

	lastTs, err := time.Parse(db.TimestampFormat, ts)
	if err != nil {
		return []Anomaly{}, fmt.Errorf("failed to parse timestamp: %s", err)
	}

	iat := time.Since(lastTs).Seconds()
	z := (iat - mean) / stddev

	slog.Debug(fmt.Sprintf("Template: %s | Timing Z score: %f", tmpl.ID, z))

	sev := SeverityFromZScore(z)
	if sev > SeverityInfo {
		anomaly := Anomaly{TemplateID: tmpl.ID, Type: AnomalyTypeTiming, Severity: sev, Timestamp: time.Now()}
		anomaly.Description = fmt.Sprintf(
			"Abnormal latency detected for template %s: IAT deviates significantly from baseline (Z = %f)",
			tmpl.ID,
			z,
		)
		anomalies := []Anomaly{anomaly}
		return anomalies, nil
	}

	return []Anomaly{}, nil
}
