package parser

import "regexp"

type MaskRule struct {
	Pattern *regexp.Regexp
	Token   string
}

// Precompiled regex rules in correct precedence order.
var preTokenizeRules = []MaskRule{
	// 1. Timestamps (allow optional fractional seconds up to nanoseconds)
	// ISO-8601
	{regexp.MustCompile(`\[?\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:[+-]\d{2}:\d{2})\]?`), "<TIMESTAMP>"},
	// RFC3339
	{regexp.MustCompile(`\[?\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?(?:Z|[+-]\d{2}:\d{2})\]?`), "<TIMESTAMP>"},
	// YYYY/MM/DD HH:MM:SS with optional fractional seconds
	{regexp.MustCompile(`\[?\d{4}/\d{2}/\d{2}[ T]\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?\]?`), "<TIMESTAMP>"},
	// YYYY-MM-DD HH:MM:SS with optional fractional seconds
	{regexp.MustCompile(`\[?\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:\.\d{1,9})?\]?`), "<TIMESTAMP>"},
	// Apache / Nginx style with optional fractional seconds
	{regexp.MustCompile(`\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}(?:\.\d{1,9})? [+-]\d{4}\]`), "<TIMESTAMP>"},
	// log levels
	{regexp.MustCompile(`\[? *(?i)(?:trace|debug|info|warn(?:ing)?|error|err|fatal|critical) *\]?`), "<LEVEL>"},
	// 2. URLs
	{regexp.MustCompile(`https?://[^\s]+`), "<URL>"},
}

var postTokenizeRules = []MaskRule{
	// 3. Paths
	{regexp.MustCompile(`(?:[A-Za-z]:)?(?:/[A-Za-z0-9._-]+)+/?`), "<PATH>"},
	// 4. IPv4 / IPv6
	{regexp.MustCompile(`(?:\d{1,3}\.){3}\d{1,3}`), "<IP>"},
	{regexp.MustCompile(`(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}`), "<IP>"},
	// 5. UUIDs
	{regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}`), "<UUID>"},
	// 6. HEX strings (hashes, correlation IDs)
	{regexp.MustCompile(`[0-9a-fA-F]{8,}`), "<HEX>"},
	// 7. Email
	{regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`), "<EMAIL>"},
	// 8. Numbers
	{regexp.MustCompile(`\d+(?:\.\d+)?`), "<NUM>"},
}
