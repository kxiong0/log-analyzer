package server

import (
	"log-analyzer/detectors/frequency"
	"log-analyzer/detectors/sequence"
	"log-analyzer/detectors/timing"

	"log-analyzer/internal/alert"
	"log-analyzer/internal/anomaly"
	"log-analyzer/internal/db"
	p "log-analyzer/internal/parser"
	"time"
)

const (
	databaseFile = "data.db"
)

func NewServer() (*Server, error) {
	tdb, err := db.NewTemplateDB(databaseFile)
	if err != nil {
		return nil, err
	}

	lp, err := p.NewLogParser(tdb)
	if err != nil {
		return nil, err
	}

	ae, err := anomaly.NewAnomalyEngine(tdb)
	if err != nil {
		return nil, err
	}
	ae.AddAnomalyDetector(&frequency.FrequencyDetector{})
	ae.AddAnomalyDetector(&sequence.SequenceDetector{})
	ae.AddAnomalyDetector(&timing.TimingDetector{})

	ale := alert.NewAlertEngine()
	done := make(<-chan bool)

	ale.Start(time.Second*5, done)
	ae.Start(done)

	s := Server{
		lp:  lp,
		ae:  ae,
		ale: ale,
	}
	return &s, nil
}

type Server struct {
	lp  *p.LogParser
	ae  *anomaly.AnomalyEngine
	ale *alert.AlertEngine
}
