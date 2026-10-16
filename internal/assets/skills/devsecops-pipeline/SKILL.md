---
name: devsecops-pipeline
description: "Trigger: CI/CD security scanning, SAST, SCA, secrets scanning, container scanning, GitHub Actions, SARIF. Builds a DevSecOps pipeline."
when_to_use: "pipeline de seguridad, CI/CD seguridad, GitHub Actions seguridad, automatizar escaneo, DevSecOps"
---

# DevSecOps Pipeline Generator

This skill generates ready-to-commit GitHub Actions workflow YAML files for multi-stage security CI/CD pipelines. Unlike scanning skills that report findings or test generators that produce test code, this skill outputs complete `.github/workflows/security.yml` files with SAST, SCA, secrets detection, container scanning, and DAST stages — auto-configured for the detected project ecosystem. No external tool installation is required; the generated workflow uses GitHub-hosted actions that run in CI.

## When to Use

- When the user asks to "generate a security pipeline" or "create a security workflow"
- When the user mentions "DevSecOps", "CI/CD security", or "GitHub Actions security"
- When the user wants to "add security scanning to CI" or "set up automated security checks"
- When the user asks to "create a security.yml" or "generate a GitHub Actions security workflow"
- When a project has no existing security CI/CD pipeline and the user wants one generated
- When the user asks to "shift security left" or "automate security scanning"

## When NOT to Use

**DO NOT activate if the request is not about setting up a CI/CD security pipeline, even if the word "security" appears.** The presence of security keywords alone is not a trigger — the request must be about wiring security into a GitHub Actions workflow or equivalent CI/CD system.

- When the user wants to **run a security scan now** on code or dependencies — you **MUST** decline and recommend `bandit-sast` (Python), `crypto-audit` (cryptography), `security-review` (general), `docker-scout-scanner` (containers), or `socket-sca` (dependencies) as appropriate
- When the user wants to **generate security test code** — you **MUST** decline and recommend `security-test-generator`
- When the user is asking about security concepts without wanting a CI/CD pipeline generated — **Decline** and answer the conceptual question directly without generating a workflow
- When the user already has a security pipeline and wants to debug a specific GitHub Actions issue — **Decline** pipeline generation and help debug the existing workflow instead
- When the user wants to scan a live application — you **MUST** decline and recommend DAST tools like `dast-nuclei`

## Prerequisites

### Tool Installed (Preferred)

No external tool required. This skill generates GitHub Actions workflow YAML using project analysis only.

All pipeline configuration is produced through project file inspection and template synthesis. The generated workflow references published GitHub Actions (Semgrep, Trivy, Gitleaks, etc.) that execute in GitHub's CI environment — nothing needs to be installed locally.

### Tool Not Installed (Fallback)

This skill is always available as a pure config-generation skill. There is no fallback mode because there is no external tool dependency. The skill analyzes the project structure and produces a complete workflow file directly.

## Project Detection

Before generating the workflow, detect the project ecosystem by inspecting files in the repository root:

| File Detected | Ecosystem | SCA Tool | Semgrep Ruleset |
|---------------|-----------|----------|-----------------|
| `package.json` | Node.js | `npm audit` | `p/javascript` |
| `setup.py`, `pyproject.toml`, or `requirements.txt` | Python | `pip-audit` | `p/python` |
| `Dockerfile` or `docker-compose.yml` | Container | Trivy image scan | `p/docker` |
| `go.mod` | Go | `govulncheck` | `p/golang` |

If multiple ecosystems are detected, generate stages for **all** of them. For example, a project with `package.json` and `Dockerfile` gets Node.js SCA, container scanning, and both JavaScript and Docker Semgrep rulesets.

## Workflow Structure

The generated workflow file follows this anatomy:

```yaml
name: Security Pipeline

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]
  schedule:
    # Weekly scan: Monday 6am UTC
    - cron: '0 6 * * 1'

permissions:
  security-events: write  # Required for SARIF upload to GitHub Security tab
  contents: read          # Required for actions/checkout
  actions: read           # Required for github/codeql-action/upload-sarif

concurrency:
  group: security-${{ github.ref }}
  cancel-in-progress: true  # Cancel in-progress runs for the same PR to save CI minutes
```

**Key design decisions:**

