# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Teicos is a fork of [Gentle-AI](https://github.com/Gentleman-Programming/gentle-ai) (Go module `github.com/Santiago1809/teicos`, Go 1.25.10). It is a CLI installer that configures AI coding agents (Claude Code, Cursor, OpenCode, ...) by injecting an always-on **security context** plus a set of embedded **skills**. It works at the config layer, not the model layer. Much of the tree is still inherited Gentle-AI code (review subsystem, telemetry, OpenCode runtime plugins, `gentle-*` naming); expect upstream names and `gentle-ai` strings in places.

## Commands

```bash
go build -o teicos ./cmd/teicos        # build the CLI
go vet ./...
go test ./...                          # full suite (~900 test files)
go test ./internal/components/skills -run TestName -v   # single test
go run ./internal/gofmtcheck           # CI formatting gate (gofmt -l over tracked .go files)
./scripts/deadcode-ratchet.sh          # CI gate: fails if NEW unreachable functions appear
go test ./internal/components/ -run <Test> -update      # regenerate golden files (internal/components/golden_test.go)
```

There is no Makefile. CI lives in `.github/workflows/ci.yml`; it also runs shell/Python contract scripts under `scripts/` and builds with `-tags bench_fixture` for the bench fixture binary.

## Architecture

- `cmd/teicos/main.go` is the entrypoint; `cmd/gentle-telemetry` is a separate collector server (inherited).
- `internal/` holds everything else. Install flow: `app` / `cli` / `tui` drive `planner` + `pipeline`, which execute `components/*` (one package per injectable thing: `skills`, `agentguidance`, `persona`, `mcp`, `permissions`, `engram`, ...) against a per-agent `agents/*` adapter. `catalog`, `model`, `state`, `backup`, `verify`, and `doctor` support it.
- **Security context**: `internal/components/agentguidance/security_context.go` (+ `inject.go`) injects the always-on context; the source text lives under `internal/assets/skills/security/`. This is the product's core value.
- **Skills are embedded at compile time** via `go:embed` (`internal/assets/assets.go`), deliberately, as a supply-chain decision. Do not replace with a runtime fetcher; see `ARCHITECTURE.md`. The skills component reads through an `fs.FS` (`internal/components/skills/inject.go`) — keep that seam intact.
- Skills live in `internal/assets/skills/<name>/SKILL.md`. The skill registry parser (`internal/skillregistry`) reads `when_to_use`, and bundled skills carry Spanish activation phrases alongside English ones. Vendored skill attributions go in `internal/assets/skills/THIRD_PARTY_NOTICES.md`.
- Review subsystem (`reviewtransaction`, `reviewerprovider`, `filecoord`, `consentenvelope`) is inherited and not cleanly separable: `engram` -> `filecoord` -> `reviewtransaction` share a lock chain. Don't delete one piece in isolation.
- Release tooling (`releasepolicy`, `releaseprovenance`, `providercontractbundle` and their `*cmd` mains) is validated by GoReleaser config + `release.yml`; `contracts/review-provider-contract/CONTRACT_SEMVER` is the version source.

## Conventions

- Golden-file tests in `internal/components` mean any change to injected content (skills, persona, guidance) usually requires regenerating goldens with `-update` and reviewing the diff.
- Installer tests assert module-path and install behavior (`scripts/test-install-module-path.sh`); module path casing matters (`Santiago1809/teicos`).
- `odd/` is gitignored (local task notes). `.codegraph/` is a local index, not committed.
- Commits: Conventional Commits, no AI attribution lines.
