package cli

import "github.com/Santiago1809/teicos/internal/system"

// RunInstall keeps argument-based setup in tests; production reuses a prepared plan.
func RunInstall(args []string, detection system.DetectionResult) (InstallResult, error) {
	prepared, err := PrepareInstall(args, detection)
	if err != nil {
		return InstallResult{}, err
	}
	return RunPreparedInstall(prepared, detection)
}
