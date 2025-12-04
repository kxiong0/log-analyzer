package alert

import "log-analyzer/internal/anomaly"

type AlertTarget interface {
	Alert(anomalies []anomaly.Anomaly) (ok bool)
}

type BufferKey struct {
	AnomalyType anomaly.AnomalyType
	TemplateID  string
}
