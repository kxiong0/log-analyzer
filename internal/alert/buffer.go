package alert

import "log-analyzer/internal/anomaly"

const (
	alertThreshold = anomaly.SeverityMedium
)

type AnomalyBuffer map[BufferKey]anomaly.Anomaly

func (ab *AnomalyBuffer) Add(a anomaly.Anomaly) {
	bk := BufferKey{AnomalyType: a.Type, TemplateID: a.TemplateID}
	prevA, ok := (*ab)[bk]

	// Update Severity / Description but keep timestamp
	if ok {
		prevA.Severity = a.Severity
		prevA.Description = a.Description
		(*ab)[bk] = prevA
		return
	}

	if a.Severity >= alertThreshold {
		(*ab)[bk] = a
		return
	}

}

// Return a list of resolved and active anomalies.
// Resolved anomalies are marked as resolved and removed from the buffer
func (ab *AnomalyBuffer) Flush() (anomalies []anomaly.Anomaly) {
	for k, v := range *ab {
		if v.Severity < alertThreshold {
			v.Severity = anomaly.SeverityResolved
			delete(*ab, k)
		}
		anomalies = append(anomalies, v)
	}
	return
}
