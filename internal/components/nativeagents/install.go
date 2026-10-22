package nativeagents

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Santiago1809/teicos/internal/agents"
	"github.com/Santiago1809/teicos/internal/assets"
	"github.com/Santiago1809/teicos/internal/components/agentguidance"
	"github.com/Santiago1809/teicos/internal/components/filemerge"
	"github.com/Santiago1809/teicos/internal/components/mutationjournal"
	"github.com/Santiago1809/teicos/internal/model"
)

// NativeAgentManifest is an explicit allowlist: never enumerate the embedded agents directory.
//
// It carries the Judgment Day agents and Kimi's main agent. The review lenses
// and the refuter belonged to receipt-driven development and are gone.
var NativeAgentManifest = map[model.AgentID][]string{
	model.AgentClaudeCode: {"jd-fix-agent.md", "jd-judge-a.md", "jd-judge-b.md"},
	model.AgentKiroIDE:    {"jd-fix-agent.md", "jd-judge-a.md", "jd-judge-b.md"},
	model.AgentCursor:     {"jd-fix-agent.md", "jd-judge-a.md", "jd-judge-b.md"},
	model.AgentKimi:       {"gentleman.yaml"},
}

var engramToolPlaceholder = regexp.MustCompile(`\{\{ENGRAM_TOOL_PREFIX\}\}([A-Za-z0-9_]+)`)

type kiroModelResolver interface {
	KiroModelID(model.KiroModelAlias) string
}
type claudeModelResolver interface {
	ClaudeModelID(model.ClaudeModelAlias) string
}

// NativeAgentsSupported reports whether the native agent installer handles a
// runtime: it installs agents there.
func NativeAgentsSupported(agent model.AgentID) bool {
	_, installs := NativeAgentManifest[agent]
	return installs
}

// NativeAgentFileNames lists every agent file the installer may write for a
// runtime, so a caller can snapshot all of them before it runs.
func NativeAgentFileNames(agent model.AgentID) []string {
	return append([]string(nil), NativeAgentManifest[agent]...)
}

type InstallOptions struct {
	ClaudeModelAssignments    map[string]model.ClaudeModelAlias
	ClaudePhaseAssignments    map[string]model.ClaudePhaseAssignment
	KiroModelAssignments      map[string]model.KiroModelAlias
	CodeGraphGuidanceMarkdown string
}
type InstallResult struct {
	Changed bool
	Files   []string
	Skipped []string
}

