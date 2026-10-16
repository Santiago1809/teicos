# Teicos — identity & payload replacement

## Objective

Turn this Gentle-AI fork into Teicos (`github.com/Santiago1809/teicos`): new binary name,
new module path, review subsystem removed, theme assets repainted, central skill asset
replaced by a security-context layer, identity strings swapped.

## Problem

The fork still carries the upstream identity (module path, binary name, CLI strings,
docs) and payload that does not belong to a security-only configurator: the whole
receipt-driven-development review subsystem and the ODD/SDD/RDD skill corpus.

## Why

Teicos is a minimalist AI-agent configurator that injects a security context layer.
It is not a workflow tool and ships no development methodologies.

## Scope

- `cmd/gentle-ai` -> `cmd/teicos`, module path -> `github.com/Santiago1809/teicos`
- Delete the review subsystem (CLI files, 3 packages, consent envelope, dispatch entry)
- Repaint theme assets to the Teicos palette; rename theme ids
- Replace the skills asset corpus with `skills/security/teicos-security-context.md`
- Identity string swap + README placeholder

## Out of scope (untouched by instruction)

- `internal/components/permissions/`, `internal/components/filemerge/`,
  `internal/components/engram/`, `internal/pipeline/`

## Constraints

- No new logic: only delete and repair.
- After each deletion step run `go vet ./...` and repair only error caused by the deletion.
- Stop rule: if a step produces more than 10 compile errors unrelated to what was deleted,
  stop and report before continuing.

## Tasks

- [x] T1 — Rename binary + module path; `go build ./...` — done, `cmd/teicos`, module `github.com/Santiago1809/teicos`, 787 files / 1911 occurrences, build OK, vet exit 0
- [ ] T2 — Delete review *surface* only (cli `review*.go`, `reviewerprovider`, `reviewassets`, `consentenvelope`, dispatch entry in `internal/app/app.go:131`). **`internal/reviewtransaction` stays** (owner decision): it holds the hardened lock primitive that `internal/filecoord` and therefore `internal/components/engram` depend on
- [ ] T3 — Repaint theme palette in `internal/components/theme/inject.go` (Go literals, NOT assets); rename `gentleman`/`gentleman-cute` -> `teicos`/`teicos-dark`
- [ ] T4 — Replace skills corpus; register the security skill correctly (see Blockers)
- [ ] T5 — Identity string swap + README placeholder

## Blockers / corrections found

- T3 target corrected: no Kanagawa files under `internal/assets/`. Palette lives in
  `internal/components/theme/inject.go` (`gentlemanClaudeTheme`, `gentlemanCuteClaudeTheme`,
  `gentlemanOpenCodeTheme`, `gentlemanCuteOpenCodeTheme`, `themeOverlayJSON`, `VisualThemePaths`).
- T4 as specified would fail **silently at runtime**, not at build time:
  - `inject.go:69,78` resolves a skill as `"skills/" + string(skillID)`, so the asset must live in
    `internal/assets/skills/<skillID>/`, and an unregistered directory is only logged and skipped.
  - A `SkillID` constant (`internal/model/types.go:122-146`) and a `catalog/skills.go` `mvpSkills`
    entry are required for the skill to be installed at all.
  - Repo convention is `SKILL.md` per skill dir (see `docs/skill-style-guide.md`).
  - `SharedSkillDir` (`internal/assets/shared_skill_files.go:12`) is the `_shared` *references* root,
    not a skill body. Repointing it to `skills/security` mismatches that contract, and
    `support.go:54-59` errors hard if the listing is empty — so `skills/_shared/` must not be emptied.
- T5 collateral from T1: `internal/update/instructions.go:26,33` hardcodes the upstream module path
  **without `/v4`**, so the `sed` missed it; `isGentleAIRepo()` validates owner/repo. Also
  `.goreleaser.yaml`, `internal/releasepolicy/policy.go:561`, `scripts/test-install-module-path.sh`.

## Progress log

- PASO 1 complete and verified.
- PASO 2 blocked on `engram -> filecoord -> reviewtransaction`; resolved as "keep reviewtransaction".
- Security context layer draft written to `internal/assets/skills/security/teicos-security-context.md`.
  Not yet wired: needs the `SkillID` + `mvpSkills` entry + `SKILL.md` placement decision.
- Security context layer delivered (T4 part 1): `internal/assets/skills/security/SKILL.md` (8 rules) plus the
  full always-on contract in `internal/assets/generic/teicos-security-context.md`. `model.SkillSecurity` added
  to `mvpSkills` and to `selectableFoundationSkills` (line 36) — deliberately NOT `orchestrationSkills`, so the
  minimal preset stays judgment-day-only and `TestSkillsForPresetMinimalRetainsJudgmentDay` is untouched.
  The generic contract is still **inert**: not yet wired into an injection path.
