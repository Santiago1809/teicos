package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Santiago1809/teicos/internal/opencode"
)

// Presentation evidence never changes runtime selection or installation policy.
type openCodePresentationMsg struct{ major opencode.RuntimeMajor }

func openCodePresentationCommand() tea.Cmd {
	return func() tea.Msg {
		major, _ := opencode.DetectRuntimeMajor(context.Background())
		return openCodePresentationMsg{major: major}
	}
}