// InstallNativeAgents installs the Judgment Day agents and Kimi's main agent.
// It never removes user-owned agents, and it never removes another product's
// agents: the review lenses belong to gentle-ai, not to Teicos.
func InstallNativeAgents(home string, adapter agents.Adapter, opts InstallOptions) (InstallResult, error) {
	if !NativeAgentsSupported(adapter.Agent()) {
		return InstallResult{}, fmt.Errorf("unsupported native agent runtime: %s", adapter.Agent())
	}
	names := NativeAgentManifest[adapter.Agent()]
	dir := adapter.SubAgentsDir(home)
	if dir == "" {
		return InstallResult{}, fmt.Errorf("empty native agents directory")
	}
	// Render all inputs before touching the target so missing embedded assets do not partially install.
	rendered := make(map[string]string, len(names))
	for _, name := range names {
		content, err := renderNativeAgent(adapter, name, opts)
		if err != nil {
			return InstallResult{}, err
		}
		rendered[name] = content
	}
	if len(names) == 0 {
		// A cleanup-only runtime never creates its agents directory: with no
		// directory there is nothing earlier releases left behind to remove.
		if _, err := os.Lstat(dir); os.IsNotExist(err) {
			return InstallResult{}, nil
		}
	} else if err := os.MkdirAll(dir, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("create native agents directory: %w", err)
	}
	ledgerFile := ledgerPath(dir)
	journal := mutationjournal.New(dir)
	// The ledger is captured first; no agent mutation may precede its snapshot.
	if err := journal.Capture(ledgerFile); err != nil {
		return InstallResult{}, fmt.Errorf("capture ownership ledger: %w", err)
	}
	ledger, _, err := readOwnership(ledgerFile, names)
	if err != nil {
		return InstallResult{}, err
	}
	type candidate struct {
		name, path    string
		data          []byte
		exists, owned bool
	}
	candidates := make([]candidate, 0, len(names))
	result := InstallResult{}
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := journal.Validate(path); err != nil {
			return result, err
		}
		if err := journal.Capture(path); err != nil {
			return result, fmt.Errorf("capture native agent %s: %w", name, err)
		}
		data, exists, err := nativeFile(path)
		if err != nil {
			return result, err
		}
		expected, known := ledger.Files[name]
		owned := exists && known && installedHash(data) == expected
		if exists && !owned {
			result.Skipped = append(result.Skipped, path)
		}
		candidates = append(candidates, candidate{name: name, path: path, data: data, exists: exists, owned: owned})
	}
	rollback := func(err error) (InstallResult, error) {
		return InstallResult{Skipped: result.Skipped}, errors.Join(err, journal.Restore())
	}
	// A legacy installation with every candidate skipped has no owned bytes
	// to record. Do not create a ledger merely because it was absent.
	ledgerChanged := false
	for _, c := range candidates {
		if c.exists && !c.owned {
			continue
		}
		desired := []byte(rendered[c.name])
		if !c.exists || string(c.data) != string(desired) {
			if _, err := journal.WriteWithMode(c.path, desired, 0o644); err != nil {
				return rollback(fmt.Errorf("write native agent %s: %w", c.name, err))
			}
			result.Changed = true
			result.Files = append(result.Files, c.path)
		}
		// Hash the bytes at the installed destination rather than the template.
		actual, err := os.ReadFile(c.path)
		if err != nil {
			return rollback(fmt.Errorf("read installed agent %s: %w", c.name, err))
		}
		if string(actual) != string(desired) {
			return rollback(fmt.Errorf("native agent changed after write: %s", c.path))
		}
		hash := installedHash(actual)
		if ledger.Files[c.name] != hash {
			ledger.Files[c.name] = hash
			ledgerChanged = true
		}
	}
	if ledgerChanged {
		encoded, err := json.MarshalIndent(ledger, "", "  ")
		if err != nil {
			return rollback(fmt.Errorf("encode ownership ledger: %w", err))
		}
		encoded = append(encoded, '\n')
		if _, err := journal.WriteWithMode(ledgerFile, encoded, 0o644); err != nil {
			return rollback(fmt.Errorf("write ownership ledger: %w", err))
		}
		result.Changed = true
		result.Files = append(result.Files, ledgerFile)
	}
	return result, nil
}

