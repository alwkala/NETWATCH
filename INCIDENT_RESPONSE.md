# Incident Response Plan

<!-- last-verified: 2026-10-06 -->

This document describes the operational protocol for handling security vulnerabilities and supply-chain incidents affecting **NETWATCH**.

- For vulnerability reporting details, see [SECURITY.md](SECURITY.md).
- For the formal system threat model, see [THREAT_MODEL.md](THREAT_MODEL.md).

---

## 1. Incident Scope

An incident constitutes any event that compromises or risks compromising the security, privacy, or integrity of NETWATCH:
- Bypass of the loopback API authentication token or Origin filter.
- Remote or local code execution via network packet parsing or SQLite deserialization.
- Unauthorized telemetry or data exfiltration.
- Compromised dependency, maintainer account, or signing keys.
- Tampered release binaries.

---

## 2. Roles & Responsibilities

- **Incident Lead**: Core project maintainers ([@alwkala](https://github.com/alwkala)). Owns remediation triage and disclosures.
- **Reporter**: Credited in the release notes and advisory (unless anonymity is requested).
- **Security Reviewers**: Domain engineers reviewing the private fix before release.

---

## 3. The 6-Stage Response Lifecycle

```
[1. Triage] ──▶ [2. Contain] ──▶ [3. Fix] ──▶ [4. Release] ──▶ [5. Disclose] ──▶ [6. Review]
```

### Stage 1: Triage (Within 48h)
1. Open a confidential **GitHub Security Advisory** in `alwkala/NETWATCH`.
2. Reproduce the vulnerability in an isolated test environment.
3. Assess CVSS severity and determine affected versions.

### Stage 2: Contain
1. If an active exploit is in the wild, retract compromised releases or issue a critical advisory.
2. Isolate compromised credentials, tokens, or repository permissions immediately.

### Stage 3: Fix & Verification
1. Develop the fix in a private branch within the Security Advisory.
2. Write automated regression tests in Go or Playwright.
3. Verify that the fix does not break local-first architectural invariants.

### Stage 4: Release
1. Tag a patch release following Semantic Versioning (e.g. `v0.1.1`).
2. Run automated release builds and compile standalone binaries.
3. Update [CHANGELOG.md](CHANGELOG.md).

### Stage 5: Responsible Disclosure
1. Publish the GitHub Security Advisory.
2. Request a CVE identifier through GitHub's CVE CNA service if applicable.
3. Credit the security researcher.

### Stage 6: Post-Mortem & Review
1. Conduct a root-cause analysis.
2. Update [THREAT_MODEL.md](THREAT_MODEL.md) and CI supply-chain policies to prevent recurrence.
