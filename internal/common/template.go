package common

import (
	"strings"

	"github.com/google/uuid"
)

type Template struct {
	ID     string   // uuid
	Tokens []string // the canonical pattern: ["GET", "<NUM>", "users", "<UUID>"]
}

type TemplateTree map[int][]Template // key = token_count

func (tt TemplateTree) Find(tokens []string) (string, bool) {
	candidates := tt[len(tokens)]
	for _, tmpl := range candidates {
		if matchesTemplate(tokens, tmpl.Tokens) {
			return tmpl.ID, true
		}
	}
	return "", false
}

func (tt TemplateTree) Save(tokens []string) string {
	// Create new template and return its UUID
	t := Template{ID: uuid.NewString(), Tokens: tokens}
	tt[len(tokens)] = append(tt[len(tokens)], t)
	return t.ID
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
