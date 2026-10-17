package state

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// clearHomeEnv unsets both override variables and captures the warnings Dir
// writes, so every test starts from a known process-level state.
func clearHomeEnv(t *testing.T) *bytes.Buffer {
	t.Helper()
	t.Setenv(HomeEnv, "")
	t.Setenv(HomeEnvLegacy, "")
	var warnings bytes.Buffer
	previous := warnOut
	warnOut = &warnings
	legacyEnvWarned.Store(false)
	legacyDirWarned.Store(false)
	t.Cleanup(func() { warnOut = previous })
	return &warnings
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestDirDefaultsToDotTeicosUnderHome(t *testing.T) {
	warnings := clearHomeEnv(t)
	home := t.TempDir()

	if got, want := Dir(home), filepath.Join(home, ".teicos"); got != want {
		t.Fatalf("Dir() = %q, want %q", got, want)
	}
	if got, want := Path(home), filepath.Join(home, ".teicos", "state.json"); got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
	if got, want := ManifestPath(home), filepath.Join(home, ".teicos", "managed-assets.manifest.json"); got != want {
		t.Fatalf("ManifestPath() = %q, want %q", got, want)
	}
	if err := ValidateHomeEnv(); err != nil {
		t.Fatalf("ValidateHomeEnv() with no override = %v, want nil", err)
	}
	if warnings.Len() != 0 {
		t.Fatalf("a fresh home must not warn, got %q", warnings.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".teicos")); !os.IsNotExist(err) {
		t.Fatalf("Dir() must not create the directory (stat err = %v)", err)
	}
}

func TestDirPrecedence(t *testing.T) {
	newOverride := filepath.Join(t.TempDir(), "teicos-home")
	legacyOverride := filepath.Join(t.TempDir(), "gentle-home")

	t.Run("TEICOS_HOME wins over everything and does not warn", func(t *testing.T) {
		warnings := clearHomeEnv(t)
		home := t.TempDir()
		mkdir(t, filepath.Join(home, ".teicos"))
		mkdir(t, filepath.Join(home, ".gentle-ai"))
		t.Setenv(HomeEnv, newOverride)
		t.Setenv(HomeEnvLegacy, legacyOverride)

		if got := Dir(home); got != newOverride {
			t.Fatalf("Dir() = %q, want %q", got, newOverride)
		}
		if warnings.Len() != 0 {
			t.Fatalf("unexpected warning %q", warnings.String())
		}
	})

	t.Run("GENTLE_AI_HOME is honoured with a deprecation warning", func(t *testing.T) {
		warnings := clearHomeEnv(t)
		home := t.TempDir()
		mkdir(t, filepath.Join(home, ".teicos"))
		t.Setenv(HomeEnvLegacy, legacyOverride)

		if got := Dir(home); got != legacyOverride {
			t.Fatalf("Dir() = %q, want %q", got, legacyOverride)
		}
		for _, want := range []string{HomeEnvLegacy, HomeEnv, "deprecated"} {
			if !strings.Contains(warnings.String(), want) {
				t.Fatalf("warning %q does not mention %q", warnings.String(), want)
			}
		}
	})

	t.Run("an existing ~/.teicos wins over ~/.gentle-ai without warning", func(t *testing.T) {
		warnings := clearHomeEnv(t)
		home := t.TempDir()
		mkdir(t, filepath.Join(home, ".teicos"))
		mkdir(t, filepath.Join(home, ".gentle-ai"))

		if got, want := Dir(home), filepath.Join(home, ".teicos"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
		if warnings.Len() != 0 {
			t.Fatalf("unexpected warning %q", warnings.String())
		}
	})

	t.Run("only ~/.gentle-ai existing is used and suggests a migration", func(t *testing.T) {
		warnings := clearHomeEnv(t)
		home := t.TempDir()
		mkdir(t, filepath.Join(home, ".gentle-ai"))

		if got, want := Dir(home), filepath.Join(home, ".gentle-ai"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
		if !strings.Contains(warnings.String(), "cp -r ~/.gentle-ai ~/.teicos") {
			t.Fatalf("warning %q does not suggest the migration command", warnings.String())
		}
		if _, err := os.Stat(filepath.Join(home, ".teicos")); !os.IsNotExist(err) {
			t.Fatalf("Dir() must never migrate on its own (stat err = %v)", err)
		}
	})

	t.Run("a regular file named .teicos does not count as the state directory", func(t *testing.T) {
		clearHomeEnv(t)
		home := t.TempDir()
		if err := os.WriteFile(filepath.Join(home, ".teicos"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		mkdir(t, filepath.Join(home, ".gentle-ai"))

		if got, want := Dir(home), filepath.Join(home, ".gentle-ai"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
	})

	t.Run("relative overrides are ignored", func(t *testing.T) {
		clearHomeEnv(t)
		home := t.TempDir()
		t.Setenv(HomeEnv, "relative/teicos")
		t.Setenv(HomeEnvLegacy, "relative/gentle")

		if got, want := Dir(home), filepath.Join(home, ".teicos"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
	})
}

func TestDirWarnsOncePerProcess(t *testing.T) {
	warnings := clearHomeEnv(t)
	home := t.TempDir()
	mkdir(t, filepath.Join(home, ".gentle-ai"))

	Dir(home)
	Dir(home)
	if got := strings.Count(warnings.String(), "cp -r"); got != 1 {
		t.Fatalf("migration warning printed %d times, want 1", got)
	}

	warnings.Reset()
	t.Setenv(HomeEnvLegacy, filepath.Join(t.TempDir(), "gentle-home"))
	Dir(home)
	Dir(home)
	if got := strings.Count(warnings.String(), "deprecated"); got != 1 {
		t.Fatalf("deprecation warning printed %d times, want 1", got)
	}
}

func TestHomeEnvRoutesStateAndManifestFiles(t *testing.T) {
	clearHomeEnv(t)
	home := t.TempDir()
	override := filepath.Join(t.TempDir(), "teicos")
	t.Setenv(HomeEnv, override)

	if got := Dir(home); got != override {
		t.Fatalf("Dir() = %q, want %q", got, override)
	}
	if err := Write(home, InstallState{InstalledAgents: []string{"pi"}}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := WriteManifestAtomic(home, Manifest{Schema: ManifestSchema}); err != nil {
		t.Fatalf("WriteManifestAtomic() error = %v", err)
	}
	if err := AppendJournal(home, OpIntent, "run", "resource"); err != nil {
		t.Fatalf("AppendJournal() error = %v", err)
	}

	for _, name := range []string{"state.json", "managed-assets.manifest.json", "managed-assets.journal"} {
		if _, err := os.Stat(filepath.Join(override, name)); err != nil {
			t.Fatalf("%s not written under %s: %v", name, HomeEnv, err)
		}
	}
	for _, dir := range []string{".teicos", ".gentle-ai"} {
		if _, err := os.Stat(filepath.Join(home, dir)); !os.IsNotExist(err) {
			t.Fatalf("~/%s must not be created when %s is set (stat err = %v)", dir, HomeEnv, err)
		}
	}
}

func TestValidateHomeEnvRejectsRelativePath(t *testing.T) {
	for _, tt := range []struct {
		name, key, value, other string
		wantErr                 string
	}{
		{"new variable", HomeEnv, "relative/teicos", "", HomeEnv},
		{"legacy variable alone", HomeEnvLegacy, "relative/gentle-ai", "", HomeEnvLegacy},
		{"new variable wins over a bad legacy one", HomeEnv, "/abs/teicos", "relative/gentle-ai", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clearHomeEnv(t)
			t.Setenv(tt.key, tt.value)
			if tt.other != "" {
				t.Setenv(HomeEnvLegacy, tt.other)
			}

			err := ValidateHomeEnv()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateHomeEnv() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateHomeEnv() = nil, want error for a relative path")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidateHomeEnv() error %q does not name %s", err, tt.wantErr)
			}
		})
	}
}

func TestHomeOverrideSet(t *testing.T) {
	clearHomeEnv(t)
	if HomeOverrideSet() {
		t.Fatal("HomeOverrideSet() = true with no override")
	}
	t.Setenv(HomeEnvLegacy, "/abs/gentle")
	if !HomeOverrideSet() {
		t.Fatal("HomeOverrideSet() = false with the legacy variable set")
	}
	t.Setenv(HomeEnvLegacy, "")
	t.Setenv(HomeEnv, "/abs/teicos")
	if !HomeOverrideSet() {
		t.Fatal("HomeOverrideSet() = false with the new variable set")
	}
}
