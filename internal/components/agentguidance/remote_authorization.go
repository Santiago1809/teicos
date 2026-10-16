package agentguidance

import (
	"strings"

	"github.com/Santiago1809/teicos/internal/assets"
	"github.com/Santiago1809/teicos/internal/components/filemerge"
)

// InjectRemoteAuthorization projects the canonical behavioral boundary into a
// prompt without changing native permissions. Repeated projection is a no-op.
// Executor integrations can reuse it without importing coordinator routing.
func InjectRemoteAuthorization(prompt string) string {
	contract := assets.MustRead("generic/remote-authorization-contract.md")
	return filemerge.InjectMarkdownSection(prompt, "remote-authorization", strings.TrimSpace(contract))
}
