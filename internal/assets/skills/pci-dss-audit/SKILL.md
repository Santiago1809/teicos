---
name: pci-dss-audit
description: "Trigger: PCI-DSS v4.0, cardholder data, credit-card storage and transmission, PAN logging. Audits code for PCI compliance."
when_to_use: "cumplimiento PCI, auditar pagos, PCI-DSS, datos de tarjetas, compliance financiero"
---

# PCI-DSS Audit

This skill performs static code analysis for PCI-DSS v4.0 compliance violations across JavaScript/TypeScript, Python, Java, Go, and C#/.NET projects. It identifies 12 common PCI-DSS code-level anti-patterns — unprotected cardholder data, weak encryption of PANs, missing audit trails, insufficient access controls, and more — mapping each finding to CWE and PCI-DSS v4.0 requirement numbers with concrete UNSAFE/SAFE code pairs for remediation.

**Scope:** This skill covers application-code-level controls only. It does NOT audit infrastructure, network segmentation, physical security, or organizational policies — those require separate assessment tools and processes.

## When to Use

- When the user asks to "audit PCI compliance", "check PCI-DSS", or "review payment card handling"
- When the user mentions "PCI audit", "cardholder data", "PAN protection", or "payment security"
- When scanning code that handles credit card numbers, CVVs, expiration dates, or payment tokens
- When reviewing code that stores, processes, or transmits cardholder data
- When a pull request modifies payment processing, card storage, or checkout flows
- When the user asks about "card data in logs", "PAN masking", "payment encryption", or "audit logging"
- When preparing for a PCI-DSS v4.0 Self-Assessment Questionnaire (SAQ) or Report on Compliance (RoC)

## When NOT to Use

- When the user is asking about network segmentation, firewall rules, or physical security (PCI-DSS Req 1, 9)
- When the user wants to audit infrastructure configurations (use `iac-scanner`)
- When the user needs general cryptographic review not related to payment data (use `crypto-audit`)
- When reviewing code that does not handle cardholder data or payment flows
- When the user needs a runtime scan of a live payment environment (use a DAST or ASV tool)
- When the user asks about **Dockerfiles, container security, or network segmentation** — you **MUST** decline and recommend `docker-scout-scanner` or `iac-scanner`
- When the user asks about **general cryptographic review** not related to payment data — you **MUST** decline and recommend `crypto-audit`

## Prerequisites

### Tool Installed (Preferred)

No external tool required. This skill uses code analysis only.

All 12 checks are performed through pattern matching and code inspection — no CLI tool needs to be installed, configured, or invoked.

### Tool Not Installed (Fallback)

This skill is always available as a pure analysis skill. There is no fallback mode because there is no external tool dependency. All checks run directly through code analysis.

## Workflow

> **MANDATORY FIRST ACTION — Verify PCI-DSS scope before activating.**
>
> PCI-DSS applies only to code that handles cardholder data. Before starting the audit, grep the target for at least one of these signals:
>
> - File or directory names containing `payment`, `checkout`, `card`, `billing`, `transaction`, `stripe`, `braintree`, `adyen`, `square`, `paypal`, `merchant`, or `pci`
> - Imports of a payment SDK (`stripe`, `braintree`, `adyen`, `paypal`, `square`)
> - References to `cardNumber`, `pan`, `cvv`, `cvc`, `expiry`, `credit_card`, `card_number`, `payment_method`, `invoice`
> - Test fixtures explicitly scoped to PCI-DSS or cardholder data
>
> If none of these appear in the target, the target is outside PCI-DSS scope. Decline and redirect:
>
> > *"This code does not appear to handle cardholder data or payment flows — PCI-DSS does not apply. For general cryptographic review, use `crypto-audit`. For generic SAST, use `bandit-sast` or `security-review`."*
>
> A file of general cryptographic utilities (key management, hashing, session generation) is not in PCI-DSS scope just because it uses crypto primitives — that is the `crypto-audit` skill's domain. PCI-DSS findings require *cardholder data* to actually be present in the code. Framing generic crypto code as a "payment gateway SDK" to justify activation produces findings that do not describe what the code actually does and misleads the user about their compliance posture.

