package theme

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Santiago1809/teicos/internal/agents"
	"github.com/Santiago1809/teicos/internal/agents/claude"
	"github.com/Santiago1809/teicos/internal/agents/hermes"
	"github.com/Santiago1809/teicos/internal/agents/kimi"
	"github.com/Santiago1809/teicos/internal/agents/opencode"
	"github.com/Santiago1809/teicos/internal/model"
)

func claudeAdapter() agents.Adapter   { return claude.NewAdapter() }
func opencodeAdapter() agents.Adapter { return opencode.NewAdapter() }

func TestInjectMergesThemeOverlayIntoAdapterSettings(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll(settings dir) error = %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte("{\n  \"permissions\": {\n    \"allow\": [\"Bash(go test ./...)\"]\n  },\n  \"theme\": \"existing-theme\"\n}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(settings) error = %v", err)
	}

	first, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("Inject() first changed = false")
	}

	second, err := Inject(home, claudeAdapter())
	if err != nil {
		t.Fatalf("Inject() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("Inject() second changed = true")
	}

	if len(first.Files) != 1 || first.Files[0] != settingsPath {
		t.Fatalf("files = %#v, want only %q", first.Files, settingsPath)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings) error = %v", err)
	}
	var root struct {
		Permissions map[string][]string `json:"permissions"`
		Theme       string              `json:"theme"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("Unmarshal(settings) error = %v", err)
	}
	if root.Theme != "teicos" {
		t.Fatalf("theme = %q, want teicos", root.Theme)
	}
	if got := root.Permissions["allow"]; len(got) != 1 || got[0] != "Bash(go test ./...)" {
		t.Fatalf("permissions.allow = %#v, want preserved existing permission", got)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "themes", "teicos.json")); !os.IsNotExist(err) {
		t.Fatalf("Inject() should not write Claude custom theme file; stat error = %v", err)
	}
}

func TestInjectCreatesAdapterSettingsWhenMissing(t *testing.T) {
	home := t.TempDir()

	result, err := Inject(home, opencodeAdapter())
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	settingsPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if !result.Changed {
		t.Fatalf("Inject() changed = false")
	}
	if len(result.Files) != 1 || result.Files[0] != settingsPath {
		t.Fatalf("files = %#v, want only %q", result.Files, settingsPath)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile(settings) error = %v", err)
	}
	var root struct {
		Theme string `json:"theme"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("Unmarshal(settings) error = %v", err)
	}
	if root.Theme != "teicos" {
		t.Fatalf("theme = %q, want teicos", root.Theme)
	}
}

