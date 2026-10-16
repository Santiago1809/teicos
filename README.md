# Teicos

> The wall in the code

<img src="https://img.shields.io/badge/agents-17-blue" alt="17 agents">

Teicos is a secure-by-default AI coding environment. It configures the AI coding
agent you already use — Claude Code, Cursor, OpenCode — and injects a persistent
security context that enforces safe patterns automatically, without you having to ask.

## How it works

Teicos works at the configuration layer, not the model layer. When you install it,
it injects two things into your agent:

- A **security context** loaded on every session: authentication, injection, secrets,
  authorization, transport, dependencies — always on, not opt-in.
- A set of **security and engineering skills** the agent can invoke when the context
  calls for them: crypto audit, SAST, supply chain analysis, TDD, code review, and more.

No new agent. No new IDE. The tools you already use, with a security profile they
were missing.

## Install

Requires Go 1.25.10+.

```bash
go install github.com/Santiago1809/teicos/cmd/teicos@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`, then:

```bash
teicos install --agent claude-code   # or cursor, opencode
```

Homebrew and Scoop support coming in v1.0.

## Skills included

| Category | Skill | Trigger |
|---|---|---|
| Security | `api-security-tester` | audit API, revisar endpoints |
| Security | `crypto-audit` | audit crypto, auditar cifrado |
| Security | `bandit-sast` | scan Python, escanear Python |
| Security | `socket-sca` | scan dependencies, escanear dependencias |
| Security | `security-headers-audit` | review headers, revisar headers |
| Security | `security-test-generator` | generate security tests |
| Security | `mobile-security` | mobile security, seguridad móvil |
| Security | `devsecops-pipeline` | security pipeline, CI/CD seguridad |
| Security | `docker-scout-scanner` | scan Docker image |
| Security | `pci-dss-audit` | PCI compliance, cumplimiento PCI |
| Engineering | `tdd` | TDD, red-green-refactor |
| Engineering | `diagnosing-bugs` | diagnose bug, diagnosticar bug |
| Engineering | `code-review` | review PR, revisar código |

## Requirements

- Go 1.25.10+
- One of: Claude Code, Cursor, OpenCode

## License

MIT — forked from [Gentleman-Programming/gentle-ai](https://github.com/Gentleman-Programming/gentle-ai).
See [THIRD_PARTY_NOTICES](internal/assets/skills/THIRD_PARTY_NOTICES.md) for vendored skill attributions.
