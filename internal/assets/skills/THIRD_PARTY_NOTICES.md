# Third-party notices

Skills vendored into `internal/assets/skills/` from external repositories. Each
entry records the upstream source, its license, and the skills taken from it.

## kalshamsi/claude-security-skills

- License: MIT
- Copyright (c) 2026 Khalfan
- Source: https://github.com/kalshamsi/claude-security-skills
- Source path: `skills/<name>/SKILL.md`
- Skills included: api-security-tester, crypto-audit, bandit-sast, socket-sca,
  security-headers-audit, security-test-generator, mobile-security,
  devsecops-pipeline, docker-scout-scanner, pci-dss-audit
- Modifications: description field normalized to <=160 chars and a when_to_use
  field added, per Teicos frontmatter policy; body unchanged

## mattpocock/skills

- License: MIT
- Copyright (c) 2026 Matt Pocock
- Source: https://github.com/mattpocock/skills
- Source path: `skills/engineering/<name>/SKILL.md`
- Skills included: tdd, diagnosing-bugs, code-review
- Modifications: description field normalized to <=160 chars and a when_to_use
  field added, per Teicos frontmatter policy; body unchanged

## License text

Both repositories are distributed under the MIT License:

```
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
