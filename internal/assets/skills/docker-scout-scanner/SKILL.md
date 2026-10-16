---
name: docker-scout-scanner
description: "Trigger: Docker image CVEs, Dockerfile misconfiguration, base image choice, container SBOM, docker-compose. Audits container/OCI security."
when_to_use: "escanear imagen Docker, vulnerabilidades contenedor, auditar Docker, seguridad contenedor"
---

# Docker Scout Scanner

This skill performs container security analysis for Docker projects using Docker Scout, identifying CVEs in image layers, insecure Dockerfile patterns, outdated base images, and misconfigured container builds, then mapping findings to CWE and OWASP Top 10:2021 standards. When Docker Scout is unavailable, the skill falls back to a ten-point static Dockerfile review covering the most critical container hardening checks.

## When to Use

- When the user asks to "scan a Docker image for vulnerabilities" or "run Docker Scout"
- When the user mentions "container CVE scan", "image security", or "container SAST"
- When reviewing a `Dockerfile` or `docker-compose.yml` before deployment
- When a pull request contains changes to `Dockerfile`, `.dockerignore`, or `docker-compose.yml` and a security check is requested
- When the user wants to check for outdated base images, exposed secrets in build args, or privilege escalation risks
- When generating a Software Bill of Materials (SBOM) for a container image

## When NOT to Use

- When scanning application source code for logic vulnerabilities (use `bandit-sast`, `semgrep-rule-creator`, or `security-review` instead)
- When the user wants runtime monitoring or intrusion detection (use dedicated runtime security tools such as Falco)
- When the user is asking about Kubernetes manifests or Helm charts (use `iac-scanner` for IaC-level review)
- When the `security-review` skill already covers the request at a general level and no container-specific depth is needed
- When the target directory contains **no Dockerfiles or container configuration** — you **MUST** decline and recommend `bandit-sast` (Python), `security-review` (general), or the appropriate skill
- When the user wants to **generate a CI/CD pipeline** — you **MUST** decline and recommend `devsecops-pipeline`

## Prerequisites

### Tool Installed (Preferred)

```bash
# Detection
docker scout version

# Installation — Docker Scout is bundled with Docker Desktop 4.17+
# For CLI-only environments:
curl -fsSL https://raw.githubusercontent.com/docker/scout-cli/main/install.sh | sh

# Verify
docker scout version
```

Minimum version: Docker Scout CLI 1.0+. A Docker Hub account is required for full CVE database access (`docker login`). Local SBOM analysis works without authentication.

### Tool Not Installed (Fallback)

> **Note:** This is a limited review based on static Dockerfile analysis. Install Docker Scout (bundled with Docker Desktop) for comprehensive CVE scanning with full image layer inspection.

When Docker Scout is not available, perform these ten manual Dockerfile security checks:

1. **Running as root — missing USER directive** — Search for `Dockerfile` files that have no `USER` instruction before the final `CMD`/`ENTRYPOINT`. Containers default to root (UID 0) when no `USER` is set.
2. **Using `latest` tag for base images** — Search for `FROM` lines ending in `:latest` or containing no tag at all. Unpinned tags make builds non-deterministic and can silently pull vulnerable image versions.
3. **Exposed secrets in build args** — Search for `ARG` instructions with names containing `PASSWORD`, `SECRET`, `TOKEN`, `KEY`, or `CREDENTIAL`. Build args are visible in image history (`docker history --no-trunc`).
4. **Unnecessary packages** — Search for `apt-get install` or `apk add` lines that do not include `--no-install-recommends` (apt) or `--no-cache` (apk). Excess packages increase the attack surface.
5. **Missing HEALTHCHECK directive** — Search for `Dockerfile` files with no `HEALTHCHECK` instruction. Without health checks, orchestrators cannot detect degraded containers.
6. **Multi-stage build opportunities** — Identify single-stage builds that install build tools (e.g., `gcc`, `make`, `maven`, `npm`, `go`) in the same stage as the final runtime. These tools and intermediate artifacts should be confined to a builder stage.
7. **ADD with remote URLs or archives when COPY suffices** — Search for `ADD` instructions. `ADD` silently decompresses archives and can fetch remote URLs, which bypasses integrity checks. Prefer `COPY` for local files.
8. **Insecure or non-official base images** — Check `FROM` lines for images not sourced from Docker Official Images (`library/`), Docker Verified Publishers, or a trusted internal registry.
9. **Shell form CMD/ENTRYPOINT instead of exec form** — Search for `CMD` or `ENTRYPOINT` that use the shell form (bare string, not a JSON array). Shell form spawns a shell process (`/bin/sh -c`), which does not forward OS signals to the application, preventing graceful shutdown.
10. **Missing `COPY --chown` for application files** — Search for `COPY` instructions that transfer application code without `--chown=<user>:<group>`. Files copied without `--chown` are owned by root even when a non-root `USER` is set later.

