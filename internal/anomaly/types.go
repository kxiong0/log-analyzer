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

func SeverityFromScore(score float64) Severity {
	switch {
	case score < 0.25:
		return SeverityInfo
	case score < 0.5:
		return SeverityLow
	case score < 0.7:
		return SeverityMedium
	case score < 0.9:
		return SeverityHigh
	default:
		return SeverityCritical
	}
}

type AnomalyDetector interface {
	Check(tdb *db.TemplateDB, tid string) ([]Anomaly, error)
}
