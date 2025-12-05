package anomaly

import (
	"fmt"
	"log-analyzer/internal/common"
	"log-analyzer/internal/db"
	"time"
)

type AnomalyDetector interface {
	Init(tdb *db.TemplateDB) error
	Start(done <-chan bool) error                  // done to send a signal to start clean up
	Check(tmpl common.Template) ([]Anomaly, error) // called each time a template is ingested
}

type Anomaly struct {
	TemplateID  string
	Type        AnomalyType
	Severity    Severity
	Description string
	Timestamp   time.Time
}

type AnomalyType int

const (
	AnomalyTypeNewTemplate AnomalyType = iota
	AnomalyTypeFrequency
	AnomalyTypeSequence
	AnomalyTypeTiming
)

func (at AnomalyType) String() string {
	switch at {
	case AnomalyTypeNewTemplate:
		return "New Template"
	case AnomalyTypeFrequency:
		return "Frequency"
	case AnomalyTypeSequence:
		return "Sequence"
	case AnomalyTypeTiming:
		return "Timing"
	default:
		return fmt.Sprintf("unknown(%d)", at)
	}
}

type Severity int

const (
	SeverityResolved Severity = iota
	SeverityInfo
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityResolved:
		return "resolved"
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
