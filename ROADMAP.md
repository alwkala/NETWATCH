# NETWATCH Project Roadmap

<!-- last-verified: 2026-10-06 -->

> **Local Network Intelligence & Device Inventory**  
> *Instant LAN discovery, historical reconciliation, and persistent device tracking — 100% offline, zero cloud, zero telemetry.*

---

## Strategic Vision & Architectural Filters

**"Know Every Device on Your LAN. Without the Cloud Watching."**  
NETWATCH is not an ephemeral scanner; it is a persistent local network intelligence suite. While discovery sweeps act as the ingestion sensor, the core value lies in continuous asset tracking, historical state reconciliation (joins, leaves, IP migrations), and offline hardware fingerprinting inside a sovereign SQLite ledger.

### The Three Invariant Filters
No feature or diagnostic tool enters NETWATCH unless it passes three architectural filters:
1. **Invariant Alignment**: Zero cloud dependencies, zero external egress, and unprivileged execution wherever possible.
2. **The Ledger Over the Scanner**: Prioritize capabilities that enrich historical memory, state reconciliation, and longitudinal diffs over transient, one-off scan utilities.
3. **Attack Surface & Integrity Discipline**:
   - **No packet capture drivers** (Npcap / WinPcap) or promiscuous-mode sniffing.
   - **No offensive exploitation** or penetration testing tools.
   - **Authentic Data Only**: No speculative threat scores, simulated health percentages, or hallucinated AI ratings.

---

## Milestone Progress

```
[M1: Windows Native]          ✅ Completed (Unprivileged ICMP/ARP, Rotating Logger, Standalone GUI)
        │
[M2: CI & Governance]         ✅ Completed (Multi-OS CI, Rulesets, Dependabot, Threat Model)
        │
[M3: Settings & Database]     ✅ Completed (SQLite Settings, Auto-Scan, DB Compaction, Toasts, Autostart)
        │
[M4: Discovery & Evidence]    ✅ Completed (Randomized MAC, Device Merge, Evidence Bus: mDNS/SSDP/NBNS)
        │
[M5: Watchdog & Diagnostics]  📅 Planned (ARP Conflict Detection, Traceroute, Latency & Uptime Ledger)
        │
[M6: Diff, Reports & i18n]    📅 Planned (Snapshot Diff, 7-Language i18n, Safe Export, Playwright E2E)
        │
[M7: Signed Production]       📅 Planned (SignPath.io Authenticode, MSI/NSIS, SBOM, Notices)
```

### SemVer Release Alignment

| Milestone | Release Target | SemVer Impact | Focus & Deliverables | Status |
|---|:---:|:---:|---|:---:|
| **M1 + M2** | `v0.1.0-alpha.1` | Initial Baseline | Core engine, Windows unprivileged probes, threat model, multi-OS CI | **Shipped** |
| **M3 + M4** | `v0.2.0-alpha.1` | **MINOR** | Settings, auto-scan, toast notifications, multi-protocol evidence bus | **Ready** |
| **M5** | `v0.3.0-alpha.1` | **MINOR** | Network watchdog, rogue gateway alerts, traceroute, latency trends | Planned |
| **M6** | `v0.4.0-beta.1` | **MINOR** | Historical snapshot diffs, 7-language i18n, safe CSV/JSON exports, E2E | Planned |
| **M7** | `v1.0.0` | **MAJOR** | Authenticode signing, Windows MSI/NSIS installers, public release | Planned |

---

### ✅ Milestone 1: Windows Native Validation (Completed)
- [x] Verified `GetAdaptersAddresses`, `GetIpNetTable`, and unprivileged `IcmpSendEcho` on live LAN (`192.168.1.0/24`).
- [x] Resolved 4-byte ARP table structure offset drift (`sizeof(MIB_IPNETROW) = 24`).
- [x] Wired local data folder handler to Windows Explorer with hidden-window subprocess isolation.
- [x] Built standalone native Windows binary `build/bin/netwatch.exe` with GUI subsystem suppression.
- [x] Thread-safe rotating logger in `%LOCALAPPDATA%\NetWatch\data\netwatch.log`.

---

