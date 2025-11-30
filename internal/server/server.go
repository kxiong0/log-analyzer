package server

import (
	"log-analyzer/internal/anomaly"
	p "log-analyzer/internal/parser"
)

const (
	outputFile = "access.log"
)

func NewServer() (*Server, error) {
	lp, err := p.NewLogParser()
	if err != nil {
		return nil, err
	}

	ae, err := anomaly.NewAnomalyEngine()
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
