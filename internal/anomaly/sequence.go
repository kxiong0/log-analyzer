package anomaly

import (
	"fmt"
	db "log-analyzer/internal/db"
	"log/slog"
)

const (
	probabilityThreshold = 0.05
)

type SequenceDetector struct{}

func (sd SequenceDetector) Check(tdb *db.TemplateDB, tid string) ([]Anomaly, error) {
	probability, err := tdb.GetTransitionProbability(tid)
	if err != nil {
		return nil, err
	}

	slog.Debug(fmt.Sprintf("Template: %s | Sequence probability: %f", tid, probability))

	if probability < probabilityThreshold {
		return []Anomaly{{
				TemplateID:  tid,
				Type:        "sequence",
				Severity:    SeverityMedium,
				Description: fmt.Sprintf("detected unusual transition of probability %f", probability),
			}},
			nil
	}
	return []Anomaly{}, nil
}
