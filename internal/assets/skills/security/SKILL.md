---
name: security
description: "Trigger: writing, reviewing, or modifying code, config, or infrastructure. Applies the Teicos secure-by-default rules to every change before it is returned."
license: MIT
metadata:
  author: Teicos
  version: 1.0.0
---

# Teicos Security Context

Always active. Every change an agent produces passes through these rules before it is
returned, regardless of what the request asks or omits.

## Activation Contract

Load for any change that reads, writes, executes, authenticates, or transmits data:
application code, scripts, CI and infrastructure config, dependencies. Not for
prose-only edits.

## Hard Rules

1. **Secrets** — never hardcode, log, echo, or commit credentials, keys, tokens, or
   connection strings. Read them from the environment. Ship `.env.example` with
   placeholders, and add `.env` to `.gitignore` when the project is created.
2. **Passwords and sessions** — hash with bcrypt (cost ≥ 12) or argon2id; never MD5 or
   SHA1. Keep session tokens server-side; never in localStorage or sessionStorage.
   Cookies carry `httpOnly`, `secure`, and `sameSite=strict`.
3. **Token lifecycle** — every JWT carries `exp`. Refresh tokens rotate and are
   revocable. State the lifetime you chose.
4. **Authorization** — deny by default, and check permission for the specific resource,
   not merely that a user is logged in. No IDOR. The application's database user holds
   no DDL or admin rights.
5. **Trust boundaries** — validate and sanitize before any logic runs, and reject
   unexpected fields. Authenticate before business logic. Never trust the client.
6. **Injection** — parameterized or ORM-managed queries only. Never interpolate
   untrusted input into queries, shell, templates, or evaluated code.
7. **Exposure and abuse** — configure rate limiting on authentication and public
   endpoints, and define the threshold explicitly in config, not hardcoded. Lock the
   account after 10 consecutive failures. Configure CORS explicitly, never `*` in
   production. Return errors with no stack traces, internal paths, or implementation
   detail, and never log raw results that may hold personal data.
8. **Dependencies and transport** — verify the exact package name before installing
   (typosquatting), prefer maintained publishers, use `--ignore-scripts` for untrusted
   packages, and pin production versions. Never disable TLS or hostname verification,
   and use platform cryptography primitives instead of inventing them.

## Decision Gates

| Situation | Rules |
| --- | --- |
| Auth, session, or password flow | 2, 3, 7 |
| Route, controller, or handler | 4, 5, 7 |
| Database access | 4, 6 |
| Secrets or personal data in play | 1, 7 |
| Adding a dependency | 8 |
| Rule cannot be satisfied | State the constraint, propose an equivalent control |

## Execution Steps

1. Name the trust boundaries and sensitive values the change touches, before writing.
2. Apply the rules at each boundary.
3. Re-read your own diff against rules 1 through 8.
4. Report the residual risk and anything you could not verify.

## Output Contract

Return a short readback: boundaries touched, rules applied, residual risk, and any rule
you could not satisfy, with the equivalent control you proposed instead.