### ✅ Milestone 2: CI, Security & Governance Baseline (Completed)
- [x] GitHub Actions CI with SHA-pinned actions and multi-OS matrix (Ubuntu `-race` & Windows builds).
- [x] Automated dependency monitoring and security patch alerts via Dependabot (`gomod`, `npm`, `github-actions`).
- [x] Dual licensing (MIT & Apache 2.0) and comprehensive OSS community health files (`CODE_OF_CONDUCT`, `CONTRIBUTING`, `SECURITY`).
- [x] Formal STRIDE Threat Model (`THREAT_MODEL.md`) and Incident Response Plan (`INCIDENT_RESPONSE.md`).
- [x] Unit test suites for API security invariants (Bearer token, DNS-rebinding Host check, Origin isolation) and SQLite Store persistence/transactions.
- [x] Cross-platform build fixes (`openfolder` separation and `frontend/dist/.gitkeep`) ensuring clean checkouts pass on Linux and Windows CI.
- [x] GitHub Repository Rulesets: enforced branch protection on `main` (blocking branch deletion, non-fast-forward pushes, requiring linear history and green CI status checks) and tag immutability on `refs/tags/v*`.

---

### ✅ Milestone 3: Real Settings & Database Management (Completed)
- [x] **Settings Persistence & Engine Binding**:
  - [x] Persist user preferences in SQLite (`GET /v1/settings`, `PUT /v1/settings`) instead of transient frontend state.
  - [x] Scheduled background scan execution at configurable intervals (`1m`, `5m`, `15m`, `1h`, `manual`) with error recovery.
  - [x] Native Windows toast notifications for new device discovery, device offline, and network change (`internal/notifier`).
  - [x] System Startup integration: start-minimized on login via HKCU Run registry (`internal/autostart`), `--minimized` argument flag, and single-instance restoration.
- [x] **Database Inspection & Maintenance Panel**:
  - [x] Real-time SQLite statistics card: live database file size, total inventoried devices, event counts, and WAL mode indicators.
  - [x] Robust `Open Data Folder` integration with absolute path resolution (`%WINDIR%\explorer.exe` / fallback) and distinct error/success feedback states in UI.
  - [x] Database health and maintenance operations: `PRAGMA integrity_check` and `VACUUM` (compact database).
  - [x] Safe event log pruning and retention policy (`POST /v1/data/prune`, `PruneEvents`).
  - [x] Flap-resistant reconciliation requiring both consecutive misses (`Missed >= 2`) and minimum time threshold (`>= 5m`).

---

### ✅ Milestone 4: Multi-Protocol Discovery, Evidence & Trust (Completed)
- [x] **Randomized MAC Detection & Composite Identity**:
  - [x] IEEE 802 Locally Administered Address (LAA) bit detection (`mac[0] & 0x02 != 0`).
  - [x] Distinct UI badge for private/randomized Wi-Fi addresses (iOS, Android, Windows 10/11) with explanatory tooltip.
  - [x] Device Merge & Aliasing workflow (`POST /v1/devices/{targetId}/merge`) to unify fragmented device histories when private MACs rotate.
- [x] **Trust Status & Asset Allowlist**:
  - [x] Device trust classification: `Known` (approved asset), `Guest` (authorized temporary device), and `Unknown` (default for newly discovered assets).
  - [x] User-editable custom device aliases (with Arabic and UTF-8 support up to 80 characters), custom device types, and notes.
- [x] **Multi-Protocol Discovery & Evidence Bus (Unprivileged)**:
  - [x] Passive/multicast DNS (mDNS / DNS-SD RFC 6762/6763) `.local` service discovery and IP correlation (`224.0.0.251:5353`).
  - [x] NetBIOS Name Service (NBNS RFC 1002) directed UDP 137 query prober with 6-byte Unit ID MAC correlation.
  - [x] SSDP / UPnP M-SEARCH multicast listener (`239.255.255.250:1900`) with Zero-Fetch Location invariant.
  - [x] Wi-Fi SSID / Network Context separation (`network_contexts` table, semantic normalization).
  - [x] Protocol-aware Destination Policy (`IsAllowedDiscoveryDestination`) enforcing zero-egress at socket layer.
  - [x] Evidence Bus & Identity Resolver: deterministic Canonical Name precedence, conflict detection, and authentic reason rules.
  - [x] SQLite schema migration v3 (`evidence` and `network_contexts` tables).
  - [x] Identification Evidence UI panel in Device Details with protocol badges and non-clickable copy-only SSDP location.
- [x] **Security & Untrusted Input Hardening**:
  - [x] Strict hostname sanitization pipeline: strip ANSI escapes, control characters, and Unicode Bidi override runes.
  - [x] Automated CI Zero-Egress gate: AST parser test verifying no external `net/http` client dialers exist in Go code and no remote CDN imports in frontend assets.

---

### 📅 Milestone 5: Watchdog Anomaly Detection & Network Diagnostics
- [ ] **Network Watchdog & Anomaly Detection**:
  - [ ] ARP Conflict / Duplicate IP detection: trigger `ip_conflict` alert when multiple distinct MACs claim the same IP within the flap window.
  - [ ] Gateway Impersonation Alert: high-severity event triggered if the default gateway MAC address changes without network interface reconnection.
  - [ ] Local Rules Engine: in-memory rule evaluation on reconciliation diffs (*"Notify if an Unknown device appears after hours"*).
