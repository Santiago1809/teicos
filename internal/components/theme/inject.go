package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Santiago1809/teicos/internal/agents"
	"github.com/Santiago1809/teicos/internal/components/filemerge"
	"github.com/Santiago1809/teicos/internal/model"
)

type InjectionResult struct {
	Changed bool
	Files   []string
}

var themeOverlayJSON = []byte("{\n  \"theme\": \"teicos\"\n}\n")

type claudeTheme struct {
	Name      string            `json:"name"`
	Base      string            `json:"base"`
	Overrides map[string]string `json:"overrides"`
}

type openCodeTheme struct {
	Schema string            `json:"$schema"`
	Theme  map[string]string `json:"theme"`
}

const openCodeThemeSchema = "https://opencode.ai/theme.json"

func palette(pairs ...string) map[string]string {
	colors := make(map[string]string, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		colors[pairs[i]] = pairs[i+1]
	}
	return colors
}

var teicosClaudeTheme = claudeTheme{
	Name: "teicos", Base: "dark",
	Overrides: palette(
		"claude", "#E5484D", "claudeShimmer", "#FF7A7D", "text", "#FFFFFF", "inactive", "#6A6A6A", "subtle", "#2E2E2E", "suggestion", "#E5484D",
		"permission", "#E5484D", "promptBorder", "#8B0000", "planMode", "#5C8AB8", "autoAccept", "#4E9A51", "bashBorder", "#D9A441",
		"remember", "#D9A441", "success", "#4E9A51", "merged", "#4E9A51", "error", "#E5484D", "warning", "#D9A441",
		"diffAdded", "#131F13", "diffRemoved", "#1F1315", "diffAddedWord", "#2D5A45", "diffRemovedWord", "#7A2948",
		"userMessageBackground", "#1A1A1A", "userMessageBackgroundHover", "#242424", "selectionBg", "#3A1414", "memoryBackgroundColor", "#141414", "bashMessageBackgroundColor", "#121212",
		"chromeYellow", "#D9A441", "briefLabelYou", "#D9A441", "rainbow_yellow", "#D9A441", "yellow_FOR_SUBAGENTS_ONLY", "#D9A441",
	),
}

var teicosCuteClaudeTheme = claudeTheme{
	Name: "Teicos Cute", Base: "dark",
	Overrides: palette(
		"claude", "#C0392B", "claudeShimmer", "#E5484D", "text", "#FFFFFF", "inactive", "#6A6A6A", "subtle", "#2E2E2E", "suggestion", "#E5484D",
		"permission", "#E5484D", "promptBorder", "#C0392B", "planMode", "#5C8AB8", "autoAccept", "#4E9A51", "bashBorder", "#D9A441",
		"remember", "#D9A441", "success", "#4E9A51", "merged", "#4E9A51", "error", "#E5484D", "warning", "#D9A441",
		"diffAdded", "#131F13", "diffRemoved", "#1F1315", "diffAddedWord", "#2D5A45", "diffRemovedWord", "#7A2948",
		"userMessageBackground", "#1F1A1A", "userMessageBackgroundHover", "#292424", "selectionBg", "#3A1414", "memoryBackgroundColor", "#141414", "bashMessageBackgroundColor", "#121212",
		"chromeYellow", "#D9A441", "briefLabelYou", "#D9A441", "rainbow_yellow", "#D9A441", "yellow_FOR_SUBAGENTS_ONLY", "#D9A441",
	),
}

var teicosOpenCodeTheme = openCodeTheme{
	Schema: openCodeThemeSchema,
	Theme: palette(
		"background", "#0D0D0D", "backgroundPanel", "#1A1A1A", "backgroundElement", "#1A1A1A", "text", "#FFFFFF", "textMuted", "#8A8A8A",
		"primary", "#8B0000", "secondary", "#B3B3B3", "accent", "#8B0000", "error", "#E5484D", "warning", "#D9A441", "success", "#4E9A51", "info", "#5C8AB8",
		"border", "#2E2E2E", "borderActive", "#8B0000", "borderSubtle", "#1F1F1F", "diffAdded", "#4E9A51", "diffRemoved", "#E5484D", "diffContext", "#8A8A8A",
		"diffHunkHeader", "#B3B3B3", "diffHighlightAdded", "#6BBF6E", "diffHighlightRemoved", "#FF7A7D", "diffAddedBg", "#131F13", "diffRemovedBg", "#1F1315", "diffContextBg", "#141414",
		"diffLineNumber", "#6A6A6A", "diffAddedLineNumberBg", "#131F13", "diffRemovedLineNumberBg", "#1F1315", "markdownText", "#FFFFFF", "markdownHeading", "#E8E8E8",
		"markdownLink", "#E5484D", "markdownLinkText", "#E5484D", "markdownCode", "#4E9A51", "markdownBlockQuote", "#8A8A8A", "markdownEmph", "#C7C7C7", "markdownStrong", "#FFFFFF",
		"markdownHorizontalRule", "#2E2E2E", "markdownListItem", "#B3B3B3", "markdownListEnumeration", "#8A8A8A", "markdownImage", "#B3B3B3", "markdownImageText", "#8A8A8A", "markdownCodeBlock", "#E8E8E8",
		"syntaxComment", "#6A6A6A", "syntaxKeyword", "#E5484D", "syntaxFunction", "#C7C7C7", "syntaxVariable", "#FFFFFF", "syntaxString", "#4E9A51", "syntaxNumber", "#D9A441", "syntaxType", "#5C8AB8", "syntaxOperator", "#8A8A8A", "syntaxPunctuation", "#8A8A8A",
	),
}

