# NETWATCH Threat Model

<!-- last-verified: 2026-10-06 -->

This document formalizes the threat landscape, security boundaries, and mitigations for **NETWATCH**, adhering to the STRIDE methodology.

For vulnerability reporting instructions, see [SECURITY.md](SECURITY.md).<br>
For emergency incident handling, see [INCIDENT_RESPONSE.md](INCIDENT_RESPONSE.md).

---

## 1. Scope & System Boundaries

NETWATCH is a local-first desktop application composed of:
1. **Wails / WebView2 Desktop Shell**: Host process rendering the bundled React 19 UI.
2. **Go Core Engine**: Native process interfacing with OS network adapters, ARP caches, and ICMP sockets.
3. **Local Loopback API**: `127.0.0.1:<random-port>` REST + SSE service protected by an ephemeral 192-bit Bearer Token.
4. **Local SQLite Database**: `%LOCALAPPDATA%\NetWatch\data\network.db` holding local device inventory and historical logs.

### Key Architectural Invariant:
**Zero Cloud / Zero Telemetry**: Network telemetry never leaves the host machine. The UI makes zero external requests (all fonts and assets are air-gapped).

---

## 2. Assets to Protect

| Asset | Why It Matters |
|---|---|
| **Host System Integrity** | NETWATCH interacts with low-level network APIs (`iphlpapi.dll`). Vulnerabilities must never grant unprivileged callers elevated rights. |
| **Local Network Topology Privacy** | Discovered MAC addresses, private IP maps, hostnames, and vendor identities must remain confidential to the local user. |
| **Loopback API Authentication** | The engine must never accept commands from malicious web pages open in local browsers (DNS-rebinding attacks). |
| **Local SQLite Inventory** | Device history and notes must be guarded against unauthorized local process tampering. |

---

## 3. STRIDE Threat Analysis

| Threat (STRIDE) | Attack Vector | Potential Impact | NETWATCH Mitigation | Status |
|---|---|---|---|---|
| **Spoofing (S)** | Malicious web page in Chrome/Edge attempts to invoke `http://127.0.0.1:<port>/v1/scans`. | Unauthorized network sweeps or settings manipulation. | **Triple-Layer Loopback Protection**: Ephemeral 192-bit Bearer Token, strict `Host` validation (anti-DNS rebinding), and `Origin` allow-list enforcement (`internal/api/server.go`). | **Active** |
| **Tampering (T)** | Malicious local process alters `%LOCALAPPDATA%\NetWatch\data\network.db`. | Corruption of device history or misleading device records. | User-profile ACLs (`%LOCALAPPDATA%`), thread-safe SQLite transactions, serialized inventory diff reconciliation. | **Active** |
| **Repudiation (R)** | Denying network changes or actions performed during a scan session. | Inability to audit when an unknown device appeared on LAN. | Immutable Append-Only Event Log (`internal/store/store.go`) tracking timestamps and state transitions. | **Active** |
| **Information Disclosure (I)** | Exfiltration of LAN MAC addresses or hostnames via OUI lookup. | Third-party analytics tracking private home/enterprise hardware. | **Local Embedded Database**: OUI lookup uses offline embedded IEEE file (`internal/oui/ieee-oui.txt`). Zero remote queries. | **Active** |
| **Denial of Service (D)** | LAN host floods ICMP/ARP replies during scan sweep. | Engine memory exhaustion or UI thread lockup. | Worker pool concurrency limits, strict socket read deadlines, non-blocking SSE streaming channels. | **Active** |
| **Elevation of Privilege (E)** | Exploiting raw socket or packet parsing to execute code as Administrator. | System compromise. | **Unprivileged Native APIs**: Windows implementation uses `iphlpapi.dll` (`IcmpSendEcho`, `GetIpNetTable`, `SendARP`) that execute safely in user-mode without UAC/Admin rights. | **Active** |

---

## 4. Security Verification Checklist

- [x] Backend listens exclusively on `127.0.0.1` (never on `0.0.0.0`).
- [x] Random ephemeral port selected per startup (`api.ListenLoopback(0)`).
- [x] Per-session cryptographically secure 192-bit Bearer Token.
- [x] Rejection of external `Origin` headers and forged `Host` headers.
- [x] Rotating file logger capped at 5 MB with 3 backups (`internal/appdata/logger.go`).
- [x] Offline OUI parsing with zero network egress.