1. **Detect project languages** — Inspect project files to determine which languages are in use: `package.json` or `*.ts`/`*.js` (JavaScript/TypeScript), `requirements.txt`/`*.py` (Python), `pom.xml`/`*.java` (Java), `go.mod`/`*.go` (Go), `*.csproj`/`*.cs` (C#/.NET).
2. **Identify payment-relevant files** — Search for files that reference cardholder data or payment processing:
   - Filenames containing: `payment`, `checkout`, `card`, `billing`, `transaction`, `stripe`, `braintree`, `adyen`
   - Code importing payment SDKs: `stripe`, `braintree`, `adyen`, `paypal`, `square`
   - Code referencing card patterns: `cardNumber`, `pan`, `cvv`, `cvc`, `expiry`, `credit_card`, `card_number`
   - Code referencing encryption/tokenization: `encrypt`, `tokenize`, `vault`, `mask`
3. **Run the 12 PCI-DSS code checks** against each identified file (see Checks section below).
4. **For each finding:**
   a. Determine severity (Critical / High / Medium / Low) using the Reference Tables
   b. Map to the relevant CWE identifier
   c. Map to the relevant PCI-DSS v4.0 requirement number
   d. Record file path and line number
   e. Generate the UNSAFE pattern found and the corresponding SAFE fix
   f. Draft a remediation recommendation
5. **Deduplicate and sort** findings by severity: Critical > High > Medium > Low.
6. **Generate the findings report** using the Findings Format below.
7. **Summarize** — State total findings, breakdown by severity, and top 3 remediation priorities.

## Checks

### Check 1: PAN Logged in Plaintext

**CWE-532** (Insertion of Sensitive Information into Log File) | **PCI-DSS Req 3.4, 10.3** | Severity: **Critical**

**WHY:** PCI-DSS Req 3.4 requires PANs to be rendered unreadable anywhere they are stored — including log files. Logging full card numbers creates an uncontrolled copy of cardholder data that persists in log aggregators, monitoring systems, and backups. Attackers who gain access to logs obtain card numbers without needing to breach the payment database.

**UNSAFE:**

```javascript
// JavaScript — full PAN in console log
function processPayment(cardNumber, amount) {
  console.log('Processing payment for card: ' + cardNumber);
  console.log(`Transaction: card=${cardNumber}, amount=${amount}`);
}
```

```python
# Python — card number in logging output
import logging
logger = logging.getLogger(__name__)

def charge_card(card_number, amount):
    logger.info(f"Charging card {card_number} for ${amount}")
```

```java
// Java — PAN in log statement
import org.slf4j.Logger;
public void processPayment(String cardNumber, double amount) {
    logger.info("Processing card: " + cardNumber);
}
```

**SAFE:**

```javascript
// Mask PAN to show only last 4 digits
function maskPan(cardNumber) {
  return '****-****-****-' + cardNumber.slice(-4);
}
function processPayment(cardNumber, amount) {
  console.log('Processing payment for card: ' + maskPan(cardNumber));
}
```

```python
# Mask PAN before logging
def mask_pan(card_number: str) -> str:
    return f"****-****-****-{card_number[-4:]}"

def charge_card(card_number, amount):
    logger.info(f"Charging card {mask_pan(card_number)} for ${amount}")
```

---

### Check 2: Cardholder Data in URL Parameters

**CWE-598** (Use of GET Request Method With Sensitive Query Strings) | **PCI-DSS Req 4.2** | Severity: **Critical**

**WHY:** Card data in URLs is recorded in browser history, web server access logs, proxy logs, and referrer headers. PCI-DSS Req 4.2 prohibits sending unprotected PANs via end-user messaging technologies, and URL parameters are inherently logged and cached by infrastructure outside your control.

**UNSAFE:**

```javascript
// JavaScript — card data in query string
const url = `/api/payment?cardNumber=${cardNumber}&cvv=${cvv}&expiry=${expiry}`;
fetch(url);
```

```python
# Python — card data as GET parameters
import requests
response = requests.get(
    f"https://api.example.com/charge?card={card_number}&cvv={cvv}"
)
```

```java
// Java — card in URL path
String url = "https://api.example.com/pay?pan=" + cardNumber + "&cvv=" + cvv;
HttpURLConnection conn = (HttpURLConnection) new URL(url).openConnection();
```

**SAFE:**

```javascript
// Send card data in POST body over HTTPS
const response = await fetch('/api/payment', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ cardToken: tokenizedCard, amount }),
});
```

```python
# Use POST with tokenized card data
response = requests.post(
    "https://api.example.com/charge",
    json={"card_token": token, "amount": amount},
)
```

---

### Check 3: Weak Encryption of Stored Card Data

**CWE-327** (Use of a Broken or Risky Cryptographic Algorithm) | **PCI-DSS Req 3.5** | Severity: **Critical**

**WHY:** PCI-DSS Req 3.5 requires strong cryptography to protect stored PANs. MD5, Base64 encoding, SHA1, simple XOR, or reversible encoding are not encryption — they provide no meaningful protection. Attackers can trivially reverse Base64 or crack MD5/SHA1 hashes of 16-digit card numbers.

**UNSAFE:**

```javascript
// JavaScript — Base64 "encryption" of card number
function storeCard(cardNumber) {
  const encoded = Buffer.from(cardNumber).toString('base64');
  db.save({ encryptedCard: encoded });
}
```

```python
# Python — MD5 hash of card number (reversible via rainbow tables for 16-digit space)
import hashlib
card_hash = hashlib.md5(card_number.encode()).hexdigest()
```

```java
// Java — XOR "encryption"
byte[] encrypted = new byte[cardNumber.length()];
for (int i = 0; i < cardNumber.length(); i++) {
    encrypted[i] = (byte) (cardNumber.charAt(i) ^ 0x5A);
}
```

**SAFE:**

```javascript
// Use AES-256-GCM with proper key management
const crypto = require('crypto');
function encryptPan(pan) {
  const key = Buffer.from(process.env.PAN_ENCRYPTION_KEY, 'hex');
  const iv = crypto.randomBytes(12);
  const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
  const encrypted = Buffer.concat([cipher.update(pan, 'utf8'), cipher.final()]);
  const tag = cipher.getAuthTag();
  return { iv: iv.toString('hex'), data: encrypted.toString('hex'), tag: tag.toString('hex') };
}
```

```python
# Use AES-256-GCM or a tokenization service
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import os
key = os.environ['PAN_ENCRYPTION_KEY'].encode()
aesgcm = AESGCM(key)
nonce = os.urandom(12)
encrypted = aesgcm.encrypt(nonce, pan.encode(), None)
```

---

### Check 4: Missing Audit Trail for Payment Operations

**CWE-778** (Insufficient Logging) | **PCI-DSS Req 10.2** | Severity: **High**

**WHY:** PCI-DSS Req 10.2 requires logging of all access to cardholder data, all actions by privileged users, and all access to audit trails. Payment operations without audit logging make it impossible to detect fraud, investigate breaches, or demonstrate compliance during assessments.

**UNSAFE:**

```javascript
// JavaScript — no logging of payment operation
async function chargeCard(card, amount) {
  const result = await paymentGateway.charge(card, amount);
  return result;
}
```

```python
# Python — payment with no audit trail
def process_refund(transaction_id, amount):
    gateway.refund(transaction_id, amount)
    return {"status": "refunded"}
```

```java
// Java — no logging of who performed the action or what happened
public PaymentResult processPayment(PaymentRequest request) {
    return gateway.charge(request.getAmount(), request.getToken());
}
```

**SAFE:**

```javascript
// Log payment operations with who, what, when (without cardholder data)
async function chargeCard(card, amount, userId) {
  const auditLogger = require('./audit-logger');
  auditLogger.log({
    action: 'PAYMENT_CHARGE',
    userId,
    amount,
    cardLastFour: card.number.slice(-4),
    timestamp: new Date().toISOString(),
    sourceIp: req.ip,
  });
  const result = await paymentGateway.charge(card, amount);
  auditLogger.log({
    action: 'PAYMENT_CHARGE_RESULT',
    userId,
    success: result.success,
    transactionId: result.id,
    timestamp: new Date().toISOString(),
  });
  return result;
}
```

```python
# Structured audit logging for payment operations
import structlog
audit_log = structlog.get_logger("audit")

def process_refund(transaction_id, amount, user_id):
    audit_log.info("payment.refund.initiated",
        transaction_id=transaction_id,
        amount=amount,
        user_id=user_id)
    result = gateway.refund(transaction_id, amount)
    audit_log.info("payment.refund.completed",
        transaction_id=transaction_id,
        success=result.success,
        user_id=user_id)
    return result
```

---

### Check 5: Hardcoded Payment Gateway Credentials

**CWE-798** (Use of Hard-coded Credentials) | **PCI-DSS Req 2.2, 8.6** | Severity: **Critical**

**WHY:** PCI-DSS Req 2.2 prohibits using vendor-supplied defaults and insecure configurations. Req 8.6 requires secure management of system and application accounts. Hardcoded API keys, merchant IDs, and gateway passwords in source code are extractable from repositories, build artifacts, and binaries — enabling unauthorized charges, refunds, and data exfiltration.

**UNSAFE:**

```javascript
// JavaScript — hardcoded Stripe secret key
const stripe = require('stripe')('HARDCODED_STRIPE_KEY_EXAMPLE');
```

```python
# Python — hardcoded gateway credentials
PAYMENT_API_KEY = "HARDCODED_PAYMENT_KEY_EXAMPLE"
MERCHANT_SECRET = "msk_prod_xyzzy987"
gateway = PaymentGateway(api_key=PAYMENT_API_KEY, secret=MERCHANT_SECRET)
```

```java
// Java — hardcoded credentials
private static final String API_KEY = "HARDCODED_STRIPE_KEY_EXAMPLE";
private static final String MERCHANT_ID = "merchant_prod_12345";
```

```csharp
// C# — hardcoded gateway credentials
var gateway = new BraintreeGateway {
    Environment = Braintree.Environment.PRODUCTION,
    MerchantId = "prod_merchant_123",
    PublicKey = "public_key_abc",
    PrivateKey = "private_key_xyz_secret"
};
```

**SAFE:**

```javascript
// Load credentials from environment variables or secrets manager
const stripe = require('stripe')(process.env.STRIPE_SECRET_KEY);
```

```python
# Load from environment or secrets manager
import os
gateway = PaymentGateway(
    api_key=os.environ["PAYMENT_API_KEY"],
    secret=os.environ["MERCHANT_SECRET"],
)
```

```java
// Load from environment or vault
String apiKey = System.getenv("PAYMENT_API_KEY");
String merchantId = System.getenv("MERCHANT_ID");
```

---

### Check 6: No Input Validation on Card Numbers

**CWE-20** (Improper Input Validation) | **PCI-DSS Req 6.2** | Severity: **High**

**WHY:** PCI-DSS Req 6.2 requires secure development practices including input validation. Accepting unvalidated card input enables injection attacks, application errors, and processing of malformed data. Luhn check failures waste gateway API calls and may trigger rate limiting or account flags.

**UNSAFE:**

```javascript
// JavaScript — no validation before processing
function processPayment(cardNumber, cvv, expiry, amount) {
  return gateway.charge({ cardNumber, cvv, expiry, amount });
}
```

```python
# Python — card number passed directly to gateway
def charge(card_number, cvv, amount):
    return gateway.charge(card_number=card_number, cvv=cvv, amount=amount)
```

```java
// Java — no validation
public void charge(String cardNumber, String cvv, double amount) {
    gateway.process(cardNumber, cvv, amount);
}
```

**SAFE:**

```javascript
// Validate card number format and Luhn check before processing
function luhnCheck(num) {
  let sum = 0;
  let alternate = false;
  for (let i = num.length - 1; i >= 0; i--) {
    let n = parseInt(num[i], 10);
    if (alternate) { n *= 2; if (n > 9) n -= 9; }
    sum += n;
    alternate = !alternate;
  }
  return sum % 10 === 0;
}

function processPayment(cardNumber, cvv, expiry, amount) {
  const sanitized = cardNumber.replace(/[\s-]/g, '');
  if (!/^\d{13,19}$/.test(sanitized)) throw new Error('Invalid card number format');
  if (!luhnCheck(sanitized)) throw new Error('Card number fails Luhn check');
  if (!/^\d{3,4}$/.test(cvv)) throw new Error('Invalid CVV format');
  if (typeof amount !== 'number' || amount <= 0) throw new Error('Invalid amount');
  return gateway.charge({ cardNumber: sanitized, cvv, expiry, amount });
}
```

```python
# Validate with Luhn algorithm
def luhn_check(card_number: str) -> bool:
    digits = [int(d) for d in card_number]
    odd_digits = digits[-1::-2]
    even_digits = digits[-2::-2]
    total = sum(odd_digits) + sum(d * 2 - 9 if d * 2 > 9 else d * 2 for d in even_digits)
    return total % 10 == 0

def charge(card_number: str, cvv: str, amount: float):
    sanitized = card_number.replace(" ", "").replace("-", "")
    if not sanitized.isdigit() or not (13 <= len(sanitized) <= 19):
        raise ValueError("Invalid card number format")
    if not luhn_check(sanitized):
        raise ValueError("Card number fails Luhn check")
    if not cvv.isdigit() or len(cvv) not in (3, 4):
        raise ValueError("Invalid CVV")
    return gateway.charge(card_number=sanitized, cvv=cvv, amount=amount)
```

---

### Check 7: Cardholder Data Stored Without Encryption

**CWE-312** (Cleartext Storage of Sensitive Information) | **PCI-DSS Req 3.4** | Severity: **Critical**

**WHY:** PCI-DSS Req 3.4 mandates that PANs are rendered unreadable anywhere they are stored. Storing card numbers in plaintext in databases, files, or caches means a single data breach exposes all cardholder data. This is the most fundamental PCI-DSS requirement for stored data.

**UNSAFE:**

```javascript
// JavaScript — plaintext card data in database
async function saveCard(userId, cardNumber, expiry, cvv) {
  await db.query(
    'INSERT INTO cards (user_id, card_number, expiry, cvv) VALUES (?, ?, ?, ?)',
    [userId, cardNumber, expiry, cvv]
  );
}
```

```python
# Python — plaintext storage
def save_card(user_id, card_number, expiry):
    db.execute(
        "INSERT INTO cards (user_id, card_number, expiry) VALUES (%s, %s, %s)",
        (user_id, card_number, expiry)
    )
```

```java
// Java — plaintext card in entity
@Entity
public class CreditCard {
    @Column(name = "card_number")
    private String cardNumber;  // Stored in cleartext

    @Column(name = "cvv")
    private String cvv;  // CVV must NEVER be stored post-authorization
}
```

**SAFE:**

```javascript
// Use tokenization — never store raw PANs
async function saveCard(userId, cardNumber) {
  const token = await paymentProvider.tokenize(cardNumber);
  await db.query(
    'INSERT INTO cards (user_id, card_token, last_four) VALUES (?, ?, ?)',
    [userId, token, cardNumber.slice(-4)]
  );
  // NEVER store CVV post-authorization (PCI-DSS Req 3.3.2)
}
```

```python
# Tokenize via payment provider
def save_card(user_id, card_number):
    token = payment_provider.tokenize(card_number)
    db.execute(
        "INSERT INTO cards (user_id, card_token, last_four) VALUES (%s, %s, %s)",
        (user_id, token, card_number[-4:])
    )
```

---

### Check 8: Missing TLS for Cardholder Data Transmission

**CWE-319** (Cleartext Transmission of Sensitive Information) | **PCI-DSS Req 4.2** | Severity: **Critical**

**WHY:** PCI-DSS Req 4.2 requires strong cryptography for transmission of cardholder data over open, public networks. Transmitting card data over HTTP, unencrypted sockets, or with disabled TLS verification allows network-level interception of PANs and credentials.

**UNSAFE:**

```javascript
// JavaScript — HTTP for payment API
const response = await fetch('http://payment-api.example.com/charge', {
  method: 'POST',
  body: JSON.stringify({ cardNumber, amount }),
});
```

```python
# Python — HTTP endpoint for card data
import requests
response = requests.post(
    "http://api.example.com/payments",
    json={"card_number": card_number, "amount": amount}
)
```

```go
// Go — disabled TLS verification for payment calls
import "crypto/tls"
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    },
}
resp, _ := client.Post("https://payment-api.example.com/charge", "application/json", body)
```

**SAFE:**

```javascript
// Always use HTTPS with certificate validation
const https = require('https');
const response = await fetch('https://payment-api.example.com/charge', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ cardToken: token, amount }),
});
```

```python
# HTTPS with default certificate verification
response = requests.post(
    "https://api.example.com/payments",
    json={"card_token": token, "amount": amount},
    # verify=True is default — never set verify=False for payment endpoints
)
```

```go
// Use default TLS config (validates certificates)
client := &http.Client{}
resp, err := client.Post("https://payment-api.example.com/charge", "application/json", body)
```

---

### Check 9: Missing Access Control on Payment Endpoints

**CWE-862** (Missing Authorization) | **PCI-DSS Req 7.2** | Severity: **High**

**WHY:** PCI-DSS Req 7.2 requires access control systems that restrict access based on a user's need to know and are set to deny all unless specifically allowed. Payment endpoints without authentication and authorization checks allow unauthorized users to initiate charges, view transaction data, or process refunds.

**UNSAFE:**

```javascript
// JavaScript — no auth middleware on payment route
app.post('/api/payments/charge', async (req, res) => {
  const result = await gateway.charge(req.body.amount, req.body.cardToken);
  res.json(result);
});

app.post('/api/payments/refund', async (req, res) => {
  const result = await gateway.refund(req.body.transactionId, req.body.amount);
  res.json(result);
});
```

```python
# Python Flask — no authentication
@app.route('/api/payments/charge', methods=['POST'])
def charge():
    data = request.get_json()
    result = gateway.charge(data['amount'], data['card_token'])
    return jsonify(result)
```

```java
// Java Spring — no security annotation
@RestController
public class PaymentController {
    @PostMapping("/api/payments/charge")
    public ResponseEntity<PaymentResult> charge(@RequestBody PaymentRequest request) {
        return ResponseEntity.ok(paymentService.charge(request));
    }
}
```

**SAFE:**

```javascript
// Require authentication and authorization
const { authenticate, authorize } = require('./middleware/auth');

app.post('/api/payments/charge',
  authenticate,
  authorize('payments:charge'),
  async (req, res) => {
    const result = await gateway.charge(req.body.amount, req.body.cardToken);
    res.json(result);
  }
);
```

```python
# Flask with authentication required
from functools import wraps

@app.route('/api/payments/charge', methods=['POST'])
@login_required
@require_permission('payments:charge')
def charge():
    data = request.get_json()
    result = gateway.charge(data['amount'], data['card_token'])
    return jsonify(result)
```

```java
// Spring Security with role-based access
@RestController
@PreAuthorize("hasRole('PAYMENT_PROCESSOR')")
public class PaymentController {
    @PostMapping("/api/payments/charge")
    @PreAuthorize("hasAuthority('PAYMENT_CHARGE')")
    public ResponseEntity<PaymentResult> charge(@RequestBody PaymentRequest request) {
        return ResponseEntity.ok(paymentService.charge(request));
    }
}
```

---

### Check 10: Card Numbers in Error Messages

**CWE-209** (Generation of Error Message Containing Sensitive Information) | **PCI-DSS Req 3.4, 6.2** | Severity: **High**

**WHY:** Error messages containing card numbers leak cardholder data to end users, client-side logs, error tracking services (Sentry, Datadog), and browser consoles. PCI-DSS requires PANs to be masked everywhere they appear, and Req 6.2 requires secure error handling that does not expose sensitive data.

**UNSAFE:**

```javascript
// JavaScript — card number in error response
function validateCard(cardNumber) {
  if (!isValidCard(cardNumber)) {
    throw new Error(`Invalid card number: ${cardNumber}`);
  }
}
```

```python
# Python — card data in exception
def process_payment(card_number, amount):
    try:
        gateway.charge(card_number, amount)
    except GatewayError as e:
        raise PaymentError(f"Failed to charge card {card_number}: {e}")
```

```java
// Java — card number in error message
public void validate(String cardNumber) {
    if (!isValid(cardNumber)) {
        throw new IllegalArgumentException("Invalid card: " + cardNumber);
    }
}
```

**SAFE:**

```javascript
// Never include card data in errors — use masked or reference IDs
function validateCard(cardNumber) {
  if (!isValidCard(cardNumber)) {
    throw new Error(`Invalid card number ending in ${cardNumber.slice(-4)}`);
  }
}
```

```python
# Mask card data in exceptions
def process_payment(card_number, amount):
    try:
        gateway.charge(card_number, amount)
    except GatewayError as e:
        masked = f"****{card_number[-4:]}"
        raise PaymentError(f"Failed to charge card {masked}: {e}")
```

---

### Check 11: CVV/CVC Stored Post-Authorization

**CWE-257** (Storing Passwords in a Recoverable Format) | **PCI-DSS Req 3.3.2** | Severity: **Critical**

**WHY:** PCI-DSS Req 3.3.2 explicitly prohibits storing the card verification code (CVV/CVC) after authorization. There is no legitimate reason to retain this value. Storing it — even encrypted — is a direct PCI-DSS violation that can result in immediate non-compliance and significant fines.

**UNSAFE:**

```javascript
// JavaScript — saving CVV to database
async function savePaymentMethod(userId, cardNumber, cvv, expiry) {
  await db.query(
    'INSERT INTO payment_methods (user_id, card_number, cvv, expiry) VALUES (?, ?, ?, ?)',
    [userId, cardNumber, cvv, expiry]
  );
}
```

```python
# Python — CVV persisted in model
class PaymentMethod(db.Model):
    card_number = db.Column(db.String(19))
    cvv = db.Column(db.String(4))  # NEVER store CVV
    expiry = db.Column(db.String(7))
```

```java
// Java — CVV in entity field
@Entity
public class PaymentMethod {
    @Column(name = "cvv")
    private String cvv;  // PCI-DSS violation — CVV must never be stored
}
```

**SAFE:**

```javascript
// Use CVV only during authorization, never persist it
async function authorizePayment(cardNumber, cvv, expiry, amount) {
  // CVV used only for this authorization call
  const authResult = await gateway.authorize({ cardNumber, cvv, expiry, amount });

  // Store only the token — NEVER the CVV
  await db.query(
    'INSERT INTO payment_methods (user_id, card_token, last_four) VALUES (?, ?, ?)',
    [userId, authResult.token, cardNumber.slice(-4)]
  );
  return authResult;
}
```

```python
# CVV used in-memory only, never persisted
class PaymentMethod(db.Model):
    card_token = db.Column(db.String(64))   # Tokenized reference
    last_four = db.Column(db.String(4))      # Display purposes only
    # NO cvv column — CVV is never stored
```

---

### Check 12: Weak Session Management for Payment Flows

**CWE-384** (Session Fixation) | **PCI-DSS Req 8.2, 8.3** | Severity: **High**

**WHY:** PCI-DSS Req 8.2 and 8.3 require strong authentication and session management for access to cardholder data. Weak session handling — predictable tokens, no session regeneration after authentication, excessive session lifetimes, or missing secure cookie flags — allows session hijacking and unauthorized access to payment operations.

**UNSAFE:**

```javascript
// JavaScript — weak session configuration
app.use(session({
  secret: 'keyboard cat',
  cookie: {
    secure: false,      // Transmitted over HTTP
    httpOnly: false,     // Accessible to JavaScript
    maxAge: 86400000 * 30,  // 30-day session
  },
}));
```

```python
# Python Flask — insecure session settings
app.config['SECRET_KEY'] = 'simple-secret'
app.config['SESSION_COOKIE_SECURE'] = False
app.config['SESSION_COOKIE_HTTPONLY'] = False
app.config['PERMANENT_SESSION_LIFETIME'] = timedelta(days=30)
```

```java
// Java — no session regeneration after login
@PostMapping("/login")
public String login(HttpServletRequest request, String username, String password) {
    if (authService.authenticate(username, password)) {
        // Session fixation — same session ID used before and after auth
        request.getSession().setAttribute("user", username);
        return "redirect:/payments";
    }
    return "login";
}
```

**SAFE:**

```javascript
// Secure session configuration
app.use(session({
  secret: process.env.SESSION_SECRET,
  cookie: {
    secure: true,       // HTTPS only
    httpOnly: true,     // Not accessible to JavaScript
    sameSite: 'strict', // CSRF protection
    maxAge: 900000,     // 15-minute session for payment flows
  },
  resave: false,
  saveUninitialized: false,
}));
```

```python
# Secure session settings
app.config['SECRET_KEY'] = os.environ['SESSION_SECRET']
app.config['SESSION_COOKIE_SECURE'] = True
app.config['SESSION_COOKIE_HTTPONLY'] = True
app.config['SESSION_COOKIE_SAMESITE'] = 'Strict'
app.config['PERMANENT_SESSION_LIFETIME'] = timedelta(minutes=15)
```

```java
// Regenerate session after authentication
@PostMapping("/login")
public String login(HttpServletRequest request, String username, String password) {
    if (authService.authenticate(username, password)) {
        request.getSession().invalidate();  // Destroy old session
        HttpSession newSession = request.getSession(true);  // Create new one
        newSession.setAttribute("user", username);
        newSession.setMaxInactiveInterval(900);  // 15 minutes
        return "redirect:/payments";
    }
    return "login";
}
```

---

## Findings Format

Each finding should include:

| Field | Description |
|-------|-------------|
| Severity | Critical / High / Medium / Low |
| CWE | CWE-XXX identifier |
| PCI-DSS Req | PCI-DSS v4.0 requirement number |
| Location | file:line |
| Issue | Description of the violation |
| Remediation | How to fix it |

### Example Finding

| Field | Value |
|-------|-------|
| Severity | Critical |
| CWE | CWE-532 |
| PCI-DSS Req | 3.4, 10.3 |
| Location | src/payment/handler.js:23 |
| Issue | Full PAN logged in plaintext via `console.log('Card: ' + cardNumber)` |
| Remediation | Mask PAN to last 4 digits before logging; implement structured audit logging that excludes cardholder data |

## Reference Tables

### PCI-DSS Check to CWE Mapping

| # | Check | CWE | PCI-DSS Req | Default Severity |
|---|-------|-----|-------------|------------------|
| 1 | PAN logged in plaintext | CWE-532 | 3.4, 10.3 | Critical |
| 2 | Cardholder data in URL parameters | CWE-598 | 4.2 | Critical |
| 3 | Weak encryption of stored card data | CWE-327 | 3.5 | Critical |
| 4 | Missing audit trail for payment ops | CWE-778 | 10.2 | High |
| 5 | Hardcoded payment gateway credentials | CWE-798 | 2.2, 8.6 | Critical |
| 6 | No input validation on card numbers | CWE-20 | 6.2 | High |
| 7 | Cardholder data stored unencrypted | CWE-312 | 3.4 | Critical |
| 8 | Missing TLS for card data transmission | CWE-319 | 4.2 | Critical |
| 9 | Missing access control on payment endpoints | CWE-862 | 7.2 | High |
| 10 | Card numbers in error messages | CWE-209 | 3.4, 6.2 | High |
| 11 | CVV/CVC stored post-authorization | CWE-257 | 3.3.2 | Critical |
| 12 | Weak session management for payment flows | CWE-384 | 8.2, 8.3 | High |

### PCI-DSS v4.0 Requirements Covered

| Requirement | Description | Related Checks |
|-------------|-------------|----------------|
| Req 2.2 | Secure system configurations | Check 5 |
| Req 3.3.2 | Do not store CVV after authorization | Check 11 |
| Req 3.4 | Render PAN unreadable anywhere stored | Checks 1, 3, 7, 10 |
| Req 3.5 | Protect stored PAN with strong crypto | Check 3 |
| Req 4.2 | Protect CHD during transmission | Checks 2, 8 |
| Req 6.2 | Secure development practices | Checks 6, 10 |
| Req 7.2 | Restrict access by need to know | Check 9 |
| Req 8.2, 8.3 | Strong authentication and session management | Checks 5, 12 |
| Req 8.6 | Secure application/system account management | Check 5 |
| Req 10.2 | Audit logging of CHD access | Check 4 |
| Req 10.3 | Protect audit trail from modification | Check 1 |

### CWE Reference

| CWE ID | Name | MITRE URL |
|--------|------|-----------|
| CWE-20 | Improper Input Validation | https://cwe.mitre.org/data/definitions/20.html |
| CWE-209 | Generation of Error Message Containing Sensitive Information | https://cwe.mitre.org/data/definitions/209.html |
| CWE-257 | Storing Passwords in a Recoverable Format | https://cwe.mitre.org/data/definitions/257.html |
| CWE-312 | Cleartext Storage of Sensitive Information | https://cwe.mitre.org/data/definitions/312.html |
| CWE-319 | Cleartext Transmission of Sensitive Information | https://cwe.mitre.org/data/definitions/319.html |
| CWE-327 | Use of a Broken or Risky Cryptographic Algorithm | https://cwe.mitre.org/data/definitions/327.html |
| CWE-384 | Session Fixation | https://cwe.mitre.org/data/definitions/384.html |
| CWE-532 | Insertion of Sensitive Information into Log File | https://cwe.mitre.org/data/definitions/532.html |
| CWE-598 | Use of GET Request Method With Sensitive Query Strings | https://cwe.mitre.org/data/definitions/598.html |
| CWE-778 | Insufficient Logging | https://cwe.mitre.org/data/definitions/778.html |
| CWE-798 | Use of Hard-coded Credentials | https://cwe.mitre.org/data/definitions/798.html |
| CWE-862 | Missing Authorization | https://cwe.mitre.org/data/definitions/862.html |

### OWASP Top 10:2021 Cross-Reference

| OWASP Category | Description | Related Checks |
|----------------|-------------|----------------|
| A01:2021 | Broken Access Control | Checks 9, 12 |
| A02:2021 | Cryptographic Failures | Checks 1, 2, 3, 7, 8, 11 |
| A03:2021 | Injection | Check 6 |
| A04:2021 | Insecure Design | Checks 4, 10 |
| A07:2021 | Identification and Authentication Failures | Checks 5, 12 |

## Example Usage

**User prompt:**
> "Run a PCI-DSS audit on this project"

**Expected output (abbreviated):**

```text
## PCI-DSS v4.0 Audit Results

Scanned 24 files across JavaScript, Python, Java

### Findings (8 total: 4 Critical, 3 High, 1 Medium)

| # | Severity | CWE | PCI-DSS Req | Location | Issue |
|---|----------|-----|-------------|----------|-------|
| 1 | Critical | CWE-532 | 3.4 | src/payment/handler.js:23 | Full PAN logged via console.log() |
| 2 | Critical | CWE-312 | 3.4 | src/models/card.py:15 | Card number stored in plaintext in database |
| 3 | Critical | CWE-798 | 8.6 | src/config/gateway.js:3 | Stripe secret key hardcoded in source |
| 4 | Critical | CWE-257 | 3.3.2 | src/models/Payment.java:22 | CVV stored in database column post-auth |
| 5 | High | CWE-862 | 7.2 | src/routes/payment.js:45 | No auth middleware on /api/payments/charge |
| 6 | High | CWE-778 | 10.2 | src/services/refund.py:18 | Refund operation has no audit logging |
| 7 | High | CWE-209 | 3.4 | src/payment/Validator.java:31 | Full card number in exception message |
| 8 | Medium | CWE-384 | 8.3 | src/app.js:12 | Session cookie missing secure and httpOnly flags |

### Recommendations
1. Remove all plaintext PAN logging and implement masked audit logging (Findings #1, #7)
2. Tokenize stored card data via payment provider; remove CVV column (Findings #2, #4)
3. Move gateway credentials to environment variables or secrets manager (Finding #3)
```
