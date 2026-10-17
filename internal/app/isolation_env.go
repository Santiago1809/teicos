package app

import (
	piagent "github.com/Santiago1809/teicos/internal/agents/pi"
	"github.com/Santiago1809/teicos/internal/backup"
	"github.com/Santiago1809/teicos/internal/components/engram"
	"github.com/Santiago1809/teicos/internal/state"
)

// validateIsolationEnv rejects an invalid TEICOS_HOME,
// GENTLE_AI_PI_CONFIG_DIR or GENTLE_AI_ENGRAM_BIN before detection,
// self-update or any write, then applies TEICOS_HOME to backup checks.
func validateIsolationEnv() error {
	for _, validate := range []func() error{state.ValidateHomeEnv, piagent.ValidateGentlePiConfigDirEnv, engram.ValidateCallerBinary} {
		if err := validate(); err != nil {
			return err
		}
	}
	applyStateHome()
	return nil
}

// applyStateHome points backup deletion checks at TEICOS_HOME/backups,
// where install and sync write backups when that variable is set.
func applyStateHome() {
	if !state.HomeOverrideSet() {
		return
	}
	backup.BackupRootFn = func() (string, error) { return state.BackupsDir(""), nil }
}
