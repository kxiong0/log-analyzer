package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	common "log-analyzer/internal/common"
)

var logFieldAlias = []string{"message", "msg", "log"}

func NewLogParser() *LogParser {
	return &LogParser{
		tt: make(common.TemplateTree),
	}
}

type LogParser struct {
	tt common.TemplateTree
}

func (lp LogParser) ParseLog(s string) []string {
	// Try to parse incoming log as a JSON string
	rawLog, err := parseJsonLog(s)
	if err != nil {
		rawLog = string(s)
	}

	log := preNormalize(rawLog)
	tokens := tokenize(log)
	for i, token := range tokens {
		tokens[i] = postNormalize(token)
	}

	tid, ok := lp.tt.Find(tokens)
	if !ok {
		tid = lp.tt.Save(tokens)
	}

	fmt.Println("template id of tokens:", tid)
	return tokens
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