- **Triggers:** Push to main/master catches merged vulnerabilities. PR triggers catch issues before merge. Weekly schedule catches newly disclosed CVEs in existing dependencies.
- **Permissions:** Minimal permissions — only what each action needs. `security-events: write` is required for SARIF upload to the GitHub Security tab.
- **Concurrency:** Cancels in-progress runs for the same PR/branch to avoid wasting CI minutes on superseded commits.
- **Job parallelism:** SAST, SCA, and Secrets Detection run in parallel (no dependencies). Container Scanning depends on a Docker build step. DAST depends on deployment to a staging environment.
- **Runner:** Each job uses `ubuntu-latest`.
- **Checkout:** Each job starts with `actions/checkout@v4`.

> **Anti-hallucination guardrail:** All GitHub Actions referenced in this skill are real, published actions. Verify versions against GitHub Marketplace before committing. Do NOT invent action names or flags — if unsure about a flag, omit it and add a TODO comment.

## Workflow

1. **Detect project ecosystem** — Inspect the repository for `package.json`, `setup.py`/`pyproject.toml`/`requirements.txt`, `Dockerfile`/`docker-compose.yml`, and `go.mod` to determine which languages and runtimes are in use.
2. **Select pipeline stages** — Based on detected ecosystems, determine which of the 5 stages to include and which tool variants to use within each stage.
3. **Generate the workflow YAML** — Produce a complete `.github/workflows/security.yml` file with the selected stages, using the templates in the Pipeline Stages section below.
4. **Configure severity thresholds** — Set each stage to fail on Critical/High findings, warn on Medium, and ignore Low by default. Document how to adjust thresholds.
5. **Add SARIF upload steps** — For tools that support SARIF output (Semgrep, Trivy), add `github/codeql-action/upload-sarif@v3` upload steps so findings appear in the GitHub Security tab.
6. **Add artifact upload steps** — For all stages, upload scan results as artifacts using `actions/upload-artifact@v4` for post-run review.
7. **Adapt to the specific project** — The generated workflow is a starting point. Instruct the user to review and customize: adjust severity thresholds, add/remove stages, configure notification channels, and set up branch protection rules.
8. **Output the complete YAML** — Present the full workflow file in a single code block that the user can copy-paste into `.github/workflows/security.yml`.

## Pipeline Stages

### Stage 1: SAST (Static Application Security Testing)

**WHY:** SAST catches injection vulnerabilities (SQL injection, XSS, command injection), hardcoded secrets, and insecure code patterns before code merges to main. Skipping SAST means these issues reach production where they become exploitable attack vectors.

```yaml
  sast:
    name: SAST - Semgrep
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Semgrep
        uses: returntocorp/semgrep-action@v1
        with:
          config: auto
        env:
          SEMGREP_RULES: p/default

      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: semgrep.sarif
          category: semgrep

      - name: Upload results artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: semgrep-results
          path: semgrep.sarif
```

**Configuration notes:**
- `config: auto` enables Semgrep's recommended rulesets for detected languages. Override with specific rulesets: `config: p/javascript p/typescript p/react` for a React project.
- To add custom rules, create a `.semgrep.yml` in the repository root:

```yaml
rules:
  - id: custom-no-eval
    pattern: eval(...)
    message: "Do not use eval() — it enables code injection"
    languages: [javascript, typescript]
    severity: ERROR
    metadata:
      cwe: "CWE-95"
      owasp: "A03:2021"
```

- Then reference it: `config: auto .semgrep.yml`

### Stage 2: SCA (Software Composition Analysis)

**WHY:** SCA identifies known vulnerabilities (CVEs) in third-party dependencies. A single vulnerable dependency can expose the entire application — 60%+ of code in modern apps comes from open-source packages. Skipping SCA means you ship known-exploitable code.

The SCA stage uses conditional jobs based on the detected ecosystem:

```yaml
  sca-node:
    name: SCA - npm audit
    runs-on: ubuntu-latest
    if: hashFiles('package-lock.json') != ''
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Run npm audit
        run: |
          npm audit --audit-level=high --json > npm-audit-results.json || true
          npm audit --audit-level=high
        # --audit-level=high: fail on High and Critical only
        # --omit=dev: add this flag to skip devDependencies

      - name: Upload results artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: npm-audit-results
          path: npm-audit-results.json

  sca-python:
    name: SCA - pip-audit
    runs-on: ubuntu-latest
    if: hashFiles('requirements.txt') != '' || hashFiles('pyproject.toml') != ''
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run pip-audit
        uses: pypa/gh-action-pip-audit@v1.1.0
        with:
          inputs: requirements.txt

  sca-go:
    name: SCA - govulncheck
    runs-on: ubuntu-latest
    if: hashFiles('go.mod') != ''
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run govulncheck
        uses: golang/govulncheck-action@v1.0.4
        with:
          go-version-input: 'stable'
```

