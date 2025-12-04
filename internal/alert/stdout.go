package alert

import (
	"fmt"
	"log-analyzer/internal/anomaly"
)

type StdoutTarget struct{}

func (st StdoutTarget) Alert(anomalies []anomaly.Anomaly) (ok bool) {
	for _, a := range anomalies {
		fmt.Printf("stdout alert: %s | Type %s | Template %s | Since: %s | %s\n", a.Severity, a.Type, a.TemplateID, a.Timestamp.Format("2006-01-02 15:04:05"), a.Description)
	}
	return true
}