var teicosCuteOpenCodeTheme = openCodeTheme{
	Schema: openCodeThemeSchema,
	Theme: palette(
		"background", "#0D0D0D", "backgroundPanel", "#1F1A1A", "backgroundElement", "#1F1A1A", "text", "#FFFFFF", "textMuted", "#8A8A8A",
		"primary", "#C0392B", "secondary", "#B3B3B3", "accent", "#C0392B", "error", "#E5484D", "warning", "#D9A441", "success", "#4E9A51", "info", "#5C8AB8",
		"border", "#2E2E2E", "borderActive", "#C0392B", "borderSubtle", "#241F1F", "diffAdded", "#4E9A51", "diffRemoved", "#E5484D", "diffContext", "#8A8A8A",
		"diffHunkHeader", "#B3B3B3", "diffHighlightAdded", "#6BBF6E", "diffHighlightRemoved", "#FF7A7D", "diffAddedBg", "#131F13", "diffRemovedBg", "#1F1315", "diffContextBg", "#141414",
		"diffLineNumber", "#6A6A6A", "diffAddedLineNumberBg", "#131F13", "diffRemovedLineNumberBg", "#1F1315", "markdownText", "#FFFFFF", "markdownHeading", "#E8E8E8",
		"markdownLink", "#E5484D", "markdownLinkText", "#E5484D", "markdownCode", "#4E9A51", "markdownBlockQuote", "#8A8A8A", "markdownEmph", "#C7C7C7", "markdownStrong", "#FFFFFF",
		"markdownHorizontalRule", "#2E2E2E", "markdownListItem", "#B3B3B3", "markdownListEnumeration", "#8A8A8A", "markdownImage", "#B3B3B3", "markdownImageText", "#8A8A8A", "markdownCodeBlock", "#E8E8E8",
		"syntaxComment", "#6A6A6A", "syntaxKeyword", "#E5484D", "syntaxFunction", "#C7C7C7", "syntaxVariable", "#FFFFFF", "syntaxString", "#4E9A51", "syntaxNumber", "#D9A441", "syntaxType", "#5C8AB8", "syntaxOperator", "#8A8A8A", "syntaxPunctuation", "#8A8A8A",
	),
}

func Inject(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	return injectSettings(agents.JSONSettingsPath(homeDir, adapter))
}

// InjectAtPath writes the theme to a caller-selected settings file. OpenCode
// callers pass the effective JSON/JSONC path and get a refusal for symlinked,
// non-regular or locked files; other adapters use Inject and keep the shared
// writer behavior.
func InjectAtPath(settingsPath string) (InjectionResult, error) {
	if settingsPath == "" {
		return InjectionResult{}, nil
	}
	if err := filemerge.RefuseLockedSettingsFile(settingsPath); err != nil {
		return InjectionResult{}, err
	}
	return injectSettings(settingsPath)
}

func injectSettings(settingsPath string) (InjectionResult, error) {
	if settingsPath == "" {
		return InjectionResult{}, nil
	}

	writeResult, err := mergeJSONFile(settingsPath, themeOverlayJSON)
	if err != nil {
		return InjectionResult{}, err
	}

	return InjectionResult{Changed: writeResult.Changed, Files: []string{settingsPath}}, nil
}

// InjectVisualThemes writes the managed visual theme assets without selecting one
// in an agent's active settings.
func InjectVisualThemes(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	paths := VisualThemePaths(homeDir, adapter)
	if len(paths) == 0 {
		return InjectionResult{}, nil
	}

	var values []any
	switch adapter.Agent() {
	case model.AgentClaudeCode:
		values = []any{teicosClaudeTheme, teicosCuteClaudeTheme}
	case model.AgentOpenCode:
		values = []any{teicosOpenCodeTheme, teicosCuteOpenCodeTheme}
	}

	result := InjectionResult{Files: make([]string, 0, len(paths))}
	for i, path := range paths {
		content, err := json.MarshalIndent(values[i], "", "  ")
		if err != nil {
			return InjectionResult{}, fmt.Errorf("marshal visual theme %q: %w", filepath.Base(path), err)
		}
		content = append(content, '\n')
		writeResult, err := filemerge.WriteFileAtomic(path, content, 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		result.Changed = result.Changed || writeResult.Changed
		result.Files = append(result.Files, path)
	}
	return result, nil
}

// VisualThemePaths returns the installer-owned visual theme assets for an adapter.
func VisualThemePaths(homeDir string, adapter agents.Adapter) []string {
	var root string
	switch adapter.Agent() {
	case model.AgentClaudeCode:
		root = filepath.Join(adapter.GlobalConfigDir(homeDir), "themes")
	case model.AgentOpenCode:
		root = filepath.Join(filepath.Dir(adapter.SettingsPath(homeDir)), "themes")
	default:
		return nil
	}
	return []string{filepath.Join(root, "teicos.json"), filepath.Join(root, "teicos-cute.json")}
}

func mergeJSONFile(path string, overlay []byte) (filemerge.WriteResult, error) {
	mode := filemerge.ExistingFileMode(path, 0o644)

	baseJSON, err := osReadFile(path)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	merged, err := filemerge.MergeJSONObjectsForPath(path, baseJSON, overlay)
	if err != nil {
		return filemerge.WriteResult{}, err
	}

	return filemerge.WriteFileAtomic(path, merged, mode)
}

var osReadFile = func(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read json file %q: %w", path, err)
	}

	return content, nil
}
