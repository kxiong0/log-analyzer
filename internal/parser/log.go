package parser

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	common "log-analyzer/internal/common"
	db "log-analyzer/internal/db"
)

var logFieldAlias = []string{"message", "msg", "log"}

const (
	databaseFile = "data.db"
)

func NewLogParser() (*LogParser, error) {
	lp := &LogParser{
		tt: make(common.TemplateTree),
	}

	// Init db
	tdb, err := db.NewTemplateDB(databaseFile)
	if err != nil {
		return nil, err
	}
	lp.tdb = tdb
	lp.LoadTemplates()

	return lp, nil
}

type LogParser struct {
	tt  common.TemplateTree
	tdb *db.TemplateDB
}

func (lp LogParser) ParseLog(s string) string {
	// Try to parse incoming log as a JSON string
	// Returns template ID of the parsed log
	rawLog, err := parseJsonLog(s)
	if err != nil {
		rawLog = string(s)
	}

	log := preNormalize(rawLog)
	tokens := tokenize(log)
	for i, token := range tokens {
		tokens[i] = postNormalize(token)
	}

	// Create new template if no template found
	tid, ok := lp.tt.Find(tokens)
	if !ok {
		tid = lp.tt.Save(tokens)
		lp.tdb.SaveTemplate(common.Template{ID: tid, Tokens: tokens})
	}

	return tid
}

// Try to parse a string as a json string and return the raw log line
func parseJsonLog(s string) (string, error) {
	var jsonLog map[string]interface{}
	err := json.Unmarshal([]byte(s), &jsonLog)
	if err != nil {
		return "", err
	}

	for _, alias := range logFieldAlias {
		m, ok := jsonLog[alias].(string)
		if ok {
			return m, nil
		}
	}
	return "", errors.New("no log field found in JSON")
}

func preNormalize(s string) string {
	for _, rule := range preTokenizeRules {
		s = rule.Pattern.ReplaceAllString(s, rule.Token)
	}
	return s
}

// Split the given string by spaces, linebreaks, or punctuation marks
func tokenize(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == ';' || r == ':' || r == '|'
	})
	return fields
}

// Replace common values with tokens
func postNormalize(s string) string {
	for _, rule := range postTokenizeRules {
		s = rule.Pattern.ReplaceAllString(s, rule.Token)
	}
	return s
}

// Load templates from DB
func (lp *LogParser) LoadTemplates() error {
	slog.Debug("Loading templates from DB...")

	templates, err := lp.tdb.GetAllTemplates()
	if err != nil {
		slog.Error("Failed to get templates from DB: ", slog.Any("error", err))
		return err
	}

	for _, t := range templates {
		lp.tt[len(t.Tokens)] = append(lp.tt[len(t.Tokens)], t)
	}
	return nil
}
