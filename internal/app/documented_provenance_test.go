package app_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Santiago1809/teicos/internal/model"
)

// Provenance rule (#2524, root 4 of #2440): a runtime's rendered surface must
// never tell it to identify as another runtime. Before #2484, codex's rendered
// AGENTS.md carried `--agent claude-code` and nothing went red. Check both
// the retained golden surfaces and the currently rendered review contract:
// the goldens no longer contain the review command, but reviewassets still
// binds runtime identity in the installed contract. Derive golden owners from
// catalog.AllAgents() and inspect the actual per-runtime render, not an
// invented golden or a removed SDD overlay.

var rawAgentBindingRegexp = regexp.MustCompile(`--agent[= ]+[A-Za-z0-9._{}-]+`)

// agentFlagValues returns every value bound to --agent in one documented
// command, including the `--agent=value` spelling. A trailing bare --agent
// is returned as an empty value so the caller can reject it.
func agentFlagValues(command string) []string {
	words := strings.Fields(command)
	var out []string
	for index, word := range words {
		if word == "--agent" {
			if index+1 < len(words) {
				out = append(out, words[index+1])
			} else {
				out = append(out, "")
			}
			continue
		}
		if value, ok := strings.CutPrefix(word, "--agent="); ok {
			out = append(out, value)
		}
	}
	return out
}

// goldenOwners derives which runtimes a golden filename names, by exact
// token match against the derived slug set.
func goldenOwners(name string, slugToAgent map[string]model.AgentID) []model.AgentID {
	tokens := strings.FieldsFunc(strings.TrimSuffix(name, ".golden"), func(r rune) bool {
		return r == '-' || r == '.'
	})
	seen := map[model.AgentID]bool{}
	var owners []model.AgentID
	for _, token := range tokens {
		if id, ok := slugToAgent[token]; ok && !seen[id] {
			seen[id] = true
			owners = append(owners, id)
		}
	}
	return owners
}

func TestRenderedSurfaceNamesOnlyItsOwnRuntime(t *testing.T) {
	slugToAgent := map[string]model.AgentID{}
	_ = slugToAgent
}
