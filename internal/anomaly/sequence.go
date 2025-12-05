package anomaly

import (
	"fmt"
	"log-analyzer/internal/common"
	db "log-analyzer/internal/db"
	"log/slog"
	"time"
)

const (
	warmupThreshold      = 10
	probabilityThreshold = 0.05
)

type SequenceDetector struct {
	tdb *db.TemplateDB
}

func (sd *SequenceDetector) Init(tdb *db.TemplateDB) error {
	sd.tdb = tdb
	return nil
}

func (sd SequenceDetector) Start(done <-chan bool) error {
	return nil
}

func (sd SequenceDetector) Check(tmpl common.Template) ([]Anomaly, error) {
	total, tran, err := sd.tdb.GetTransitionCounts(tmpl.ID, tmpl.K8sMetadata.PodID)
	if err != nil {
		return nil, err
	}

	sev := SeverityInfo
	anomaly := Anomaly{TemplateID: tmpl.ID, Type: AnomalyTypeSequence, Severity: sev, Timestamp: time.Now()}

	// Skip check if not enough total transitions recorded
	if total < warmupThreshold {
		anomaly.Description = fmt.Sprintf("skipping: less than %d total transitions exist for template %s", warmupThreshold, tmpl.ID)
		return []Anomaly{anomaly}, nil
	}

	probability := float64(tran) / float64(total)
	slog.Debug(fmt.Sprintf("Template: %s | Sequence probability: %f", tmpl.ID, probability))

	if probability < probabilityThreshold {
		anomaly.Severity = SeverityMedium
		anomaly.Description = fmt.Sprintf("detected unusual transition of probability %f", probability)
	}
	return []Anomaly{anomaly}, nil
}
