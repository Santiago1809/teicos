package state

import (
	"fmt"
	"os"
	"path/filepath"
)

// HomeEnv names the folder that holds Teicos' own state: state.json, the
// managed-assets manifest and journal, telemetry state, install backups and
// the Pi CodeGraph manifest. A bundled shell sets it to keep that state inside
// its prefix. It must be absolute.
const HomeEnv = "TEICOS_HOME"

// Dir returns Teicos' state folder: HomeEnv when it holds an absolute path,
// otherwise <homeDir>/.teicos.
//
// Teicos is a separate product and never reads, adopts or writes another
// tool's state directory, so an existing ~/.gentle-ai is left untouched. A
// relative override is rejected by ValidateHomeEnv at startup and ignored here.
func Dir(homeDir string) string {
	if override := absoluteEnv(HomeEnv); override != "" {
		return filepath.Clean(override)
	}
	return filepath.Join(homeDir, StateDirName)
}

// HomeOverrideSet reports whether HomeEnv is set at all, so callers that only
// need to know "is state relocated" do not have to read the variable.
func HomeOverrideSet() bool {
	return os.Getenv(HomeEnv) != ""
}

// BackupsDir returns the folder that holds install and sync backups.
func BackupsDir(homeDir string) string {
	return filepath.Join(Dir(homeDir), "backups")
}

// ValidateHomeEnv rejects a relative state override.
func ValidateHomeEnv() error {
	override := os.Getenv(HomeEnv)
	if override != "" && !filepath.IsAbs(override) {
		return fmt.Errorf("%s must be an absolute path, got %q", HomeEnv, override)
	}
	return nil
}

func absoluteEnv(name string) string {
	if value := os.Getenv(name); filepath.IsAbs(value) {
		return value
	}
	return ""
}
