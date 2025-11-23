package parser

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

type MaskRule struct {
	Pattern *regexp.Regexp
	Token   string
}

// Precompiled regex rules in correct precedence order.
var rules = []MaskRule{
	// 1. Timestamps (allow optional fractional seconds up to nanoseconds)
	// ISO-8601
	{regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:[+-]\d{2}:\d{2})\b`), "<TIMESTAMP>"},
	// RFC3339
	{regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})\b`), "<TIMESTAMP>"},
	// YYYY/MM/DD HH:MM:SS with optional fractional seconds
	{regexp.MustCompile(`\b\d{4}/\d{2}/\d{2}[ T]\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?\b`), "<TIMESTAMP>"},
	// YYYY-MM-DD HH:MM:SS with optional fractional seconds
	{regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?\b`), "<TIMESTAMP>"},
	// Apache / Nginx style with optional fractional seconds
	{regexp.MustCompile(`\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}(?:\.\d{1,9})? [+-]\d{4}\]`), "<TIMESTAMP>"},

	// 2. URLs
	{regexp.MustCompile(`https?://[^\s]+`), "<URL>"},
	// 3. Paths
	{regexp.MustCompile(`(?:[A-Za-z]:)?(?:/[A-Za-z0-9._-]+)+/?`), "<PATH>"},
	// 4. IPv4 / IPv6
	{regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`), "<IP>"},
	{regexp.MustCompile(`\b(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}\b`), "<IP>"},
	// 5. UUIDs
	{regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}\b`), "<UUID>"},
	// 6. HEX strings (hashes, correlation IDs)
	{regexp.MustCompile(`\b[0-9a-fA-F]{8,}\b`), "<HEX>"},
	// 7. Email
	{regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`), "<EMAIL>"},
	// 8. Numbers
	{regexp.MustCompile(`\b\d+(?:\.\d+)?\b`), "<NUM>"},
	// 9. log levels
	{regexp.MustCompile(`(?i)\b(?:trace|debug|info|warn(?:ing)?|error|err|fatal|critical)\b`), "<LEVEL>"},
}

var logFieldAlias = []string{"message", "msg", "log"}

func ParseLog(s string) []string {
	// Try to parse incoming log as a JSON string
	rawLog, err := parseJsonLog(s)
	if err != nil {
		rawLog = string(s)
	}

	normalizedLog := normalize(rawLog)
	tokens := tokenize(normalizedLog)
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

// Split the given string by spaces, linebreaks, or punctuation marks
func tokenize(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == ';' || r == ':' || r == '|'
	})
	return fields
}

// Replace common values (e.g. IPs, timestamps) with tokens
func normalize(word string) string {
	for _, rule := range rules {
		word = rule.Pattern.ReplaceAllString(word, rule.Token)
	}
	return word
}
