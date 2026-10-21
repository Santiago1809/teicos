package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// legacyHomeEnv is the pre-rename variable. Teicos must ignore it, so every
// test clears it (and may set it) without the package ever reading it.
const legacyHomeEnv = "GENTLE_AI_HOME"

// clearHomeEnv unsets both state variables so every test starts from a known
// process-level state.
func clearHomeEnv(t *testing.T) {
	t.Helper()
	t.Setenv(HomeEnv, "")
	t.Setenv(legacyHomeEnv, "")
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestDirDefaultsToDotTeicosUnderHome(t *testing.T) {
	clearHomeEnv(t)
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
	if _, err := os.Stat(filepath.Join(home, ".teicos")); !os.IsNotExist(err) {
		t.Fatalf("Dir() must not create the directory (stat err = %v)", err)
	}
}

// Teicos is a separate product: it must never read, adopt or write another
// tool's state directory, and the pre-rename variable must stay inert.
func TestDirIgnoresLegacyToolState(t *testing.T) {
	t.Run("GENTLE_AI_HOME is ignored", func(t *testing.T) {
		clearHomeEnv(t)
		home := t.TempDir()
		t.Setenv(legacyHomeEnv, filepath.Join(t.TempDir(), "gentle-home"))

		if got, want := Dir(home), filepath.Join(home, ".teicos"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
		if HomeOverrideSet() {
			t.Fatal("HomeOverrideSet() = true for the legacy variable alone")
		}
	})

	t.Run("an existing ~/.gentle-ai is never adopted", func(t *testing.T) {
		clearHomeEnv(t)
		home := t.TempDir()
		mkdir(t, filepath.Join(home, ".gentle-ai"))

		if got, want := Dir(home), filepath.Join(home, ".teicos"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
		if _, err := os.Stat(filepath.Join(home, ".teicos")); !os.IsNotExist(err) {
			t.Fatalf("Dir() must not create ~/.teicos (stat err = %v)", err)
		}
	})

	t.Run("a regular file named .teicos still wins over ~/.gentle-ai", func(t *testing.T) {
		clearHomeEnv(t)
		home := t.TempDir()
		mkdir(t, filepath.Join(home, ".gentle-ai"))
		if err := os.WriteFile(filepath.Join(home, ".teicos"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}

		if got, want := Dir(home), filepath.Join(home, ".teicos"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
	})
}

func TestDirPrecedence(t *testing.T) {
	override := filepath.Join(t.TempDir(), "teicos-home")

	t.Run("TEICOS_HOME wins over every directory on disk", func(t *testing.T) {
		clearHomeEnv(t)
		home := t.TempDir()
		mkdir(t, filepath.Join(home, ".teicos"))
		mkdir(t, filepath.Join(home, ".gentle-ai"))
		t.Setenv(HomeEnv, override)
		t.Setenv(legacyHomeEnv, filepath.Join(t.TempDir(), "gentle-home"))

		if got := Dir(home); got != override {
			t.Fatalf("Dir() = %q, want %q", got, override)
		}
	})

	t.Run("relative overrides are ignored", func(t *testing.T) {
		clearHomeEnv(t)
		home := t.TempDir()
		t.Setenv(HomeEnv, "relative/teicos")
		t.Setenv(legacyHomeEnv, "relative/gentle")

		if got, want := Dir(home), filepath.Join(home, ".teicos"); got != want {
			t.Fatalf("Dir() = %q, want %q", got, want)
		}
	})
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
	// filepath.IsAbs is platform-specific: a POSIX-style "/abs/teicos" is not
	// absolute on Windows, so the accepted-path case must be built per platform.
	absolute := filepath.Join(os.TempDir(), "teicos-absolute")
	for _, tt := range []struct {
		name, value, legacy, wantErr string
	}{
		{"new variable", "relative/teicos", "", HomeEnv},
		{"legacy variable alone", "", "relative/gentle-ai", ""},
		{"new variable wins over a bad legacy one", absolute, "relative/gentle-ai", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clearHomeEnv(t)
			t.Setenv(HomeEnv, tt.value)
			t.Setenv(legacyHomeEnv, tt.legacy)

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
	t.Setenv(HomeEnv, "/abs/teicos")
	if !HomeOverrideSet() {
		t.Fatal("HomeOverrideSet() = false with the new variable set")
	}
}
