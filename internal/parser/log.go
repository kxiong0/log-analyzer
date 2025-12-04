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

func NewLogParser(tdb *db.TemplateDB) (*LogParser, error) {
	lp := &LogParser{}
	lp.tdb = tdb

	// Fetch template tree from DB
	lp.LoadTemplates()

	return lp, nil
}

type LogParser struct {
	tt  common.TemplateTree
	tdb *db.TemplateDB
}

// Try to parse incoming log as a JSON string
// Returns template of the parsed log and if a new template is created
func (lp *LogParser) ParseLog(s string) (tmpl common.Template, newTmpl bool) {
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
	tmpl, ok := lp.tt.Find(tokens)
	if !ok {
		tmpl = lp.tt.Save(tokens)
		lp.tdb.SaveTemplate(tmpl)
	}
	return tmpl, !ok
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

// Load templates from DB and store in TemplateTree
func (lp *LogParser) LoadTemplates() error {
	slog.Debug("Loading templates from DB...")

	templates, err := lp.tdb.GetAllTemplates()
	if err != nil {
		slog.Error("Failed to get templates from DB: ", slog.Any("error", err))
		return err
	}

	slog.Debug("Successfully loaded template tree from DB")
	lp.tt = templates
	return nil
}