- 13 external skills vendored (T4 part 2) with attribution in
  `internal/assets/skills/THIRD_PARTY_NOTICES.md`; licenses verified against the upstream `LICENSE` files
  (both MIT). `TestEmbeddedAssetCount` 17 -> 30.
- Frontmatter normalization: all 13 vendored `description` fields rewritten to
  `"Trigger: <upstream keywords>. <one-line description>"`, longest 151 chars, bodies untouched, modification
  recorded per repo in `THIRD_PARTY_NOTICES.md`. `TestSkillFrontmatterIsLintClean` green.
- `ARCHITECTURE.md` created at repo root, section "Why skills are embedded at compile time", recording
  compile-time embedding as a v1 security decision and naming `fs.FS` in
  `internal/components/skills/inject.go` as the seam to preserve.
- Evidence: `go test ./internal/assets/...` ok; `go vet ./...` exit 0.

### Camino A — `when_to_use` wired through the registry (Spanish activation)

- `SkillEntry` gained `WhenToUse string`; `parseFrontmatter` now returns `(name, description, whenToUse)`
  and handles the `when_to_use` key; `RenderRegistry` appends it to the `Trigger / description` cell,
  space-separated, only when non-empty; `when_to_use` added to the frontmatter lint whitelist.
- Two callers had to be updated, not one: the two early `return "", ""` guards in `parseFrontmatter`, and
  `registry_test.go` `TestParseFrontmatterHandlesCRLF` which destructured two return values.
- Added `TestRenderRegistryAppendsWhenToUseToTriggerColumn` for the new concatenation branch, and a
  `when_to_use` case to the existing parser table test.
- `docs/skill-registry.md` updated: the `Trigger / description` field now includes `when_to_use`.
- All 13 vendored skills carry Spanish `when_to_use`. Max combined `description` + `when_to_use` is 247 chars
  against the real Claude Code cap of 1,536 for the combined text in the skill listing.
- Verified end-to-end, not just by unit test: built the binary, ran `teicos skill-registry refresh --force`
  on a temp project holding `tdd` and `code-review`, and confirmed the generated `.atl/skill-registry.md`
  row contains the Spanish phrases.
- NO golden file guards the rendered table. `RenderRegistry` has no test caller; `assertRegistrySkills`
  checks path presence and ordering only; `assertSkillEntries` uses struct equality, which stays valid
  because no fixture sets `WhenToUse`.
- Evidence: `go test ./internal/skillregistry/... ./internal/assets/...` ok; `go vet ./...` exit 0.

### PASO 5a — release/update identity cluster (repoint to the fork)

User decision: repoint the production sites to `github.com/Santiago1809/teicos` + `cmd/teicos`, keep the
`/vN` derivation, and repair the ~40 corrupted test expectations.

Sites repointed: `internal/update/{instructions,registry,check,advisory}.go`,
`internal/update/upgrade/{strategy,executor}.go`, `internal/releaseprovenance/provenance.go`,
`internal/releaseprovenancecmd/main.go`, `internal/releasepolicy/policy.go`, `cmd/gentle-telemetry/main.go`,
`.goreleaser.yaml`, `scripts/install.sh`, `scripts/install.ps1`, `scripts/verify-release-assets.sh`.

Two hints hardcoded the upstream owner in a URL built for *any* tool; they now use `r.Tool.Owner`, which is
also correct for engram/gga.

**Corrections to earlier reports**

1. `claudeModuleOverheadCeiling` was never the blocker. Before injection
   `core 78144, monolith 82675, overhead 1635`; after `core 81834, monolith 86365, overhead 1635` —
   identical. The default Claude monolith comes from the same `InjectRoutingWithOptions`, so it grows with
   the injection and `overhead = core + modules − monolith` cancels. The earlier reading compared a fresh
   core against a **stale** monolith baseline in testdata. Const is back at **4096**, original comment
   intact, no rationale paragraph.
2. The rename `sed` was destructive: it matched `github.com/gentleman-programming/gentle-ai/v4` and ate the
   `/v4`, corrupting expectations into a state no correct production could satisfy. Repair ran in ordered
   passes (`/v4` restoration, then repo/binary/cmd rewrites, then owner/repo literals).