## Workflow

> **MANDATORY FIRST ACTION — Verify the tool before reporting its output.**
>
> Your first Bash call must be `docker scout version`. Note that `docker` being on the PATH is not sufficient — Docker Scout is a separate plugin (bundled with Docker Desktop 4.17+, absent in CLI-only environments). Branch on the result:
>
> - **Docker Scout is available** — proceed with the installed-tool workflow (step 3a below). The report may use `## Docker Scout CVE Scan Results`, list CVE IDs, cite the Scout version, and include per-layer package/vulnerability counts, because real scanner output backs all of it.
> - **Docker Scout is not available** — proceed with the fallback workflow (step 3b below). The report must:
>   - Use header `## Dockerfile Security Review (Manual Fallback)`.
>   - Open with: `> Note: This is a limited review. Install Docker Scout for comprehensive CVE scanning.`
>   - Focus on Dockerfile static checks only. CVE IDs and per-layer counts require a real scan — without one, they would be invented data presented as measured data.
>   - Not claim a scanner, version, or image-layer metric that the turn history doesn't show.
>
> The contract is simple: **every artifact in the report must trace back to something the skill actually did in this turn.** If you did not run `docker scout cves`, don't present Scout results. The user is relying on the report matching what was actually scanned.

1. **Detect Docker project** — Confirm Docker artifacts exist by checking for `Dockerfile`, `docker-compose.yml`, `.dockerignore`, or image references in CI configuration files.
2. **Check for Docker Scout** — Run `docker scout version` to determine if Docker Scout is installed.
3. **If Docker Scout is installed:**
   a. Identify the image name and tag. If no image name is provided, look for a `docker build` tag in the project's Makefile, CI config, or `docker-compose.yml`.
   b. Run `docker scout cves <image> --format json` to retrieve a JSON list of CVEs across all image layers.
   c. Run `docker scout recommendations <image>` to get base image upgrade suggestions.
   d. Optionally run `docker scout sbom <image>` to list all packages included in the image.
   e. Parse CVE results — each entry contains `id`, `severity`, `cvss`, `package`, `version`, `fixed_version`, and `layer`.
   f. Map CVE categories to CWEs and OWASP categories using the Reference Tables below.
4. **If Docker Scout is NOT installed:**
   a. Offer to guide the user through installing Docker Scout (see Prerequisites).
   b. If the user declines or Docker is not available, run the ten manual fallback checks listed in Prerequisites against any `Dockerfile` in the project.
   c. Include the disclaimer: "This is a limited review. Install Docker Scout for comprehensive CVE scanning with layer-level detail."
5. **Compile findings** — Deduplicate results and sort by severity: Critical > High > Medium > Low.
6. **Generate report** — Present findings using the Findings Format below.
7. **Summarize** — State total findings, breakdown by severity, affected packages or Dockerfile lines, and top three remediation priorities.

## Findings Format

> **MANDATORY FORMAT:** You **MUST** include Severity, CWE, and OWASP Top 10:2021 mapping on **every** finding.

Each finding should include:

| Field | Description |
|-------|-------------|
| Severity | Critical / High / Medium / Low |
| CWE | CWE-XXX identifier |
| OWASP | A01-A10 category (OWASP Top 10:2021) |
| Location | Image layer or Dockerfile:line |
| Issue | Description of the vulnerability |
| Remediation | How to fix it |

### Example Finding