### Stage 3: Secrets Detection

**WHY:** Hardcoded secrets (API keys, passwords, tokens) in source code are the most common cause of credential leaks. Once a secret is pushed to git history, it persists even after deletion and can be extracted by anyone with repo access. Secrets detection prevents this before the commit reaches the remote.

```yaml
  secrets:
    name: Secrets Detection - Gitleaks
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2.3.7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Configuration notes:**
- `fetch-depth: 0` is required for full repo scan on push to main. For PR diff scans, Gitleaks automatically detects the PR context.
- To suppress false positives, create a `.gitleaks.toml` in the repository root:

```toml
title = "Gitleaks Config"

[allowlist]
  description = "Allowlist for false positives"
  paths = [
    '''tests/fixtures/.*''',
    '''docs/examples/.*'''
  ]
  regexes = [
    '''EXAMPLE_API_KEY''',
    '''test-token-[a-z]+'''
  ]
```

### Stage 4: Container Scanning

**WHY:** Container images bundle OS packages and language libraries that may contain known CVEs. A vulnerable base image (e.g., outdated OpenSSL, curl, or glibc) creates an entry point regardless of how secure the application code is. Container scanning catches these OS-level and library-level vulnerabilities that SCA misses.

```yaml
  container-scan:
    name: Container Scanning - Trivy
    runs-on: ubuntu-latest
    if: hashFiles('Dockerfile') != ''
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build Docker image
        run: docker build -t ${{ github.repository }}:${{ github.sha }} .

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@0.28.0
        with:
          image-ref: '${{ github.repository }}:${{ github.sha }}'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
          # --ignore-unfixed: uncomment to skip vulnerabilities with no available fix
          # ignore-unfixed: true

      - name: Upload Trivy SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif
          category: trivy

      - name: Upload results artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: trivy-results
          path: trivy-results.sarif
```

**Configuration notes:**
- `severity: 'CRITICAL,HIGH'` fails only on Critical and High findings. Adjust to include MEDIUM if desired.
- For filesystem scanning (without building the image), change `image-ref` to `scan-type: 'fs'` and `scan-ref: '.'`.
- The `ignore-unfixed` flag skips vulnerabilities that have no patch available yet — useful for reducing noise in reports.

### Stage 5: DAST (Dynamic Application Security Testing)

**WHY:** DAST tests the running application from an attacker's perspective, catching issues that static analysis misses: misconfigured CORS, missing security headers, exposed admin panels, authentication bypass, and runtime injection vulnerabilities. It validates that security controls actually work at runtime.

> **Important:** This stage is commented out by default because it requires a deployed staging environment. Uncomment and configure `STAGING_URL` to enable.

```yaml
  # ============================================================
  # DAST - Nuclei (UNCOMMENT TO ENABLE)
  # Requires: a deployed staging environment
  # Set vars.STAGING_URL in your repository settings
  # (Settings > Secrets and variables > Actions > Variables)
  # ============================================================
  # dast:
  #   name: DAST - Nuclei
  #   runs-on: ubuntu-latest
  #   # Uncomment if you have a deploy job:
  #   # needs: [deploy-staging]
  #   steps:
  #     - name: Checkout code
  #       uses: actions/checkout@v4
  #
  #     - name: Run Nuclei scanner
  #       uses: projectdiscovery/nuclei-action@v2.0.1
  #       with:
  #         target: ${{ vars.STAGING_URL }}
  #         templates: cves/,misconfiguration/
  #         severity: critical,high
  #         output: nuclei-results.txt
  #         # rate-limit: 50
  #
  #     - name: Upload results artifact
  #       if: always()
  #       uses: actions/upload-artifact@v4
  #       with:
  #         name: nuclei-results
  #         path: nuclei-results.txt
