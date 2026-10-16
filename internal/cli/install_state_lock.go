package cli

import "github.com/Santiago1809/teicos/internal/statecoord"

func withInstallStateLock(homeDir string, operation func() error) error {
	return statecoord.WithLock(homeDir, operation)
}