| Field | Value |
|-------|-------|
| Severity | Critical |
| CWE | CWE-798 |
| OWASP | A07:2021 - Security Misconfiguration |
| Location | Dockerfile:6 |
| Issue | `ARG DB_PASSWORD=hunter2` — hardcoded secret in build arg is visible in image history |
| Remediation | Remove the default value from `ARG`; pass secrets at build time with `--secret` (BuildKit) and never bake them into layers |

## Reference Tables

### Container Security Check to CWE Mapping

| Check | Description | CWE | OWASP | Severity |
|-------|-------------|-----|-------|----------|
| Running as root | No `USER` directive; container executes as UID 0 | CWE-250 | A01:2021 | High |
| `latest` tag on base image | Unpinned base image tag allows silent version drift | CWE-1104 | A06:2021 | Medium |
| Exposed secrets in build args | `ARG` with passwords/tokens stored in image history | CWE-798 | A07:2021 | Critical |
| No HEALTHCHECK | Orchestrator cannot detect degraded container state | CWE-693 | A07:2021 | Low |
| `ADD` with remote URL | Fetches remote content without integrity verification | CWE-829 | A08:2021 | Medium |
| Unnecessary packages installed | Excess packages broaden exploitable attack surface | CWE-250 | A01:2021 | Low |
| No multi-stage build | Build tools present in final image increase CVE exposure | CWE-1104 | A06:2021 | Low |
| Insecure or unofficial registry | Base image from untrusted source, no provenance | CWE-494 | A08:2021 | High |
| Shell form CMD/ENTRYPOINT | Shell wrapper prevents signal propagation; shell injection risk | CWE-693 | A07:2021 | Low |
| Outdated base image | Known CVEs in unpinned or stale base image | CWE-1104 | A06:2021 | High |
| Missing `COPY --chown` | Application files owned by root despite non-root USER | CWE-732 | A01:2021 | Medium |
| Exposed unnecessary ports | `EXPOSE` of ports not required by the application | CWE-284 | A01:2021 | Medium |

### OWASP Top 10:2021 Quick Reference

| Category | Description | Common CWEs |
|----------|-------------|-------------|
| A01:2021 | Broken Access Control | CWE-250, CWE-284, CWE-732 |
| A06:2021 | Vulnerable and Outdated Components | CWE-1104 |
| A07:2021 | Security Misconfiguration | CWE-693, CWE-798 |
| A08:2021 | Software and Data Integrity Failures | CWE-494, CWE-829 |

## Example Usage

### With Docker Scout Installed

**User prompt:**
> "Scan my Docker image for vulnerabilities"

**Expected output (abbreviated):**

```text
## Docker Scout CVE Scan Results

Image: myapp:1.4.2
Scanned 312 packages across 7 layers

### Findings (6 total: 1 Critical, 2 High, 2 Medium, 1 Low)

| # | Severity | CVE | CWE | OWASP | Location | Issue |
|---|----------|-----|-----|-------|----------|-------|
| 1 | Critical | CVE-2023-44487 | CWE-400 | A06 | layer: base/debian:11 | HTTP/2 Rapid Reset DoS in libnghttp2 < 1.57.0 |
| 2 | High | CVE-2023-4863 | CWE-787 | A06 | layer: node:18 | Heap buffer overflow in libwebp |
| 3 | High | CWE-798 | CWE-798 | A07 | Dockerfile:6 | ARG DB_PASSWORD has hardcoded default value |
| 4 | Medium | CVE-2023-29491 | CWE-787 | A06 | layer: base/ncurses | Stack buffer overflow in ncurses |
| 5 | Medium | CWE-1104 | CWE-1104 | A06 | Dockerfile:1 | FROM node:latest — unpinned base image |
| 6 | Low | CWE-693 | CWE-693 | A07 | Dockerfile:22 | No HEALTHCHECK directive |

### Base Image Recommendations (docker scout recommendations)
- Current: node:18.17.0 — 12 CVEs (2 Critical, 4 High)
- Recommended: node:20.10.0-alpine — 1 CVE (0 Critical, 0 High)

### Recommendations
1. Upgrade base image from node:18 to node:20.10.0-alpine to eliminate 11 CVEs (Finding #1, #2, #4)
2. Remove default value from ARG DB_PASSWORD and use BuildKit secrets (Finding #3)
3. Pin base image to a digest: FROM node:20.10.0-alpine@sha256:... (Finding #5)
```