```

## CWE Mapping

Each pipeline stage catches specific vulnerability classes:

| Stage | Vulnerability Classes | CWEs |
|-------|----------------------|------|
| SAST (Semgrep) | Injection, XSS, command injection, hardcoded secrets, insecure patterns | CWE-89, CWE-79, CWE-78, CWE-798, CWE-327 |
| SCA (npm audit / pip-audit / govulncheck) | Known CVEs in dependencies, vulnerable library versions | CWE-1395 |
| Secrets Detection (Gitleaks) | Hardcoded credentials, API keys, tokens in source code | CWE-798 |
| Container Scanning (Trivy) | OS package vulnerabilities, library CVEs in container images | CWE-1395 |
| DAST (Nuclei) | Runtime misconfigurations, missing headers, exposed endpoints | CWE-693, CWE-16 |

### OWASP Top 10:2021 Coverage

| OWASP Category | Stages Covering It |
|----------------|--------------------|
| A01:2021 - Broken Access Control | DAST |
| A02:2021 - Cryptographic Failures | SAST |
| A03:2021 - Injection | SAST, DAST |
| A04:2021 - Insecure Design | SAST (partial) |
| A05:2021 - Security Misconfiguration | DAST, Container Scanning |
| A06:2021 - Vulnerable and Outdated Components | SCA, Container Scanning |
| A07:2021 - Identification and Authentication Failures | SAST, DAST |
| A08:2021 - Software and Data Integrity Failures | SCA |
| A09:2021 - Security Logging and Monitoring Failures | (not covered by CI pipeline) |
| A10:2021 - Server-Side Request Forgery | SAST, DAST |

## Findings Format

This skill produces pipeline configuration rather than vulnerability findings. Each row in the findings table describes a generated pipeline stage:

| Field | Description |
|-------|-------------|
| Stage Name | Name of the pipeline stage (SAST, SCA, Secrets, Container, DAST) |
| Tools Used | GitHub Action(s) referenced in the stage |
| Vulnerability Classes | CWE categories the stage detects |
| Threshold | Severity threshold for job failure |
| Output Format | SARIF, JSON, or text |

### Example Findings Table

| Stage Name | Tools Used | Vulnerability Classes | Threshold | Output Format |
|------------|------------|-----------------------|-----------|---------------|
| SAST | `returntocorp/semgrep-action@v1` | CWE-89, CWE-79, CWE-78, CWE-798 | Configurable via rules | SARIF |
| SCA (Node.js) | `npm audit` | CWE-1395 | High (`--audit-level=high`) | JSON |
| SCA (Python) | `pypa/gh-action-pip-audit@v1.1.0` | CWE-1395 | High | Text |
| SCA (Go) | `golang/govulncheck-action@v1.0.4` | CWE-1395 | All | Text |
| Secrets Detection | `gitleaks/gitleaks-action@v2.3.7` | CWE-798 | All secrets | Text |
| Container Scanning | `aquasecurity/trivy-action@0.28.0` | CWE-1395 | Critical, High | SARIF |
| DAST | `projectdiscovery/nuclei-action@v2.0.1` | CWE-693, CWE-16 | Critical, High | Text |

## Customization

After generating the workflow, users should adapt it to their specific needs:

### Changing Severity Thresholds

- **npm audit:** Change `--audit-level=high` to `--audit-level=critical` (stricter) or `--audit-level=moderate` (more permissive)
- **Trivy:** Change `severity: 'CRITICAL,HIGH'` to include `MEDIUM` or restrict to `CRITICAL` only
- **Semgrep:** Configure rule severity in `.semgrep.yml` or use `--severity ERROR` to fail only on errors

### Adding Notifications

To add Slack notifications on security scan failures:

```yaml
  # Add this step to any job that should notify on failure:
  # - name: Notify Slack on failure
  #   if: failure()
  #   uses: slackapi/slack-github-action@v1.27.0
  #   with:
  #     payload: |
  #       {
  #         "text": "Security scan failed in ${{ github.repository }}: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}"
  #       }
  #   env:
  #     SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Branch Protection

After adding the security pipeline, configure branch protection rules:

1. Go to **Settings > Branches > Branch protection rules**
2. Enable **Require status checks to pass before merging**
3. Add the security jobs: `sast`, `sca-node`, `secrets`, `container-scan`
4. This prevents merging PRs that fail security checks

### GITHUB_TOKEN Permissions

For SARIF upload to work, the workflow needs `security-events: write` permission. This is set at the workflow level. If your repository uses restricted token permissions, ensure the security workflow is allowed to write security events in **Settings > Actions > General > Workflow permissions**.

