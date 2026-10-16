package agentguidance

import (
	"strings"

	"github.com/Santiago1809/teicos/internal/assets"
	"github.com/Santiago1809/teicos/internal/components/filemerge"
)

// InjectSecurityContext projects the canonical Teicos security context into a
// prompt without changing native permissions. Repeated projection is a no-op.
func InjectSecurityContext(prompt string) string {
	contract := assets.MustRead("generic/teicos-security-context.md")
	return filemerge.InjectMarkdownSection(prompt, "teicos-security-context", strings.TrimSpace(contract))
}
