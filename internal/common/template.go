package common

import "strings"

type Template struct {
	ID         uint32
	Tokens     []string // the canonical pattern: ["GET", "<NUM>", "users", "<UUID>"]
	RawPattern string   // optional human-readable pattern
}

type TemplateTree map[int][]Template // key = token_count

func (tt TemplateTree) Find(tokens []string) (uint32, bool) {
	candidates := tt[len(tokens)]
	for _, tmpl := range candidates {
		if matchesTemplate(tokens, tmpl.Tokens) {
			return tmpl.ID, true
		}
	}
	return 0, false
}

func (tt TemplateTree) NewTemplate(tokens []string) {
	// TODO
}

func matchesTemplate(tokens, tmpl []string) bool {
	if len(tokens) != len(tmpl) {
		return false
	}
	for i := range tokens {
		if tmpl[i] == tokens[i] {
			continue
		}
		// Wildcards match anything
		if strings.HasPrefix(tmpl[i], "<") && strings.HasSuffix(tmpl[i], ">") {
			continue
		}
		return false
	}
	return true
}
