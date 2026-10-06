# NETWATCH Threat Model

<!-- last-verified: 2026-10-06 -->

> **Privacy-First Local Network Intelligence & Device Inventory**  
> *Instant LAN discovery and persistent device tracking — 100% offline-first, zero cloud, zero telemetry.*

This document formalizes the threat landscape, security boundaries, residual risks, and controls for **NETWATCH**, adhering to the STRIDE methodology.

For vulnerability reporting instructions, see [SECURITY.md](SECURITY.md).<br>
For emergency incident handling, see [INCIDENT_RESPONSE.md](INCIDENT_RESPONSE.md).

---

## 1. Scope & System Boundaries

NETWATCH is a local-first desktop application composed of:
1. **Wails / WebView2 Desktop Shell**: Host process rendering the locally bundled React 19 UI (all fonts and assets bundled locally, zero CDN imports).
2. **Go Core Engine**: Native process interfacing with OS network adapters, ARP caches, and unprivileged ICMP sockets.
3. **Local Loopback API**: `127.0.0.1:<random-port>` REST + SSE service protected by an ephemeral 192-bit Bearer Token, strict `Host` validation, and loopback-only `Origin` allowlisting.
4. **Local SQLite Database**: `%LOCALAPPDATA%\NetWatch\data\network.db` holding local device inventory and historical audit logs.

### Key Architectural Invariant:
**Offline-First / Zero External Egress**: Network inventory, IP mappings, and MAC addresses never leave the host machine. The application executes zero outbound HTTP client requests and makes zero telemetry/analytics dials. This is verified by an automated CI gate (`internal/netenv/egress_test.go`).

---

## 2. Assets to Protect

| Asset | Why It Matters |
|---|---|
| **Host System Integrity** | NETWATCH interacts with low-level network APIs (`iphlpapi.dll`). Vulnerabilities must never grant unprivileged callers elevated rights. |
| **Local Network Topology Privacy** | Discovered MAC addresses, private IP maps, hostnames, and vendor identities must remain confidential to the local user. |
| **Loopback API Authentication** | The engine must never accept commands from malicious web pages open in local browsers (anti-DNS rebinding). |
| **Local SQLite Inventory** | Device history, notes, and audit ledger must be guarded against unauthorized cross-process tampering. |
| **Integrity of Rendered LAN Data** | Untrusted hostnames and strings broadcast by LAN devices must not execute script injection or corrupt log/export files. |

---

## 3. STRIDE Threat Analysis & Residual Risks

| Threat (STRIDE) | Attack Vector | Potential Impact | NETWATCH Mitigation | Residual Risk & Out-of-Scope | Status |
|---|---|---|---|---|---|
| **Spoofing (S)** | Malicious web page in Chrome/Edge attempts to invoke `http://127.0.0.1:<port>/v1/scans`. | Unauthorized network sweeps or settings manipulation. | **Triple-Layer Loopback Protection**: Ephemeral 192-bit Bearer Token, strict `Host` validation, and loopback-only `Origin` verification (`internal/api/server.go`). | Local malware running with the user's privilege that can read memory can acquire the token. Guarding against local host compromise is out of scope. | **Active** |
| **Tampering (T)** | Attacker-controlled hostnames from LAN (rDNS, mDNS, NetBIOS). | Log injection, Unicode Bidi spoofing, CSV formula injection on export. | **Strict LAN String Sanitization**: `SanitizeLANString` strips control runes, ANSI codes, and Bidi overrides (`U+202A-202E`, `U+2066-2069`). `EscapeCSVField` escapes formula triggers (`=,+,-,@`). | An attacker on the LAN can still set a legitimate-looking DNS PTR record to confuse a human operator. | **Active** |
| **Tampering (T)** | Malicious local process alters `%LOCALAPPDATA%\NetWatch\data\network.db`. | Corruption of device history or misleading device records. | User-profile ACLs (`%LOCALAPPDATA%`), thread-safe SQLite transactions, serialized inventory diff reconciliation. | Physical disk compromise or another process running as the same Windows user can write to the database file. | **Active** |
| **Repudiation (R)** | User or operator denies network changes or scan operations. | Inability to audit when an unknown device joined the LAN. | Local Event Ledger with timestamps, device identity continuity, and non-repudiation audit marker on history purge (`history_cleared_at`). | Ledger is local and user-clearing is supported; it does not constitute a legal, immutable blockchain. | **Active** |
| **Information Disclosure (I)** | Exfiltration of LAN MAC addresses or hostnames via OUI lookup. | Third-party analytics tracking private home/enterprise hardware. | **Local Embedded Database**: OUI lookup uses offline embedded IEEE file (`internal/oui/ieee-oui.txt`). Automated CI gate verifies zero external dials. | Unauthenticated physical shoulder-surfing of the desktop UI. | **Active** |
| **Denial of Service (D)** | LAN host floods ICMP/ARP replies during scan sweep. | Engine memory exhaustion or UI thread lockup. | Worker pool concurrency limits, strict socket read deadlines, non-blocking SSE streaming channels. | Network interface saturation by external flooding attacks can still prevent NETWATCH from receiving replies. | **Active** |
| **Elevation of Privilege (E)** | Exploiting raw socket or packet parsing to execute code as Administrator. | System compromise. | **Unprivileged Native APIs**: Windows implementation uses `iphlpapi.dll` (`IcmpSendEcho`, `GetIpNetTable`, `SendARP`) executing in user-mode without UAC/Admin rights. | If Windows kernel drivers (`iphlpapi.dll`) have an unpatched zero-day vulnerability. | **Active** |

---

## 4. Security Verification Checklist

- [x] Backend listens exclusively on `127.0.0.1` (never on `0.0.0.0`).
- [x] Random ephemeral port selected per startup (`api.ListenLoopback(0)`).
- [x] Per-session cryptographically secure 192-bit Bearer Token.
- [x] Rejection of external `Origin` headers and forged `Host` headers.
- [x] Strict server-side target validation for probe tooling (RFC 1918, CGNAT, link-local only).
- [x] Rotating file logger capped at 5 MB with 3 backups (`internal/appdata/logger.go`).
- [x] Offline OUI parsing with zero network egress.
- [x] Untrusted LAN string sanitization stripping Bidi overrides and control runes (`internal/fingerprint/sanitize.go`).
- [x] CSV formula injection defense (`internal/fingerprint/sanitize.go`).
- [x] Automated CI Zero-Egress gate (`internal/netenv/egress_test.go`).

---

## 5. Explicit Non-Goals & Out-of-Scope Threats

To preserve architectural clarity and minimize attack surface, NETWATCH explicitly excludes the following from its threat model:
1. **Compromised Host**: If the machine running NETWATCH already has malicious rootkit or user-level spyware installed, that malware can inspect process memory, read SQLite databases, or terminate the process.
2. **Malicious Local Administrator**: An administrator on the local host can override file permissions, inspect memory, or debug the application.
3. **L2 Network Impersonation at Wire Level**: Unauthenticated Layer-2 LAN protocols (ARP, mDNS) are inherently forgeable by any device connected to the physical switch or Wi-Fi AP. NETWATCH detects conflicts and reports anomalies, but cannot cryptographically authenticate physical wire transmissions.
