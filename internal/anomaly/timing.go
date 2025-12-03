package anomaly

import (
	"fmt"
	db "log-analyzer/internal/db"
	"log/slog"
	"time"
)

type TimingDetector struct{}

func (td TimingDetector) Check(tdb *db.TemplateDB, tid string) ([]Anomaly, error) {
	_, mean, stddev, ts, err := tdb.GetIATStats(tid)
	if err != nil {
		return []Anomaly{}, err
	}

	lastTs, err := time.Parse(db.TimestampFormat, ts)
	if err != nil {
		return []Anomaly{}, err
	}

	iat := time.Since(lastTs).Seconds()
	z := (iat - mean) / stddev

	slog.Debug(fmt.Sprintf("Template: %s | Timing Z score: %f", tid, z))

	sev := SeverityFromZScore(z)
	if sev > SeverityInfo {
		anomaly := Anomaly{TemplateID: tid, Type: "timing", Timestamp: time.Now()}
		anomaly.Description = fmt.Sprintf("Abnormal latency spike detected for template %s: IAT deviates significantly from baseline (Z = %f)", tid, z)
		anomalies := []Anomaly{anomaly}
		return anomalies, nil
	}

	return []Anomaly{}, nil
}
