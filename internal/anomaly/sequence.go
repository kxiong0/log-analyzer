package anomaly

import (
	"fmt"
	"log-analyzer/internal/common"
	db "log-analyzer/internal/db"
	"log/slog"
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

	if probability < probabilityThreshold {
		return []Anomaly{{
				TemplateID:  tmpl.ID,
				Type:        AnomalyTypeSequence,
				Severity:    SeverityMedium,
				Description: fmt.Sprintf("detected unusual transition of probability %f", probability),
			}},
			nil
	}
	return []Anomaly{}, nil
}
