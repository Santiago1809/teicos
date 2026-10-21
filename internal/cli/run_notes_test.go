package cli

import (
	"testing"

	"github.com/Santiago1809/teicos/internal/model"
	"github.com/Santiago1809/teicos/internal/planner"
	"github.com/Santiago1809/teicos/internal/verify"
)

func TestWithPostInstallNotesDoesNotChangeOtherComponents(t *testing.T) {
	// Set GOBIN and PATH to the same directory so that withGoInstallPathNote
	// detects that GOBIN is already in PATH and does not append a guidance note.
	gobin := "/usr/local/bin"
	t.Setenv("GOBIN", gobin)
	t.Setenv("PATH", gobin)

	report := verify.Report{Ready: true, FinalNote: "You're ready."}
	resolved := planner.ResolvedPlan{OrderedComponents: []model.ComponentID{model.ComponentEngram}}

	updated := withPostInstallNotes(report, resolved)
	if updated.FinalNote != report.FinalNote {
		t.Fatalf("FinalNote changed unexpectedly: %q", updated.FinalNote)
	}
}
