// Package opencodeagents defines the OpenCode-compatible agents owned by gentle-ai.
package opencodeagents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Santiago1809/teicos/internal/assets"
	"github.com/Santiago1809/teicos/internal/model"
)

type Spec struct {
	Name        string
	Description string
	Permission  map[string]any
}

var parity = []Spec{
	{"gentle-ai-explore", "Read-only exploration and mapping for generic ODD work.", map[string]any{"write": "deny", "edit": "deny", "bash": "deny", "task": "deny"}},
	{"gentle-ai-verify", "Read-only technical verification for generic ODD work.", map[string]any{"write": "deny", "edit": "deny", "task": "deny"}},
	{"gentle-ai-worker", "Scoped package-owned implementation writer for bounded ODD work. Edits code, runs focused tests, and returns review-ready evidence without committing.", map[string]any{"task": "deny"}},
	{"jd-judge-a", "Judgment Day blind adversarial reviewer A. Read-only; reports findings and does not fix code.", map[string]any{"write": "deny", "edit": "deny", "task": "deny"}},
	{"jd-judge-b", "Judgment Day blind adversarial reviewer B. Read-only; independently reports findings and does not fix code.", map[string]any{"write": "deny", "edit": "deny", "task": "deny"}},
	{"jd-fix-agent", "Judgment Day surgical fix agent for confirmed findings. Can edit code and run focused tests.", map[string]any{"task": "deny"}},
}

func Parity(agent model.AgentID) []Spec {
	if agent == model.AgentOpenCode {
		return append([]Spec(nil), parity...)
	}
	var specs []Spec
	for _, spec := range parity {
		if strings.HasPrefix(spec.Name, "gentle-ai-") {
			continue
		}
		specs = append(specs, spec)
	}
	return specs
}

func Roles(agent model.AgentID) []string {
	var names []string
	if agent == model.AgentOpenCode {
		names = append(names, "review-validator")
	}
	for _, spec := range Parity(agent) {
		names = append(names, spec.Name)
	}
	return names
}

// LegacyOwned reports the v3.7.0 marked roles the runtime installed or
// explicitly retired during migration. Unknown marked names remain user data.
func LegacyOwned(agent model.AgentID, name string) bool {
	if name == "gentle-orchestrator" || name == "general" || name == "explore" || strings.HasPrefix(name, "sdd-") {
		return true
	}
	for _, role := range Roles(agent) {
		if role == name {
			return true
		}
	}
	return false
}

// UninstallRole recognizes current runtime roles and retired Kilo review roles.
// Shape checks remain necessary before removing any unmarked entry.
func UninstallRole(agent model.AgentID, name string) bool {
	for _, role := range Roles(agent) {
		if role == name {
			return true
		}
	}
	return false
}

func Entry(spec Spec) (map[string]any, error) {
	prompt, err := assets.Read("opencode/agents/" + spec.Name + ".md")
	if err != nil {
		return nil, fmt.Errorf("read embedded prompt for %q: %w", spec.Name, err)
	}
	return map[string]any{"mode": "subagent", "hidden": true, "description": spec.Description, "prompt": prompt, "permission": spec.Permission}, nil
}

func Refuter() map[string]any {
	return map[string]any{"mode": "subagent", "hidden": true, "description": "Read-only refuter for provider-issued review findings", "prompt": "Evaluate only the Go-issued review refuter task. Inspect only the frozen candidate through the provided commands. Do not edit files or delegate. Return only the requested result.", "permission": map[string]any{"write": "deny", "edit": "deny", "task": "deny"}}
}
func Validator() map[string]any {
	return map[string]any{"mode": "subagent", "hidden": true, "description": "Targeted read-only validator for provider-issued review checks", "prompt": "Execute only the Go-issued targeted validation. Do not edit files or delegate. Inspect only the frozen candidate using the provided gentle-ai review inspect-candidate command, not the live worktree. Return exactly the requested JSON.", "permission": map[string]any{"write": "deny", "edit": "deny", "task": "deny", "bash": map[string]any{"gentle-ai review inspect-candidate --purpose targeted-validation *": "allow", "*": "deny"}}}
}

// GentlemanShape recognizes the persona overlay for this runtime, without
// claiming a user-modified agent under the same name.
func GentlemanShape(agent model.AgentID, entry map[string]any) (bool, error) {
	want := map[string]any{"mode": "primary", "description": "Senior Architect mentor - helpful first, challenging when it matters", "prompt": "{file:./AGENTS.md}"}
	if agent == model.AgentKilocode {
		want["tools"] = map[string]any{"write": true, "edit": true}
	}
	comparable := make(map[string]any, len(entry))
	for k, v := range entry {
		if k != "model" && k != "variant" {
			comparable[k] = v
		}
	}
	got, err := json.Marshal(comparable)
	if err != nil {
		return false, err
	}
	expected, err := json.Marshal(want)
	if err != nil {
		return false, err
	}
	return bytes.Equal(got, expected), nil
}

// Shape accepts user-selected model and variant, but no other deviations.
func Shape(name string, entry map[string]any) (bool, error) {
	var want map[string]any
	for _, spec := range parity {
		if spec.Name == name {
			var err error
			want, err = Entry(spec)
			if err != nil {
				return false, err
			}
			break
		}
	}
	if want == nil {
		return false, nil
	}
	comparable := make(map[string]any, len(entry))
	for k, v := range entry {
		if k != "model" && k != "variant" {
			comparable[k] = v
		}
	}
	got, err := json.Marshal(comparable)
	if err != nil {
		return false, err
	}
	expected, err := json.Marshal(want)
	if err != nil {
		return false, err
	}
	return bytes.Equal(got, expected), nil
}