// TestInjectLeavesNonJSONNativeSettingsUntouched covers #1829: valid native
// TOML/YAML settings are not JSON targets, so Inject neither rewrites them nor
// creates a JSON fixture when they are missing.
func TestInjectLeavesNonJSONNativeSettingsUntouched(t *testing.T) {
	tests := []struct {
		name     string
		adapter  agents.Adapter
		settings string
		content  string
	}{
		{name: "kimi legacy", adapter: kimi.NewAdapter(), settings: ".kimi/config.toml", content: "default_model = \"k2\"\n"},
		{name: "kimi current", adapter: kimi.NewAdapter(), settings: ".kimi-code/config.toml", content: "[models.k2]\nprovider = \"moonshot\"\n"},
		{name: "hermes", adapter: hermes.NewAdapter(), settings: ".hermes/config.yaml", content: "providers:\n  - name: hermes\n"},
		{name: "hermes missing", adapter: hermes.NewAdapter(), settings: ".hermes/config.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			settingsPath := filepath.Join(home, tt.settings)
			if tt.content != "" {
				if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(settingsPath, []byte(tt.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			result, err := Inject(home, tt.adapter)
			if err != nil || result.Changed || len(result.Files) != 0 {
				t.Fatalf("Inject() = %#v, %v; want a no-op without reported files", result, err)
			}
			if tt.content == "" {
				if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
					t.Fatalf("Inject() created %s; stat error = %v", settingsPath, err)
				}
				return
			}
			got, err := os.ReadFile(settingsPath)
			if err != nil || string(got) != tt.content {
				t.Fatalf("native settings = %q, %v; want bytes preserved", got, err)
			}
			wantMode := os.FileMode(0o600)
			if runtime.GOOS == "windows" {
				wantMode = 0o666 // Windows reports only the read-only attribute.
			}
			if info, err := os.Stat(settingsPath); err != nil || info.Mode().Perm() != wantMode {
				t.Fatalf("native settings mode = %v, %v; want %v", info, err, wantMode)
			}
		})
	}
}

func TestInjectAtPathAcceptsCustomSelectedFilename(t *testing.T) {
	selected := filepath.Join(t.TempDir(), "custom-opencode-settings")
	if err := os.WriteFile(selected, []byte(`{"custom":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := InjectAtPath(selected)
	if err != nil || !result.Changed || len(result.Files) != 1 || result.Files[0] != selected {
		t.Fatalf("InjectAtPath() = %#v, %v; want the selected file written", result, err)
	}
	got, err := os.ReadFile(selected)
	if err != nil || !bytes.Contains(got, []byte(`"theme": "teicos"`)) || !bytes.Contains(got, []byte(`"custom": true`)) {
		t.Fatalf("selected settings = %s, %v", got, err)
	}
}

func TestInjectAtPathPreservesSelectedJSONCAndLeavesDecoy(t *testing.T) {
	root := t.TempDir()
	selected := filepath.Join(root, "opencode.jsonc")
	decoy := filepath.Join(root, "opencode.json")
	before := []byte("// selected\n{\"theme\":\"old\",\"custom\":true}\n")
	if err := os.WriteFile(selected, before, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(decoy, []byte(`{"theme":"mine"}`), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := InjectAtPath(selected)
	if err != nil || !first.Changed || len(first.Files) != 1 || first.Files[0] != selected {
		t.Fatalf("first injection = %#v, %v", first, err)
	}
	second, err := InjectAtPath(selected)
	if err != nil || second.Changed {
		t.Fatalf("second injection = %#v, %v", second, err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(selected)
		if err != nil || info.Mode().Perm() != 0o600 {
			t.Fatalf("selected JSONC mode = %v, %v; want 0600", info, err)
		}
	}
	got, err := os.ReadFile(selected)
	if err != nil || !bytes.Contains(got, []byte(`"theme":"teicos"`)) || !bytes.Contains(got, []byte(`"custom":true`)) || !bytes.Contains(got, []byte("// selected")) {
		t.Fatalf("selected JSONC = %s, %v", got, err)
	}
	got, err = os.ReadFile(decoy)
	if err != nil || string(got) != `{"theme":"mine"}` {
		t.Fatalf("decoy JSON = %s, %v", got, err)
	}
}

func TestInjectAtPathRejectsZeroPermissionSettingsWithoutChangingThem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permission bits are not supported on Windows")
	}

	path := filepath.Join(t.TempDir(), "opencode.jsonc")
	before := []byte("// private settings\n{\"theme\":\"old\"}\n")
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("cannot set file mode 0000: %v", err)
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode().Perm() != 0 {
		t.Skipf("cannot verify file mode 0000: %v, %v", info, err)
	}

	result, err := InjectAtPath(path)
	if err == nil || !strings.Contains(err.Error(), "0000") || !strings.Contains(err.Error(), "permissions") {
		t.Errorf("InjectAtPath() = %#v, %v; want actionable mode 0000 error", result, err)
	}
	if result.Changed || len(result.Files) != 0 {
		t.Errorf("InjectAtPath() = %#v; want no changed files", result)
	}
	info, err = os.Lstat(path)
	if err != nil || info.Mode().Perm() != 0 {
		t.Fatalf("settings mode after injection = %v, %v; want 0000", info, err)
	}
	// Restore access only to inspect bytes, then leave the file locked down.
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	got, readErr := os.ReadFile(path)
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	if readErr != nil || !bytes.Equal(got, before) {
		t.Fatalf("settings bytes after injection = %q, %v; want %q", got, readErr, before)
	}
}

func TestInjectAtPathRefusesSymlinkWithoutChangingLinkOrTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "user.jsonc")
	selected := filepath.Join(dir, "opencode.jsonc")
	before := []byte("// private settings\n{\"theme\":\"old\"}\n")
	if err := os.WriteFile(target, before, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, selected); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if result, err := InjectAtPath(selected); err == nil || !strings.Contains(err.Error(), "select a regular settings file") || result.Changed {
		t.Fatalf("InjectAtPath() = %#v, %v; want symlink refusal", result, err)
	}
	if link, err := os.Readlink(selected); err != nil || link != target {
		t.Fatalf("selected symlink changed: %q, %v", link, err)
	}
	got, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(got, before) {
		t.Fatalf("target bytes = %q, %v; want %q", got, err, before)
	}
	if info, err := os.Stat(target); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0o600) {
		t.Fatalf("target mode changed: %v, %v", info, err)
	}
}

// The selected-settings refusal is OpenCode-only; other agents keep the base
// writer behavior for a dotfiles-managed (symlinked) settings file.
func TestInjectNonOpenCodeSymlinkedSettingsKeepBaseWriterBehavior(t *testing.T) {
	home := t.TempDir()
	adapter := claudeAdapter()
	settings := adapter.SettingsPath(home)
	target := filepath.Join(home, "dotfiles", "settings.json")
	before := []byte("{\"theme\":\"old\"}\n")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, before, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, settings); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	result, err := Inject(home, adapter)
	if err == nil || !strings.Contains(err.Error(), "refusing to read symlink") || strings.Contains(err.Error(), "select a regular settings file") {
		t.Fatalf("Inject() = %#v, %v; want base writer symlink error, not the OpenCode refusal", result, err)
	}
	if link, err := os.Readlink(settings); err != nil || link != target {
		t.Fatalf("settings symlink changed: %q, %v", link, err)
	}
	if got, err := os.ReadFile(target); err != nil || !bytes.Equal(got, before) {
		t.Fatalf("target bytes = %q, %v; want %q", got, err, before)
	}
}

func TestInjectVisualThemesIsIdempotentForClaude(t *testing.T) {
	home := t.TempDir()

	first, err := InjectVisualThemes(home, claudeAdapter())
	if err != nil {
		t.Fatalf("InjectVisualThemes() first error = %v", err)
	}
	if !first.Changed {
		t.Fatalf("InjectVisualThemes() first changed = false")
	}

	second, err := InjectVisualThemes(home, claudeAdapter())
	if err != nil {
		t.Fatalf("InjectVisualThemes() second error = %v", err)
	}
	if second.Changed {
		t.Fatalf("InjectVisualThemes() second changed = true")
	}

	path := filepath.Join(home, ".claude", "themes", "teicos.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected Claude theme file %q: %v", path, err)
	}
}

func TestInjectVisualThemesSkipsUnsupportedAdapter(t *testing.T) {
	home := t.TempDir()
	adapter, _ := agents.NewAdapter(model.AgentGeminiCLI)

	result, err := InjectVisualThemes(home, adapter)
	if err != nil {
		t.Fatalf("InjectVisualThemes() error = %v", err)
	}
	if result.Changed || len(result.Files) != 0 {
		t.Fatalf("InjectVisualThemes() = %#v, want no-op for unsupported adapter", result)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "themes", "teicos.json")); !os.IsNotExist(err) {
		t.Fatalf("InjectVisualThemes() should not write Claude files for Gemini; stat error = %v", err)
	}
}

func TestInjectVisualThemesPreservesTeicosClaudeTheme(t *testing.T) {
	home := t.TempDir()

	result, err := InjectVisualThemes(home, claudeAdapter())
	if err != nil {
		t.Fatalf("InjectVisualThemes() error = %v", err)
	}

	themePath := filepath.Join(home, ".claude", "themes", "teicos.json")
	if len(result.Files) != 2 || result.Files[0] != themePath {
		t.Fatalf("files = %#v, want Gentleman first at %q", result.Files, themePath)
	}

	data, err := os.ReadFile(themePath)
	if err != nil {
		t.Fatalf("ReadFile(theme) error = %v", err)
	}

	var root struct {
		Name      string            `json:"name"`
		Base      string            `json:"base"`
		Overrides map[string]string `json:"overrides"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("Unmarshal(theme) error = %v", err)
	}

	if root.Name != "teicos" || root.Base != "dark" {
		t.Fatalf("theme identity = %q/%q, want teicos/dark", root.Name, root.Base)
	}
	// The Teicos palette: Aioniq base with a deep red accent.
	expected := map[string]string{
		"text":                      "#FFFFFF",
		"subtle":                    "#2E2E2E",
		"userMessageBackground":     "#1A1A1A",
		"promptBorder":              "#8B0000",
		"selectionBg":               "#3A1414",
		"diffAdded":                 "#131F13",
		"diffRemoved":               "#1F1315",
		"diffAddedWord":             "#2D5A45",
		"diffRemovedWord":           "#7A2948",
		"chromeYellow":              "#D9A441",
		"briefLabelYou":             "#D9A441",
		"rainbow_yellow":            "#D9A441",
		"yellow_FOR_SUBAGENTS_ONLY": "#D9A441",
	}
	for key, want := range expected {
		if root.Overrides[key] != want {
			t.Fatalf("override %s = %q, want %q", key, root.Overrides[key], want)
		}
	}
	for _, forbidden := range []string{"markdown", "syntax", "keyword", "string"} {
		if _, ok := root.Overrides[forbidden]; ok {
			t.Fatalf("theme contains forbidden non-Claude theme key %q", forbidden)
		}
	}
}

const (
	teicosClaudeFixture       = `{"name":"teicos","base":"dark","overrides":{"claude":"#E5484D","claudeShimmer":"#FF7A7D","text":"#FFFFFF","inactive":"#6A6A6A","subtle":"#2E2E2E","suggestion":"#E5484D","permission":"#E5484D","promptBorder":"#8B0000","planMode":"#5C8AB8","autoAccept":"#4E9A51","bashBorder":"#D9A441","remember":"#D9A441","success":"#4E9A51","merged":"#4E9A51","error":"#E5484D","warning":"#D9A441","diffAdded":"#131F13","diffRemoved":"#1F1315","diffAddedWord":"#2D5A45","diffRemovedWord":"#7A2948","userMessageBackground":"#1A1A1A","userMessageBackgroundHover":"#242424","selectionBg":"#3A1414","memoryBackgroundColor":"#141414","bashMessageBackgroundColor":"#121212","chromeYellow":"#D9A441","briefLabelYou":"#D9A441","rainbow_yellow":"#D9A441","yellow_FOR_SUBAGENTS_ONLY":"#D9A441"}}`
	teicosCuteClaudeFixture   = `{"name":"Teicos Cute","base":"dark","overrides":{"claude":"#C0392B","claudeShimmer":"#E5484D","text":"#FFFFFF","inactive":"#6A6A6A","subtle":"#2E2E2E","suggestion":"#E5484D","permission":"#E5484D","promptBorder":"#C0392B","planMode":"#5C8AB8","autoAccept":"#4E9A51","bashBorder":"#D9A441","remember":"#D9A441","success":"#4E9A51","merged":"#4E9A51","error":"#E5484D","warning":"#D9A441","diffAdded":"#131F13","diffRemoved":"#1F1315","diffAddedWord":"#2D5A45","diffRemovedWord":"#7A2948","userMessageBackground":"#1F1A1A","userMessageBackgroundHover":"#292424","selectionBg":"#3A1414","memoryBackgroundColor":"#141414","bashMessageBackgroundColor":"#121212","chromeYellow":"#D9A441","briefLabelYou":"#D9A441","rainbow_yellow":"#D9A441","yellow_FOR_SUBAGENTS_ONLY":"#D9A441"}}`
	teicosOpenCodeFixture     = `{"$schema":"https://opencode.ai/theme.json","theme":{"background":"#0D0D0D","backgroundPanel":"#1A1A1A","backgroundElement":"#1A1A1A","text":"#FFFFFF","textMuted":"#8A8A8A","primary":"#8B0000","secondary":"#B3B3B3","accent":"#8B0000","error":"#E5484D","warning":"#D9A441","success":"#4E9A51","info":"#5C8AB8","border":"#2E2E2E","borderActive":"#8B0000","borderSubtle":"#1F1F1F","diffAdded":"#4E9A51","diffRemoved":"#E5484D","diffContext":"#8A8A8A","diffHunkHeader":"#B3B3B3","diffHighlightAdded":"#6BBF6E","diffHighlightRemoved":"#FF7A7D","diffAddedBg":"#131F13","diffRemovedBg":"#1F1315","diffContextBg":"#141414","diffLineNumber":"#6A6A6A","diffAddedLineNumberBg":"#131F13","diffRemovedLineNumberBg":"#1F1315","markdownText":"#FFFFFF","markdownHeading":"#E8E8E8","markdownLink":"#E5484D","markdownLinkText":"#E5484D","markdownCode":"#4E9A51","markdownBlockQuote":"#8A8A8A","markdownEmph":"#C7C7C7","markdownStrong":"#FFFFFF","markdownHorizontalRule":"#2E2E2E","markdownListItem":"#B3B3B3","markdownListEnumeration":"#8A8A8A","markdownImage":"#B3B3B3","markdownImageText":"#8A8A8A","markdownCodeBlock":"#E8E8E8","syntaxComment":"#6A6A6A","syntaxKeyword":"#E5484D","syntaxFunction":"#C7C7C7","syntaxVariable":"#FFFFFF","syntaxString":"#4E9A51","syntaxNumber":"#D9A441","syntaxType":"#5C8AB8","syntaxOperator":"#8A8A8A","syntaxPunctuation":"#8A8A8A"}}`
	teicosCuteOpenCodeFixture = `{"$schema":"https://opencode.ai/theme.json","theme":{"background":"#0D0D0D","backgroundPanel":"#1F1A1A","backgroundElement":"#1F1A1A","text":"#FFFFFF","textMuted":"#8A8A8A","primary":"#C0392B","secondary":"#B3B3B3","accent":"#C0392B","error":"#E5484D","warning":"#D9A441","success":"#4E9A51","info":"#5C8AB8","border":"#2E2E2E","borderActive":"#C0392B","borderSubtle":"#241F1F","diffAdded":"#4E9A51","diffRemoved":"#E5484D","diffContext":"#8A8A8A","diffHunkHeader":"#B3B3B3","diffHighlightAdded":"#6BBF6E","diffHighlightRemoved":"#FF7A7D","diffAddedBg":"#131F13","diffRemovedBg":"#1F1315","diffContextBg":"#141414","diffLineNumber":"#6A6A6A","diffAddedLineNumberBg":"#131F13","diffRemovedLineNumberBg":"#1F1315","markdownText":"#FFFFFF","markdownHeading":"#E8E8E8","markdownLink":"#E5484D","markdownLinkText":"#E5484D","markdownCode":"#4E9A51","markdownBlockQuote":"#8A8A8A","markdownEmph":"#C7C7C7","markdownStrong":"#FFFFFF","markdownHorizontalRule":"#2E2E2E","markdownListItem":"#B3B3B3","markdownListEnumeration":"#8A8A8A","markdownImage":"#B3B3B3","markdownImageText":"#8A8A8A","markdownCodeBlock":"#E8E8E8","syntaxComment":"#6A6A6A","syntaxKeyword":"#E5484D","syntaxFunction":"#C7C7C7","syntaxVariable":"#FFFFFF","syntaxString":"#4E9A51","syntaxNumber":"#D9A441","syntaxType":"#5C8AB8","syntaxOperator":"#8A8A8A","syntaxPunctuation":"#8A8A8A"}}`
)

// TestTeicosOpenCodeThemeUsesAioniqBaseWithDeepRedAccent pins the Teicos
// palette: the Aioniq base tokens plus a deep red accent. The accent is a
// surface colour, so it is used for fills and borders; text-bearing accent
// tokens use a lifted red that stays legible on the dark background.
func TestTeicosOpenCodeThemeUsesAioniqBaseWithDeepRedAccent(t *testing.T) {
	base := map[string]string{
		"background":      "#0D0D0D",
		"backgroundPanel": "#1A1A1A",
		"border":          "#2E2E2E",
		"primary":         "#8B0000",
		"accent":          "#8B0000",
		"text":            "#FFFFFF",
	}
	for token, want := range base {
		if got := teicosOpenCodeTheme.Theme[token]; got != want {
			t.Errorf("theme token %q = %q, want %q", token, got, want)
		}
	}

	// The deep red is 1.94:1 on the background: a fill, never text.
	for _, token := range []string{"markdownLink", "markdownLinkText", "syntaxKeyword"} {
		if got := teicosOpenCodeTheme.Theme[token]; got == "#8B0000" {
			t.Errorf("token %q uses the deep red accent as text colour %q", token, got)
		}
	}
}

func TestInjectVisualThemesWritesExpectedAssets(t *testing.T) {
	home, xdg := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)
	for _, adapter := range []agents.Adapter{claudeAdapter(), opencodeAdapter()} {
		settings := adapter.SettingsPath(home)
		if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(settings, []byte(`{"theme":"user-selected"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		first, err := InjectVisualThemes(home, adapter)
		if err != nil || !first.Changed {
			t.Fatalf("first injection = %#v, %v", first, err)
		}
		if second, err := InjectVisualThemes(home, adapter); err != nil || second.Changed {
			t.Fatalf("second injection = %#v, %v", second, err)
		}
		if got, err := os.ReadFile(settings); err != nil || string(got) != `{"theme":"user-selected"}` {
			t.Fatalf("settings changed = %q, %v", got, err)
		}
	}
	for _, tt := range []struct {
		path, fixture string
		value         any
	}{
		{filepath.Join(home, ".claude", "themes", "teicos.json"), teicosClaudeFixture, &claudeTheme{}},
		{filepath.Join(home, ".claude", "themes", "teicos-cute.json"), teicosCuteClaudeFixture, &claudeTheme{}},
		{filepath.Join(xdg, "opencode", "themes", "teicos.json"), teicosOpenCodeFixture, &openCodeTheme{}},
		{filepath.Join(xdg, "opencode", "themes", "teicos-cute.json"), teicosCuteOpenCodeFixture, &openCodeTheme{}},
	} {
		got, err := os.ReadFile(tt.path)
		if err != nil || !bytes.Equal(got, indentedJSON(t, tt.fixture, tt.value)) {
			t.Fatalf("%s bytes mismatch: %v", tt.path, err)
		}
	}
	cute := teicosCuteClaudeTheme
	for _, key := range []string{"background", "backgroundFullscreen", "backgroundUser", "backgroundHover", "backgroundSelection", "backgroundMemory", "backgroundBash", "shimmer"} {
		if _, exists := cute.Overrides[key]; exists {
			t.Fatalf("Claude Cute contains invalid override %q", key)
		}
	}
}

func indentedJSON(t *testing.T, fixture string, value any) []byte {
	t.Helper()
	if err := json.Unmarshal([]byte(fixture), value); err != nil {
		t.Fatal(err)
	}
	result, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(result, '\n')
}
