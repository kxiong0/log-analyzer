package server

import (
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

	s := Server{
		lp: lp,
	}
	return &s, nil
}

type Server struct {
	lp *p.LogParser
}