3. Two earlier repair passes were **wrong and reverted**:
   - `BetaModulePath` stays **unsuffixed** (`github.com/Santiago1809/teicos`): the beta composer appends
     `/cmd/teicos` to the module declared in go.mod at the checked commit, and the fork's module is
     unsuffixed. Restoring `/v4` broke `TestBetaTargetBindsModuleAndFullCommit`.
   - `validBetaModulePath` lowercased the expected base while the fork's module is
     `github.com/Santiago1809/teicos` (capital `S`), so every beta check failed closed. The `ToLower` is
     gone; the comparison is now exact-case — stricter, not looser, and required for the fork's own module.
4. `TestInstallScriptsGoInstallPackageMatchesModuleMajor` encoded a premise the fork does not satisfy: it
   required go.mod to carry a `/vN` suffix. Rewritten for the unsuffixed contract; the `/v[0-9]+/cmd/`
   staleness scan now flags any pinned major.

**Deliberate divergence (flagged, not decided)**

`BINARY_NAME` stays `gentle-ai` in both installers: GoReleaser still ships `dist/gentle-ai_*/gentle-ai` and
`internal/releasepolicy/policy.go:344-421` asserts those names, so renaming release artifacts is PASO 5.
The Go command directory is now `cmd/teicos`, so both installers pin that literal for the go-install package
and document why. Consequence: a source install produces a binary named `teicos` while the archive still
ships `gentle-ai`. Harmless today (the fork publishes no releases), but the names must converge when the
artifact rename lands.

**Still open**

- Homebrew: the tap stays upstream's (`Gentleman-Programming/homebrew-tap`); the fork has none, so Homebrew
  distribution cannot work yet.
- Versioning: `go.mod` is unsuffixed, so a real release must be tagged `v0.x`/`v1.x`. The `/vN` derivation
  in `ModulePathForVersion` is preserved as designed and still unit-tested.
- CI repo guards still gate on `github.repository == 'Gentleman-Programming/gentle-ai' || 'decode2/gentle-ai'`
  (`shell-linux-first-ci.yml:24`, `shell-macos-qualification.yml:60`). Whether the fork runs shell
  qualification is a CI-policy decision.
- Docs/community surface (README, CONTRIBUTING, PRD, ISSUE_TEMPLATE, PULL_REQUEST_TEMPLATE) still names the
  upstream project. PASO 5.

**Verification**

- `go build ./...` exit 0; `go vet ./...` exit 0.
- `go test` over `go list ./internal/...` minus `internal/cli` = **75 OK / 3 FAIL**, all three unrelated:
  - `internal/components/reviewassets` `TestOwnershipReuseRejectsLedgerWithoutMutation/hash_number` — the Go
    toolchain's JSON error now names `ownershipLedger.files.jd-judge-a.md` where the test pins
    `ownershipLedger.files`. Pre-existing.
  - `internal/shellinstaller` `TestUserFrozenLockAcquisition` — npm frozen-lock proof, `exit status 1`.
    Environmental.
  - `internal/releaseprovenancecmd` `TestCIBuildStillProducesReleaseProvenance` and
    `TestPartialReleaseIdentityNeverDowngrades/without_GITHUB_ACTIONS` — `goPattern` requires
    `^go1\.[0-9]+(?:\.[0-9]+)?$` and this machine reports `go1.27.0-X:nodwarf5`. Environmental; the pattern
    was NOT weakened to match the toolchain.
- `internal/cli` excluded on purpose: its suite invokes the native RDD review, which cannot obtain consent
  without a TTY and hangs the full 10-minute timeout.

---

## PASO 5 wave 1 — release/distribution identity (DONE)

Converged the product token in the release + distribution surface only: **32 files, 176 insertions /
172 deletions**. Verified tree-to-tree (`git diff <snapshot-tree> <now-tree>`), which is the only reliable
comparison here because PASO 1 left the whole rename uncommitted and untracked.

**Boundary.** `gentle-ai` → `teicos`, `Gentle-AI` / `Gentle AI` → `Teicos`, applied to: `.goreleaser.yaml`,
`internal/releasepolicy/policy.go` (assertions + embedded canonical copy), `internal/releaseprovenance`,
`internal/releaseprovenancecmd` tests, `internal/update` release/distribution tests, `scripts/install.*`,
`scripts/verify-release-assets.sh`, `scripts/release*-preflight.sh`, `scripts/deadcode-ratchet.sh`, all
`.github/workflows/*` touched by name, the issue/PR templates, the four e2e Dockerfiles, `e2e/lib.sh`,
`e2e/shell-*.{,sh,py}`, `.gitignore`.

