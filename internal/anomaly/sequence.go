package anomaly

import (
	"fmt"
	"log-analyzer/internal/common"
	db "log-analyzer/internal/db"
	"log/slog"
	"time"
)

const (
	probabilityThreshold = 0.05
)

type SequenceDetector struct{}

func (sd SequenceDetector) Check(tdb *db.TemplateDB, tmpl common.Template) ([]Anomaly, error) {
	probability, err := tdb.GetTransitionProbability(tmpl.ID, tmpl.K8sMetadata.PodID)
	if err != nil {
		return nil, err
	}

	slog.Debug(fmt.Sprintf("Template: %s | Sequence probability: %f", tmpl.ID, probability))

	sev := SeverityInfo
	anomaly := Anomaly{TemplateID: tmpl.ID, Type: AnomalyTypeSequence, Severity: sev, Timestamp: time.Now()}
	if probability < probabilityThreshold {
		anomaly.Severity = SeverityMedium
		anomaly.Description = fmt.Sprintf("detected unusual transition of probability %f", probability)
	}
	return []Anomaly{anomaly}, nil
}