// renderNativeAgent renders one embedded native agent for adapter exactly as
// the installer writes it.
func renderNativeAgent(adapter agents.Adapter, name string, opts InstallOptions) (string, error) {
	path := adapter.EmbeddedSubAgentsDir() + "/" + name
	source, err := assets.Read(path)
	if err != nil {
		return "", fmt.Errorf("read native agent %s: %w", path, err)
	}
	content := source
	if strings.HasPrefix(name, "jd-judge-") {
		content = replaceJudgmentSection(content, JudgmentDayReviewerContract())
	}
	phase := strings.TrimSuffix(name, ".md")
	if kmr, ok := adapter.(kiroModelResolver); ok {
		alias := model.KiroModelAuto
		if selected, found := opts.KiroModelAssignments[phase]; found {
			alias = selected
		} else if selected, found := opts.KiroModelAssignments["default"]; found {
			alias = selected
		} else if opts.KiroModelAssignments == nil {
			if selected, found := opts.ClaudeModelAssignments[phase]; found {
				alias = model.KiroModelAlias(selected)
			} else if selected, found := opts.ClaudeModelAssignments["default"]; found {
				alias = model.KiroModelAlias(selected)
			}
		}
		content = strings.ReplaceAll(content, "{{KIRO_MODEL}}", kmr.KiroModelID(alias))
	}
	if cmr, ok := adapter.(claudeModelResolver); ok {
		assignment := resolveClaudeAssignment(opts.ClaudeModelAssignments, opts.ClaudePhaseAssignments, phase)
		content = strings.ReplaceAll(content, "{{CLAUDE_MODEL}}", cmr.ClaudeModelID(assignment.Model))
		effort := ""
		if assignment.Effort != model.ClaudeEffortDefault && model.ClaudeEffortAllowedForModel(assignment.Model, assignment.Effort) {
			effort = "effort: " + string(assignment.Effort)
		}
		if effort == "" {
			content = strings.ReplaceAll(content, "{{CLAUDE_EFFORT_FRONTMATTER}}\r\n", "")
			content = strings.ReplaceAll(content, "{{CLAUDE_EFFORT_FRONTMATTER}}\n", "")
		}
		content = strings.ReplaceAll(content, "{{CLAUDE_EFFORT_FRONTMATTER}}", effort)
	}
	content = engramToolPlaceholder.ReplaceAllString(content, "mcp__engram__$1, mcp__plugin_engram_engram__$1")
	if filepath.Ext(name) == ".md" {
		content = InjectCodeGraphToolGrant(content, adapter.Agent(), opts.CodeGraphGuidanceMarkdown)
		if strings.TrimSpace(opts.CodeGraphGuidanceMarkdown) != "" {
			content = filemerge.InjectMarkdownSection(content, "codegraph-guidance", opts.CodeGraphGuidanceMarkdown)
		}
		content = filemerge.InjectMarkdownSection(content, "agent-language-contract", strings.TrimSpace(assets.MustRead("generic/agent-language-contract.md")))
		content = agentguidance.InjectRemoteAuthorization(content)
	}
	return content, nil
}

// removeRetiredNativeAgent removes one retired agent file when Gentle AI owns
// it and drops its ledger entry either way, since the runtime no longer
// manages that name. It returns the removed path and whether the ledger
// changed. A symlink or other non-regular file is never ours to delete.
func removeRetiredNativeAgent(journal *mutationjournal.Journal, dir, name string, ledger ownershipLedger, managed string) (string, bool, error) {
	path := filepath.Join(dir, name)
	recorded, known := ledger.Files[name]
	delete(ledger.Files, name)

	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", known, nil
	}
	if err != nil {
		return "", known, fmt.Errorf("stat retired native agent %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return "", known, nil
	}
	if err := journal.Validate(path); err != nil {
		return "", known, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", known, fmt.Errorf("read retired native agent %s: %w", path, err)
	}
	owned := (known && installedHash(data) == recorded) || string(data) == managed
	if !owned {
		return "", known, nil
	}
	removed, err := journal.Remove(path)
	if err != nil {
		return "", known, fmt.Errorf("remove retired native agent %s: %w", path, err)
	}
	if !removed {
		return "", known, nil
	}
	return path, known, nil
}

func resolveClaudeAssignment(legacy map[string]model.ClaudeModelAlias, phases map[string]model.ClaudePhaseAssignment, phase string) model.ClaudePhaseAssignment {
	merged := model.ClaudePhaseAssignmentsFromLegacy(model.ClaudeModelPresetBalanced())
	for key, value := range model.ClaudePhaseAssignmentsFromLegacy(legacy) {
		merged[key] = value
	}
	for key, value := range phases {
		if value.Valid() {
			merged[key] = value
		}
	}
	if value, ok := merged[phase]; ok && value.Valid() {
		return value
	}
	if value, ok := merged["default"]; ok && value.Valid() {
		return value
	}
	return model.ClaudePhaseAssignment{Model: model.ClaudeModelSonnet}
}

func replaceJudgmentSection(content, body string) string {
	const heading = "## Review ledger contract"
	start := strings.Index(content, heading)
	if start < 0 {
		return content
	}
	return strings.TrimRight(content[:start], "\n") + "\n\n" + heading + "\n\n" + body + "\n\n"
}
