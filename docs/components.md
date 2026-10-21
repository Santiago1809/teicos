# Components, Skills & Presets

> [!NOTE]
> These docs track `main`, which may include unreleased changes. For the latest release, see the [v4.0.0 docs](https://github.com/Gentleman-Programming/gentle-ai/tree/v4.0.0/docs).

← [Back to README](../README.md)

---

## Components

| Component | ID | Description |
|-----------|-----|-------------|
| Engram™ | `engram` | Persistent cross-session memory via MCP — auto-detection of project name, full-text search, git sync, project consolidation. See [engram repo](https://github.com/Gentleman-Programming/engram) |
| Skills | `skills` | Curated coding skill library |
| Context7 | `context7` | MCP server for live framework/library documentation |
| Persona | `persona` | Managed Gentleman/neutral persona injection, or unmanaged custom persona mode |
| Permissions | `permissions` | Security-first defaults and guardrails. Applied to Claude Code and OpenCode/Kilocode (the adapters with permissions overlay support). Default sensitive-paths deny list: `~/.ssh/*`, `~/.ssh/**/*`, `**/*.pem`, `**/*.key`, `**/.env*`, `~/.credentials/*`, `~/.aws/credentials`, `~/.config/gh/hosts.yml`, `~/Library/Keychains/*`, `**/secrets/*`, `**/*.p12`, `**/*.pfx`. OpenCode/Kilocode follow the Gentle Pi safety model: everything is allowed by design; recursive `rm` on `/`, `~`, `$HOME`, `.` or `..` (also behind `sudo`), hard reset, forced `git clean`, force push (also with `git -C <dir>`), `chmod -R 777` and recursive `chown` are denied; other recursive `rm`, `find -delete`, destructive SQL passed to `psql`/`mysql`/`mariadb`/`sqlite3`, `git push`, `git rebase`, `git branch -D` and `npm publish` ask; remote commands ask (#4324); and the sensitive paths are denied to `read` and `edit` (which covers write and patch). Like Pi, reading a secret through `bash` or the `grep` tool is not blocked. Wildcard rules cannot parse shell syntax the way Pi does, so other wrappers (`env`, `xargs`, `VAR=value`), nested shells, absolute executable paths, unusual flag permutations, and `rm -rf /*` are not guarded. |
| Theme | `theme` | Teicos theme overlay: Aioniq base with a deep red accent |

ODD (Organic Driven Development) is shared routing guidance, not a separate component to install. It is the only implementation workflow for direct and delegated work. See [ODD and recovery](usage.md#organic-driven-development-odd).

> **SDD is not part of Teicos.** The Spec-Driven Development workflow and its `sdd-*` skills were retired upstream in favor of ODD, and Teicos ships neither. A legacy `sdd` selection persisted in state is still read, then dropped from the current selection.

## Explicit install modifiers

Explicit `--persona` and `--skill` / `--skills` options require their consumer
component in the resolved install plan: `persona` and `skills`, respectively.
An orphaned modifier is rejected before self-update or any install files or state
are written, including with `--dry-run`. The validated plan is reused for execution.
The diagnostic names `gentle-ai install` as the continuation. Rerun it with the
required `--component` added to your existing component list, or remove the
modifier. Keep your other agents and flags unchanged; components are not added
automatically.

```bash
gentle-ai install --agent claude-code --component theme,persona --persona neutral --dry-run
gentle-ai install --agent claude-code --component skills --skill go-testing --dry-run
```

Implicit defaults do not require additional components. `--persona custom` remains
an opt-out from managed persona and does not require `persona`.
retired and is not accepted by install.

**Compatibility change:** commands that previously succeeded while silently
ignoring an explicit modifier now fail with a diagnostic and a nonzero exit code.

## Context7 sync behavior

Install and sync add the agent-specific Context7 default when its server entry is absent. Existing entries are migrated only if the complete parsed entry exactly matches a known default written by a released gentle-ai version, such as the legacy unpinned or pinned `2.2.5` npx launcher. Key order and whitespace do not matter; extra or changed fields do. Custom commands, arguments, environment, URLs, headers, and different version pins (including `4.2.0` or `@latest`) are preserved untouched. Remove a customized entry explicitly to install the current default again.

Claude Code cleans up an inert legacy `settings.json` entry only when it exactly matches a released managed default; customized entries are left alone.

## Primary remote-authorization guidance

Always-installed agent guidance includes a canonical remote-operation boundary, independent of the optional persona and permissions components. Local-development access does not authorize remote execution, transfer, or discovery/reuse of ambient SSH agents, ControlMaster sockets, credentials, or sessions. Ask for explicit destination, operation, and credential/session authorization; authorized work remains allowed within stricter user/runtime restrictions.

This first delivery covers the 15 non-Pi primary instruction carriers, not every executor role or named profile. Pi remains owned by `gentle-pi`. The behavioral section provides no sandbox or fresh-human-per-execution guarantee. Separately, OpenCode/Kilocode permission defaults ask for direct ssh/scp/sftp/rsync commands, without silently replacing personal allows or restrictions; existing installs must opt into permission sync. See [sync update instructions and limitations](usage.md#sync). Issue #4324 remains open for the remaining projections and native approvals.

## Optional Community Tools

Community Tools are opt-in and are not included by presets or automatic detection. Select them from the installer’s **Community Tools** screen.

| Tool | Behavior | Removal |
|---|---|---|
| CodeGraph | Installs its CLI and configures supported agent MCP/guidance integration. | Use CodeGraph’s upstream lifecycle commands. |

### CodeGraph initialization workspace

For project initialization, the explicit agent working directory is the target—not an enclosing Git repository. Check its local `.git` (directory or worktree file) and `.codegraph/` first. Ancestor Git-root discovery is only appropriate for an already-selected existing repository.

`gentle-ai codegraph init --cwd <project-root>` requires that target to be a Git root and rejects unsafe locations. For a new nested project without local Git, initialize Git there first (`git -C <project-root> init`) as part of the requested project setup, then retry CodeGraph initialization at the same target. Do not substitute the parent repository or initialize Git merely to answer a codebase question.

After each successful initialization, including one over an existing index, the command appends `.codegraph/` once to the worktree's local exclude file (`git rev-parse --git-path info/exclude`, shared by linked worktrees). CodeGraph's own `.codegraph/.gitignore` un-ignores itself, so without this the index directory would show up as untracked and leak into agents' diffs. Tracked files and `.gitignore` are never edited. When Git is unavailable or the target is not a Git work tree, the command prints a warning to stderr and initialization still succeeds.

---

## Skills

### Included Skills (installed by gentle-ai)

Skill files organized by category, embedded in the binary and injected into your agent's configuration:

#### Review

| Skill | ID | Description |
|-------|-----|-------------|
| Judgment Day | `judgment-day` | Parallel adversarial review — two independent judges review the same target |

#### Foundation

| Skill | ID | Description |
|-------|-----|-------------|
| Go Testing | `go-testing` | Go testing patterns including Bubbletea TUI testing |
| Skill Creator | `skill-creator` | Create new AI agent skills following the Agent Skills spec |
| Skill Improver | `skill-improver` | Audit and improve existing skills against the repository style guide |
| Branch & PR | `branch-pr` | PR creation workflow with conventional commits, branch naming, and issue-first enforcement |
| Issue Creation | `issue-creation` | Issue filing workflow with bug report and feature request templates |
| Skill Registry | `skill-registry` | Build an index of installed skills with triggers, scopes, and exact `SKILL.md` paths |
| Chained PR | `chained-pr` | Plan and create reviewable stacked/chained pull requests |
| Cognitive Doc Design | `cognitive-doc-design` | Write docs that reduce review and onboarding cognitive load |
| Comment Writer | `comment-writer` | Draft warm, direct collaboration comments and review replies |
| Work Unit Commits | `work-unit-commits` | Split implementation into reviewable work units |
| RDD Defect Workflow | `rdd-defect-workflow` | Guide receipt-driven defect work with truthful evidence and authority boundaries |

Of these, `go-testing`, `skill-creator`, `skill-improver`, `skill-registry`, `chained-pr`, `cognitive-doc-design` and `work-unit-commits` are installed by default with both the `full-gentleman` (Dev Stack + Polish) and `ecosystem-only` (Dev Stack) presets. `branch-pr`, `issue-creation`, `comment-writer` and `rdd-defect-workflow` are repository-contributor workflow skills: they remain selectable with an explicit `--skills <id>`, but no default preset installs them.

### Coding Skills (separate repository)

For framework-specific skills (React 19, Angular, TypeScript, Tailwind 4, Zod 4, Playwright, etc.), see [Gentleman-Programming/Gentleman-Skills](https://github.com/Gentleman-Programming/Gentleman-Skills). These are maintained by the community and installed separately by cloning the repo and copying skills to your agent's skills directory.

---

## Presets

| Preset | ID | What's Included |
|--------|-----|-----------------|
| Dev Stack + Polish | `full-gentleman` | All components (Engram + Skills + Context7 + Permissions + visual polish) + all skills |
| Dev Stack | `ecosystem-only` | Core components (Engram + Skills + Context7) + all skills |
| Memory Only | `minimal` | Engram only |
| Custom | `custom` | You choose components and skills manually while keeping any existing persona/settings unmanaged |

Persona is selected separately on the Persona screen and applied independently of the preset.
