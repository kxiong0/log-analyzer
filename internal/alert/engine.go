package alert

import (
	"fmt"
	"log-analyzer/internal/anomaly"
	"time"
)

func NewAlertEngine() *AlertEngine {
	ae := AlertEngine{}

	buffer := make(map[BufferKey]anomaly.Anomaly)
	ae.buffer = buffer

	ae.AddAlertTarget(StdoutTarget{})
	return &ae
}

type AlertEngine struct {
	alertTargets []AlertTarget
	buffer       AnomalyBuffer
}

func (ae *AlertEngine) AddAlertTarget(at AlertTarget) {
	ae.alertTargets = append(ae.alertTargets, at)
}

func (ae *AlertEngine) AddAnomalies(as []anomaly.Anomaly) {
	for _, a := range as {
		ae.buffer.Add(a)
	}
}

func (ae *AlertEngine) Start(interval time.Duration, done <-chan bool) {
	ticker := time.NewTicker(interval)

	sendAlerts := func(anomalies []anomaly.Anomaly) {
		for _, at := range ae.alertTargets {
			at.Alert(anomalies)
		}
	}

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				anomalies := ae.buffer.Flush()
				sendAlerts(anomalies)
			case <-done:
				fmt.Println("Stopping flush scheduler...")
				anomalies := ae.buffer.Flush() // Final flush before stopping
				sendAlerts(anomalies)
				return
			}
		}

	}()
}
