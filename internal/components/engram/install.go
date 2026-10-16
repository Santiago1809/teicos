package engram

import (
	"github.com/Santiago1809/teicos/internal/installcmd"
	"github.com/Santiago1809/teicos/internal/model"
	"github.com/Santiago1809/teicos/internal/system"
)

func InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	return installcmd.NewResolver().ResolveComponentInstall(profile, model.ComponentEngram)
}