- [ ] **Diagnostics & Latency Ledger**:
  - [ ] Per-device latency & uptime history ledger (bounded ring buffer / SQLite table) with interactive trend charts.
  - [ ] Hop-by-hop Traceroute tool using unprivileged TTL-incremented ICMP echo probes.
  - [ ] Reverse DNS lookup benchmark and local gateway response time tracking.

---

### 📅 Milestone 6: Data Portability, Internationalization (i18n) & Quality Assurance
- [ ] **Internationalization (i18n) & Multi-Language Architecture**:
  - [ ] Lightweight, air-gapped i18n engine with zero external network requests and fully bundled local translation files.
  - [ ] Bundled 7 core locale translation files:
    - English (`en` — default fallback)
    - Arabic (`ar` — native RTL layout direction, Arabic typography, and bidi-hardened strings)
    - German (`de`)
    - French (`fr`)
    - Spanish (`es`)
    - Japanese (`ja`)
    - Simplified Chinese (`zh-CN`)
  - [ ] Extensible JSON locale registry (`frontend/src/locales/<lang>.json`) with strictly typed translation keys, automatic fallback to `en`, and documentation for adding community translations without code refactoring.
  - [ ] Persistent user language preference stored in SQLite settings (`GET/PUT /v1/settings` -> `settings.language`) with immediate UI re-rendering.
- [ ] **Historical Diff & Digest Reports**:
  - [ ] Snapshot Diff Engine (`GET /v1/reports/diff?from=...&to=...`): calculate joined, departed, and IP-drifted devices between any two historical dates.
  - [ ] Scheduled local digest: weekly summary of new, active, and dormant devices.
- [ ] **Safe Data Export**:
  - [ ] CSV export (`GET /v1/export/devices.csv`, `GET /v1/export/events.csv`) with automatic spreadsheet formula injection protection (escaping `=`, `+`, `-`, `@`).
  - [ ] Structured JSON export and backup verification.
- [ ] **Automated End-to-End Testing**:
  - [ ] Playwright E2E smoke test suite exercising `netwatchd` loopback API and production React UI.
  - [ ] Unit tests for `internal/netenv` platform parsing across simulated Windows and Linux environments.

---

### 📅 Milestone 7: Production Release, Code Signing & Distribution
- [ ] **Authenticode Code Signing**:
  - [ ] Automated production binary signing via **[SignPath.io](https://signpath.io/)** Foundation certificate.
  - [ ] Publish SHA-256 integrity checksums for all release artifacts.
- [ ] **Packaging & Installers**:
  - [ ] Windows Installer (MSI / NSIS) with uninstaller and Winget manifest submission.
  - [ ] Cross-platform desktop builds (Linux AppImage/DEB and macOS DMG).
- [ ] **Supply Chain & Licensing Hygiene**:
  - [ ] Software Bill of Materials (SBOM) generation via `syft` or GitHub Dependency Graph.
  - [ ] Dedicated `THIRD_PARTY_NOTICES.md` documenting licenses for bundled OFL fonts (*Plus Jakarta Sans*, *JetBrains Mono*) and IEEE OUI data.

---

### 🔮 Future Horizons (Post-v1.0)
- [ ] **Multi-Subnet & VLAN Support**: Scanning routed RFC 1918 subnets via explicit subnet configuration.
- [ ] **Read-Only SNMP v3**: Query enterprise switches, managed routers, and network printers for interface statistics and VLAN mappings.
- [ ] **Physical Topology Mapping**: Correlate LLDP, ARP tables, and switch port data to generate visual network connection maps.
- [ ] **Standalone Engine CLI**: Unified CLI tooling (`netwatchctl discover`, `netwatchctl diff`, `netwatchctl export`).

---

### 🚫 Non-Goals & Architectural Exclusions
To protect product stability, legal compliance, and user security, the following will **never** be implemented in NETWATCH:
1. **Cloud Telemetry or SaaS Sync**: Network inventory and topology will never leave the host system.
2. **Promiscuous Packet Capture (Npcap / WinPcap)**: Requires elevated kernel drivers, complicates installation, and shifts the risk profile into a surveillance tool.
3. **Offensive Exploitation & Vulnerability Scanning**: NETWATCH is a local asset ledger and intelligence utility, not an offensive penetration testing tool.
4. **Synthetic AI Health Scores or Speculative Threat Ratings**: We adhere strictly to authentic, observable network metrics. Unknown fields remain `Unknown`.
5. **Remote Agent Deployment**: NETWATCH does not install background agents on target network endpoints.