## Reference Tables

### GitHub Actions Reference

| Action | Version | Purpose |
|--------|---------|---------|
| `actions/checkout` | `@v4` | Clone repository |
| `actions/upload-artifact` | `@v4` | Upload scan results |
| `actions/setup-node` | `@v4` | Install Node.js for npm audit |
| `github/codeql-action/upload-sarif` | `@v3` | Upload SARIF to Security tab |
| `returntocorp/semgrep-action` | `@v1` | SAST scanning |
| `pypa/gh-action-pip-audit` | `@v1.1.0` | Python SCA |
| `golang/govulncheck-action` | `@v1.0.4` | Go SCA |
| `gitleaks/gitleaks-action` | `@v2.3.7` | Secrets detection |
| `aquasecurity/trivy-action` | `@0.28.0` | Container scanning |
| `projectdiscovery/nuclei-action` | `@v2.0.1` | DAST scanning |
| `slackapi/slack-github-action` | `@v1.27.0` | Slack notifications |

### CWE Reference

| CWE ID | Name |
|--------|------|
| CWE-16 | Configuration |
| CWE-78 | Improper Neutralization of Special Elements used in an OS Command |
| CWE-79 | Improper Neutralization of Input During Web Page Generation (XSS) |
| CWE-89 | Improper Neutralization of Special Elements used in an SQL Command |
| CWE-327 | Use of a Broken or Risky Cryptographic Algorithm |
| CWE-693 | Protection Mechanism Failure |
| CWE-798 | Use of Hard-coded Credentials |
| CWE-1395 | Dependency on Vulnerable Third-Party Component |

## Example Usage

### Node.js + Docker Project

**User prompt:**
> "Generate a security pipeline for this project"

**Project files detected:** `package.json`, `Dockerfile`

**Generated `.github/workflows/security.yml`:**

```yaml
name: Security Pipeline

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]
  schedule:
    - cron: '0 6 * * 1'

permissions:
  security-events: write
  contents: read
  actions: read

concurrency:
  group: security-${{ github.ref }}
  cancel-in-progress: true

jobs:
  sast:
    name: SAST - Semgrep
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Semgrep
        uses: returntocorp/semgrep-action@v1
        with:
          config: auto
        env:
          SEMGREP_RULES: p/javascript p/typescript p/docker

      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: semgrep.sarif
          category: semgrep

      - name: Upload results artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: semgrep-results
          path: semgrep.sarif

  sca-node:
    name: SCA - npm audit
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Run npm audit
        run: |
          npm audit --audit-level=high --json > npm-audit-results.json || true
          npm audit --audit-level=high

      - name: Upload results artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: npm-audit-results
          path: npm-audit-results.json

  secrets:
    name: Secrets Detection - Gitleaks
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2.3.7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  container-scan:
    name: Container Scanning - Trivy
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build Docker image
        run: docker build -t ${{ github.repository }}:${{ github.sha }} .

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@0.28.0
        with:
          image-ref: '${{ github.repository }}:${{ github.sha }}'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'

      - name: Upload Trivy SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif
          category: trivy

      - name: Upload results artifact
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: trivy-results
          path: trivy-results.sarif
```

### Python-Only Project

**User prompt:**
> "Add security scanning to CI for my Python API"

**Project files detected:** `requirements.txt`, `pyproject.toml`

**Generated `.github/workflows/security.yml` (abbreviated — shows key differences):**

```yaml
name: Security Pipeline

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]
  schedule:
    - cron: '0 6 * * 1'

permissions:
  security-events: write
  contents: read
  actions: read

concurrency:
  group: security-${{ github.ref }}
  cancel-in-progress: true

jobs:
  sast:
    name: SAST - Semgrep
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Semgrep
        uses: returntocorp/semgrep-action@v1
        with:
          config: auto
        env:
          SEMGREP_RULES: p/python

      - name: Upload SARIF
        if: always()
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: semgrep.sarif
          category: semgrep

  sca-python:
    name: SCA - pip-audit
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run pip-audit
        uses: pypa/gh-action-pip-audit@v1.1.0
        with:
          inputs: requirements.txt

  secrets:
    name: Secrets Detection - Gitleaks
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2.3.7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Key difference:** No container scanning stage (no Dockerfile detected). Semgrep uses `p/python` ruleset. SCA uses `pip-audit` instead of `npm audit`.
