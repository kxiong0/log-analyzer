package server

import (
	"log-analyzer/internal/anomaly"
	"log-analyzer/internal/db"
	p "log-analyzer/internal/parser"
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

	s := Server{
		lp: lp,
		ae: ae,
	}
	return &s, nil
}

type Server struct {
	lp *p.LogParser
	ae *anomaly.AnomalyEngine
}
