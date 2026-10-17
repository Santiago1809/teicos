package state

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

// HomeEnv names the folder that holds Teicos' own state: state.json, the
// managed-assets manifest and journal, telemetry state, install backups and
// the Pi CodeGraph manifest. A bundled shell sets it to keep that state inside
// its prefix. It must be absolute.
const HomeEnv = "TEICOS_HOME"

// HomeEnvLegacy is the name this variable had before the Teicos rename. It is
// still honoured, with a deprecation warning, when HomeEnv is not set.
const HomeEnvLegacy = "GENTLE_AI_HOME"

// warnOut receives the one-time warnings Dir emits. Tests replace it.
var warnOut io.Writer = os.Stderr

var (
	legacyEnvWarned atomic.Bool
	legacyDirWarned atomic.Bool
)

// Dir returns Teicos' state folder, resolved in this order:
//
//  1. HomeEnv, when set to an absolute path;
//  2. HomeEnvLegacy, when set to an absolute path (deprecation warning);
//  3. <homeDir>/.teicos, when it exists;
//  4. <homeDir>/.gentle-ai, when it exists (migration hint, never migrated);
//  5. <homeDir>/.teicos, the default for a new install.
//
// The two directories are never merged: the first one that applies wins. A
// relative override is rejected by ValidateHomeEnv at startup and ignored here.
func Dir(homeDir string) string {
	if override := absoluteEnv(HomeEnv); override != "" {
		return filepath.Clean(override)
	}
	if override := absoluteEnv(HomeEnvLegacy); override != "" {
		if legacyEnvWarned.CompareAndSwap(false, true) {
			fmt.Fprintf(warnOut, "teicos: %s is deprecated; set %s instead\n", HomeEnvLegacy, HomeEnv)
		}
		return filepath.Clean(override)
	}
	current := filepath.Join(homeDir, StateDirName)
	if isDir(current) {
		return current
	}
	legacy := filepath.Join(homeDir, LegacyStateDirName)
	if isDir(legacy) {
		if legacyDirWarned.CompareAndSwap(false, true) {
			fmt.Fprintf(warnOut, "teicos: using the legacy state directory ~/%s; to move to ~/%s run: cp -r ~/%s ~/%s\n",
				LegacyStateDirName, StateDirName, LegacyStateDirName, StateDirName)
		}
		return legacy
	}
	return current
}

// HomeOverrideSet reports whether HomeEnv or HomeEnvLegacy is set at all, so
// callers that only need to know "is state relocated" do not have to know
// there are two names.
func HomeOverrideSet() bool {
	return os.Getenv(HomeEnv) != "" || os.Getenv(HomeEnvLegacy) != ""
}

// BackupsDir returns the folder that holds install and sync backups.
func BackupsDir(homeDir string) string {
	return filepath.Join(Dir(homeDir), "backups")
}

// ValidateHomeEnv rejects a relative state override. HomeEnv is checked first;
// HomeEnvLegacy is only checked when HomeEnv is not set, matching Dir.
func ValidateHomeEnv() error {
	name := HomeEnv
	override := os.Getenv(name)
	if override == "" {
		name = HomeEnvLegacy
		override = os.Getenv(name)
	}
	if override != "" && !filepath.IsAbs(override) {
		return fmt.Errorf("%s must be an absolute path, got %q", name, override)
	}
	return nil
}

func absoluteEnv(name string) string {
	if value := os.Getenv(name); filepath.IsAbs(value) {
		return value
	}
	return ""
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
