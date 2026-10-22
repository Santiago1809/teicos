package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Santiago1809/teicos/internal/components/filemerge"
	"github.com/Santiago1809/teicos/internal/system"
)

// kiloParityAgentNames is the Kilo agent set: the v3.7.0 Judgment Day roles
// (#4471). Kilo never received review-validator (it hosts no provider relay)
// nor the gentle-ai-* ODD trio (its orchestrator routes to native
// delegation), and it receives no review agent at all because receipt-driven
// development ships only to its own runtimes.
var kiloParityAgentNames = []string{
	"jd-judge-a", "jd-judge-b", "jd-fix-agent",
}

// kiloRetiredReviewAgentNames are the RDD agents earlier releases installed on
// Kilo; upgrades remove the entries Gentle AI owns.
var kiloRetiredReviewAgentNames = []string{
	"review-risk", "review-readability", "review-reliability", "review-resilience",
	"review-refuter", "review-validator",
}

func kiloSettingsPath(home string) string {
	return filepath.Join(home, ".config", "kilo", "opencode.json")
}

func kiloAgents(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	root, err := filemerge.UnmarshalJSONObject(raw)
	if err != nil {
		t.Fatalf("decode Kilo settings: %v", err)
	}
	agents, _ := root["agent"].(map[string]any)
	return agents
}

func assertKiloParityAgents(t *testing.T, agents map[string]any) {
	t.Helper()
	for _, name := range kiloParityAgentNames {
		entry, ok := agents[name].(map[string]any)
		if !ok {
			t.Fatalf("Kilo settings missing agent %q; got %v", name, agents)
		}
		if entry["mode"] != "subagent" || entry["hidden"] != true {
			t.Fatalf("Kilo agent %q missing mode/hidden: %#v", name, entry)
		}
		if prompt, _ := entry["prompt"].(string); strings.TrimSpace(prompt) == "" || strings.Contains(prompt, "obsolete") {
			t.Fatalf("Kilo agent %q has no current prompt: %#v", name, entry)
		}
		if _, ok := entry["permission"].(map[string]any); !ok {
			t.Fatalf("Kilo agent %q missing permission: %#v", name, entry)
		}
	}
	for _, name := range append([]string{"gentle-ai-explore", "gentle-ai-verify", "gentle-ai-worker"}, kiloRetiredReviewAgentNames...) {
		if _, ok := agents[name]; ok {
			t.Fatalf("Kilo must not install %q", name)
		}
	}
	for name, raw := range agents {
		if strings.HasPrefix(name, "sdd-") {
			// SDD agents belong to gentle-ai: they keep their ownership
			// marker so the two products can share the same agent settings.
			continue
		}
		entry, _ := raw.(map[string]any)
		if _, marked := entry["__managed_by"]; marked {
			t.Fatalf("Kilo agent %q still carries __managed_by (#4471): %#v", name, entry)
		}
	}
	orchestrator, _ := agents["gentle-orchestrator"].(map[string]any)
	permission, _ := orchestrator["permission"].(map[string]any)
	// Kilo shares the OpenCode permission defaults and the orchestrator prompt
	// that relays blocking prompts through the native question tool (#4816).
	if permission["question"] != "allow" {
		t.Fatalf("Kilo gentle-orchestrator permission.question = %#v, want \"allow\"", permission["question"])
	}
	task, _ := permission["task"].(map[string]any)
	for _, name := range kiloParityAgentNames {
		if task[name] != "allow" {
			t.Fatalf("Kilo gentle-orchestrator does not allow delegating to %q: %#v", name, task)
		}
	}
	for _, name := range kiloRetiredReviewAgentNames {
		if _, ok := task[name]; ok {
			t.Fatalf("Kilo gentle-orchestrator still delegates to RDD agent %q: %#v", name, task)
		}
	}
}

func TestKiloInstallWritesParityAgentsWithoutManagedByMarker(t *testing.T) {
	home := installTestHome(t)
	args := []string{"--agent", "kilocode", "--preset", "full-gentleman"}
	if _, err := RunInstall(args, system.DetectionResult{}); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(kiloSettingsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	agents := kiloAgents(t, first)
	assertKiloParityAgents(t, agents)
	for _, name := range kiloParityAgentNames {
		for key := range agents[name].(map[string]any) {
			if !openCodeAgentConfigAllowedKeys[key] {
				t.Fatalf("Kilo agent %q carries non-AgentConfig key %q: %#v", name, key, agents[name])
			}
		}
	}
	if _, err := RunInstall(args, system.DetectionResult{}); err != nil {
		t.Fatal(err)
	}
	if second, _ := os.ReadFile(kiloSettingsPath(home)); !bytes.Equal(first, second) {
		t.Fatalf("second Kilo install changed settings:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// TestKiloReviewTaskPermissionCleanupRequiresOwnershipProof proves install and
// sync keeps an orchestrator task permission for a review agent gentle-ai
// owns: the review agents belong to that product, not to Teicos, so neither
// the entry nor its permission is removed here. A permission whose agent was
// already absent is the user's and survives too.
func TestKiloReviewTaskPermissionCleanupRequiresOwnershipProof(t *testing.T) {
	runs := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"install", func(t *testing.T) {
			t.Helper()
			if _, err := RunInstall([]string{"--agent", "kilocode", "--preset", "full-gentleman"}, system.DetectionResult{}); err != nil {
				t.Fatal(err)
			}
		}},
		{"sync", func(t *testing.T) {
			t.Helper()
			if _, err := RunSync([]string{"--agent", "kilocode"}); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range []struct {
		name      string
		agents    map[string]any
		wantTask  bool
		wantAgent bool
	}{
		{"user permission without agent entry survives", map[string]any{}, true, false},
		{"a review permission without its agent entry survives", map[string]any{}, true, false},
		{"a legacy marked review agent keeps its entry and permission", map[string]any{
			"review-risk": map[string]any{"__managed_by": "gentle-ai/sdd", "mode": "subagent", "prompt": "legacy"},
		}, true, true},
	} {
		for _, run := range runs {
			t.Run(tc.name+"/"+run.name, func(t *testing.T) {
				home := installTestHome(t)
				path := kiloSettingsPath(home)
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				agents := map[string]any{
					"gentle-orchestrator": map[string]any{"permission": map[string]any{"task": map[string]any{"review-risk": "allow"}}},
				}
				for name, entry := range tc.agents {
					agents[name] = entry
				}
				seed, err := json.MarshalIndent(map[string]any{"agent": agents}, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, seed, 0o600); err != nil {
					t.Fatal(err)
				}

				run.run(t)
				first, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				after := kiloAgents(t, first)
				if _, got := after["review-risk"]; got != tc.wantAgent {
					t.Errorf("review-risk agent present = %v, want %v:\n%s", got, tc.wantAgent, first)
				}
				orchestrator, _ := after["gentle-orchestrator"].(map[string]any)
				permission, _ := orchestrator["permission"].(map[string]any)
				task, _ := permission["task"].(map[string]any)
				if _, got := task["review-risk"]; got != tc.wantTask {
					t.Errorf("review-risk task permission present = %v, want %v:\n%s", got, tc.wantTask, first)
				}
				run.run(t)
				if second, _ := os.ReadFile(path); !bytes.Equal(first, second) {
					t.Errorf("second %s changed settings bytes:\nfirst:\n%s\nsecond:\n%s", run.name, first, second)
				}
			})
		}
	}
}
