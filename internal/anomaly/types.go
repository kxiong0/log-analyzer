package anomaly

import (
	"fmt"
	"log-analyzer/internal/db"
	"time"
)

type Anomaly struct {
	TemplateID  string
	Type        string
	Severity    Severity
	Description string
	Timestamp   time.Time
}

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

func SeverityFromZScore(score float64) Severity {
	sev := SeverityInfo
	switch {
	case score >= 5.0:
		sev = SeverityCritical
	case score >= 4.0:
		sev = SeverityHigh
	case score >= 3.0:
		sev = SeverityMedium
	case score >= 2.0:
		sev = SeverityLow
	}
	return sev
}

type AnomalyDetector interface {
	Check(tdb *db.TemplateDB, tid string) ([]Anomaly, error)
}