**Deliberately masked (protocol/safety identifiers, not brand).** Renaming these would be a protocol or
safety change, so they keep `gentle-ai`:
- `gentle-ai.review-integration` — the RDD wire contract, shared with `contracts/` and the provider bundles.
- `gentle-ai-review-provider-contract` — artifact of the review surface, which T2 deletes.
- `gentle-ai-release-policy-snapshot-start`, `gentle-ai-release.key`, `gentle-ai-release-signing.pub`.

**Real bugs found and fixed, not cosmetic:**
- PASO 1's `git mv cmd/gentle-ai cmd/teicos` silently broke every build reference to the command directory:
  `ci.yml` (3 sites), `shell-linux-first-ci.yml`, `shell-macos-qualification.yml` (path filters AND build),
  `scripts/deadcode-ratchet.sh` (the deadcode ratchet was scanning a non-existent package), the four e2e
  Dockerfiles, `e2e/lib.sh`, `e2e/shell-windows-crosscheck-guest.sh`, `e2e/shell-macos-ci-contract.test.py`
  (its `REQUIRED_PATHS` contract still asserted `cmd/gentle-ai/main.go`), and `.gitignore`.
- `scripts/install.ps1` built the *stable* go-install package as `"{0}/cmd/{1}" -f $module, $BINARY_NAME`
  while the beta path hardcoded `cmd/teicos`. With `BINARY_NAME` changed, stable was resolving
  `.../cmd/gentle-ai@<tag>` — a directory that does not exist. Both paths now name `cmd/teicos` literally,
  and `TestInstallScriptsGoInstallPackageMatchesModuleMajor` gained the stable-path case (RED before the
  fix, GREEN after).

**Golden updated (diff shown before the edit).** `internal/assets/issue_taxonomy_contract_test.go` pins a
SHA-256 of the region from `- type: checkboxes` to EOF for both issue forms. Wave 1 changed two brand
strings inside that region (the issues URL, and `Gentle AI Version` → `Teicos Version` /
`Which area of Gentle AI` → `Which area of Teicos`). No control id, `required: true` count, or option list
changed — confirmed by diffing the region before/after. Hashes updated to `732c6857…` and `004e4ec0…`, and
the pin was mutation-tested (a label edit still fails the test).

**Verification.** `go build ./...` exit 0; `go vet ./...` exit 0. `go test` over `go list ./internal/...`
minus `internal/cli` = **74 OK / 4 FAIL**. Three FAILs are the previously recorded environmental ones
(`reviewassets` hash_number, `shellinstaller` frozen-lock, `releaseprovenancecmd` go-version pattern). The
fourth was my own new breakage in `internal/assets` (the pinned form hash), now fixed.
The embedded canonical `.goreleaser.yaml` / `release.yml` copies in `policy.go` were proven semantically
in sync with the real files by a temporary test calling `validateYAMLSemantics` directly (deleted after —
`internal/releasepolicycmd` already performs this comparison at release time).

### Wave 2 / wave 3 — NOT done, and one needs your call

Wave 2 is not cosmetic, it is a **behaviour change**, so it wants an explicit nod:
- `GENTLE_AI_*` environment variables → `TEICOS_*` (543 occurrences, 153 files). Any script or shell profile
  a user already has stops working.
- `~/.gentle-ai/` state paths (upgrade backups) → `~/.teicos/`.
- `GentleAI*` Go identifiers (~209 in 47 files) and the `internal/update` tool `Name: "gentle-ai"`, which is
  the identity the updater matches on.

Wave 3: theme ids `gentleman`/`gentleman-cute` → `teicos`/`teicos-dark` (`internal/components/theme/inject.go`),
TUI/CLI display strings, and README/CONTRIBUTING/PRD/docs. Also a **product-copy decision**: the current
tagline is `Ecosystem, Frameworks, Workflows for AI coding agents`, which no longer describes Teicos.

### Distribution seam for the next step

Brew is left deliberately unresolved because it is the next step and it depends on a repository that does
not exist yet:
- `.goreleaser.yaml` `brews.repository` owner/name, `policy.go`'s `resolved Homebrew publisher changed`
  assertion, and `install.sh`'s `BREW_TAP` / `BREW_FORMULA_REF` still point at upstream's
  `Gentleman-Programming/homebrew-tap`. `BREW_FORMULA_REF` interpolates `${BINARY_NAME}`, so it currently
  reads `gentleman-programming/tap/teicos` — a formula that does not exist. That resolves the moment the
  tap owner is decided.
- Likewise `installcmd/resolver.go:347` still taps `Gentleman-Programming/homebrew-tap` (left untouched by
  standing instruction).
- No Scoop section exists in `.goreleaser.yaml` yet.