### Without Docker Scout (Fallback Mode)

**User prompt:**
> "Check this Dockerfile for security issues"

> **Note:** This is a limited review based on static Dockerfile analysis. Install Docker Scout for comprehensive CVE scanning with layer-level detail.

**Expected output (abbreviated):**

```text
## Dockerfile Security Review (Manual Fallback)

> Note: This is a limited review. Install Docker Scout for comprehensive CVE scanning.

Analyzed: Dockerfile (28 lines)

### Findings (7 total: 1 Critical, 2 High, 2 Medium, 2 Low)

| # | Severity | CWE | OWASP | Location | Issue |
|---|----------|-----|-------|----------|-------|
| 1 | Critical | CWE-798 | A07 | Dockerfile:6 | ARG DB_PASSWORD=hunter2 — hardcoded secret in build arg |
| 2 | High | CWE-250 | A01 | Dockerfile:— | No USER directive; container runs as root |
| 3 | High | CWE-494 | A08 | Dockerfile:1 | Base image from unofficial registry: myregistry.local/node |
| 4 | Medium | CWE-1104 | A06 | Dockerfile:1 | FROM node:latest — unpinned base image tag |
| 5 | Medium | CWE-829 | A08 | Dockerfile:12 | ADD http://example.com/config.tar.gz /app/ — remote fetch without integrity check |
| 6 | Low | CWE-693 | A07 | Dockerfile:— | No HEALTHCHECK directive |
| 7 | Low | CWE-693 | A07 | Dockerfile:26 | CMD uses shell form — signals not forwarded to application |

### Recommendations
1. Remove hardcoded default from ARG; inject secrets at build time with BuildKit --secret (Finding #1)
2. Add a non-root USER directive before CMD (e.g., USER 1001) (Finding #2)
3. Pin base image to a specific digest or immutable tag (Finding #4)
```

### UNSAFE vs SAFE Dockerfile Patterns

**Running as root (CWE-250)**

```dockerfile
# UNSAFE — no USER directive
FROM node:20-alpine
WORKDIR /app
COPY . .
CMD ["node", "server.js"]
```

```dockerfile
# SAFE — explicit non-root user
FROM node:20-alpine
WORKDIR /app
COPY --chown=node:node . .
USER node
CMD ["node", "server.js"]
```

**Exposed secrets in build args (CWE-798)**

```dockerfile
# UNSAFE — default value baked into image history
ARG DB_PASSWORD=supersecret
RUN echo "password=$DB_PASSWORD" > /app/.env
```

```dockerfile
# SAFE — BuildKit secret mount; never stored in a layer
RUN --mount=type=secret,id=db_password \
    DB_PASSWORD=$(cat /run/secrets/db_password) && \
    echo "password=$DB_PASSWORD" > /app/.env
```

**Unpinned latest tag (CWE-1104)**

```dockerfile
# UNSAFE — silent version drift
FROM python:latest
```

```dockerfile
# SAFE — pinned to digest for reproducibility
FROM python:3.12.3-slim@sha256:a5d2e6d4e5f3b1c8a9f0e7d2c4b6a8e1d3f5c7b9a2e4d6f8c0b2a4e6d8f0c2b4
```

**Shell form vs exec form CMD (CWE-693)**

```dockerfile
# UNSAFE — shell form; app does not receive SIGTERM
CMD node server.js
```

```dockerfile
# SAFE — exec form; SIGTERM delivered directly to process
CMD ["node", "server.js"]
```

**ADD with remote URL (CWE-829)**

```dockerfile
# UNSAFE — remote fetch with no hash verification
ADD https://example.com/setup.sh /tmp/setup.sh
RUN bash /tmp/setup.sh
```

```dockerfile
# SAFE — download with checksum verification, then COPY
RUN curl -fsSL https://example.com/setup.sh -o /tmp/setup.sh && \
    echo "expectedsha256hash  /tmp/setup.sh" | sha256sum -c -
```
