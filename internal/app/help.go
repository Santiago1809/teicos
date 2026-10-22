package app

import (
	"fmt"
	"io"
)

func printHelp(w io.Writer, version string) {
	fmt.Fprintf(w, `teicos — The wall in the code (%s)

USAGE
  teicos                        Launch interactive TUI
  teicos <command> [flags]

COMMANDS
  install      Configure AI coding agents on this machine
  uninstall    Remove Teicos managed files from this machine
  sync         Sync agent configs and skills to current version
  codegraph init --cwd <project-root>
               Initialize a project's CodeGraph index; codegraph [init] --help shows local help
  skill-registry refresh
               Refresh .atl/skill-registry.md with cache-hit fast path
  update       Check for available updates
  upgrade      Apply updates to managed tools
  telemetry <status|enable|disable|preview|trigger> [--json]
               Anonymous, opt-out usage telemetry; preview shows the exact payload without sending it;
               trigger runs the opportunistic check for hosts that never call install/update/sync
  restore      Restore a config backup
  doctor       Run ecosystem health diagnostics (bounded tool version probes)
  version      Print version

FLAGS
  --help, -h    Show global help; every review subcommand also supports help

Run 'teicos help' for this message.
Documentation: https://github.com/Santiago1809/teicos
`, version)
}
