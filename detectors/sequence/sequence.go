package sequence

import (
	"fmt"
	"log-analyzer/internal/anomaly"
	"log-analyzer/internal/common"
	"log-analyzer/internal/db"
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

func (sd SequenceDetector) Check(tmpl common.Template) ([]anomaly.Anomaly, error) {
	total, tran, err := sd.tdb.GetTransitionCounts(tmpl.ID, tmpl.K8sMetadata.PodID)
	if err != nil {
		return nil, err
	}

	sev := anomaly.SeverityInfo
	a := anomaly.Anomaly{TemplateID: tmpl.ID, Type: anomaly.AnomalyTypeSequence, Severity: sev, Timestamp: time.Now()}

	// Skip check if not enough total transitions recorded
	if total < warmupThreshold {
		a.Description = fmt.Sprintf("skipping: less than %d total transitions exist for template %s", warmupThreshold, tmpl.ID)
		return []anomaly.Anomaly{}, nil
	}

	probability := float64(tran) / float64(total)
	slog.Debug(fmt.Sprintf("Template: %s | Sequence probability: %f", tmpl.ID, probability))

	if probability < probabilityThreshold {
		a.Severity = anomaly.SeverityMedium
		a.Description = fmt.Sprintf("detected unusual transition of probability %f", probability)
	}
	return []anomaly.Anomaly{a}, nil
}
