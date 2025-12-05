package timing

import (
	"database/sql"
	"errors"
	"fmt"
	"log-analyzer/internal/anomaly"
	"log-analyzer/internal/common"
	"log-analyzer/internal/db"
	"log/slog"
	"time"
)

const (
	warmupThreshold = 10
	timingCVFilter  = 0.6 // stddev / mean threshold
)

type TimingDetector struct {
	tdb *db.TemplateDB
}

func (td *TimingDetector) Init(tdb *db.TemplateDB) error {
	td.tdb = tdb
	return nil
}

func (td TimingDetector) Start(done <-chan bool) error {
	return nil
}

func (td TimingDetector) Check(tmpl common.Template) ([]anomaly.Anomaly, error) {
	count, mean, stddev, ts, err := td.tdb.GetIATStats(tmpl.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []anomaly.Anomaly{}, nil
		}
		return []anomaly.Anomaly{}, fmt.Errorf("failed to get IAT Stats: %s", err)
	}

	// Not enough data to calculate IAT z score yet, or that its meaningless
	if count < warmupThreshold || stddev == 0 {
		return []anomaly.Anomaly{}, nil
	}

	// Stats are too unstable to create alert
	if stddev/mean > timingCVFilter {
		return []anomaly.Anomaly{}, nil
	}

	lastTs, err := time.Parse(db.TimestampFormat, ts)
	if err != nil {
		return []anomaly.Anomaly{}, fmt.Errorf("failed to parse timestamp: %s", err)
	}

	iat := time.Since(lastTs).Seconds()
	z := (iat - mean) / stddev

	slog.Debug(fmt.Sprintf("Template: %s | Timing Z score: %f", tmpl.ID, z))

	sev := anomaly.SeverityFromZScore(z)
	a := anomaly.Anomaly{TemplateID: tmpl.ID, Type: anomaly.AnomalyTypeTiming, Severity: sev, Timestamp: time.Now()}
	if sev > anomaly.SeverityInfo {
		a.Description = fmt.Sprintf(
			"Abnormal latency detected for template %s: IAT deviates significantly from baseline (Z = %f)",
			tmpl.ID,
			z,
		)
	}
	return []anomaly.Anomaly{a}, nil
}
