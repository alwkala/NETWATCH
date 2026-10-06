# NETWATCH Project Roadmap

<!-- last-verified: 2026-10-06 -->

> **Local Network Intelligence & Device Inventory**  
> *Instant LAN discovery, historical reconciliation, and persistent device tracking — 100% offline, zero cloud, zero telemetry.*

---

## Strategic Vision

**Know Every Device on Your LAN. Without the Cloud Watching.**  
NETWATCH is not an ephemeral scanner; it is a persistent local network intelligence suite. While discovery sweeps act as the ingestion sensor, the core value lies in continuous asset tracking, historical state reconciliation (joins, leaves, IP migrations), and offline hardware fingerprinting inside a sovereign SQLite ledger (100% offline, zero telemetry, no accounts).

---

## Milestone Progress

```
[M1: Windows Native]       ✅ Completed
        │
[M2: CI & Governance]      ✅ Completed (Multi-OS CI, Dependabot, Licensing, Threat Model)
        │
[M3: Settings & Database]  🔄 In Progress (SQLite Settings, Auto-Scan, DB Maintenance & Stats)
        │
[M4: Better Identity]      📅 Planned (mDNS, NetBIOS, SSDP, Confidence Scores)
        │
[M5: Diagnostics]          📅 Planned (Traceroute, Reverse DNS, Latency History Charts)
        │
[M6: QA & Data Lifecycle]  📅 Planned (Comprehensive Unit/API Tests, Playwright E2E, Export/Backup)
        │
[M7: Signed Release]       📅 Planned (MSI/NSIS Packaging, SignPath.io Authenticode Signing)
```

---

### ✅ Milestone 1: Windows Native Validation (Completed)
- [x] Verified `GetAdaptersAddresses`, `GetIpNetTable`, and unprivileged `IcmpSendEcho` on live LAN (`192.168.1.0/24`).
- [x] Resolved 4-byte ARP table structure offset drift (`sizeof(MIB_IPNETROW) = 24`).
- [x] Wired local data folder handler to Windows Explorer.
- [x] Built standalone native Windows binary `build/bin/netwatch.exe` with GUI subsystem.
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

### 🔄 Milestone 3: Real Settings & Database Management (In Progress)
- [ ] **Settings Persistence & Engine Binding**:
  - [x] Persist user preferences in SQLite (`GET /v1/settings`, `PUT /v1/settings`) instead of transient frontend state.
  - [x] Scheduled background scan execution at configurable intervals (`1m`, `5m`, `15m`, `1h`, `manual`).
  - [ ] Native Windows toast notifications for new device discovery, device offline, and network change.
  - [ ] System Tray integration: minimize-to-tray and start-minimized on login via HKCU Run key.
- [ ] **Database Inspection & Maintenance Panel**:
  - [x] Real-time SQLite statistics card: live database file size, total inventoried devices, event counts, and WAL mode indicators.
  - [x] Robust `Open Data Folder` integration with absolute path resolution (`%WINDIR%\explorer.exe` / fallback) and distinct error/success feedback states in UI.
  - [x] Database health and maintenance operations: `PRAGMA integrity_check` and `VACUUM` (compact database).
  - [ ] Safe event log pruning and retention policy.

---

### 📅 Milestone 4: Enhanced Device Fingerprinting
- [ ] Multi-protocol device discovery: mDNS (Bonjour), NetBIOS, SSDP/UPnP.
- [ ] DHCP lease inspection and Wi-Fi SSID discovery via WLAN API.
- [ ] Probabilistic device classification with confidence score and explicit "Unknown" fallback.
- [ ] User-editable custom device types, aliases, and custom icons.
- [ ] Offline IEEE OUI database updater script.

---

### 📅 Milestone 5: Deep Network Diagnostics
- [ ] Visual hop-by-hop Traceroute tool.
- [ ] Reverse DNS lookup and local DNS latency benchmarking.
- [ ] Per-device latency history charts and jitter measurement.
- [ ] Router and default gateway health metrics from real probes.

---

### 📅 Milestone 6: Quality Assurance, Hardening & Data Portability
- [ ] **End-to-End & Integration Testing**:
  - Playwright end-to-end smoke test suite against `netwatchd` and production desktop UI.
  - Unit tests for `internal/netenv` adapter parsing helpers across Linux and Windows.
  - End-to-end regression tests for new M4/M5 device fingerprinting and diagnostic probes.
- [ ] **Data Portability & Retention**:
  - Export device inventory and event history to JSON and CSV formats.
  - Configurable data retention policies (auto-purge events older than X days).
  - Versioned database schema migrations with automated backup/restore verification.

---

### 📅 Milestone 7: Production Release & Code Signing
- [ ] Windows Installer (MSI / NSIS) and Winget package integration.
- [ ] Free Authenticode code signing integration via **[SignPath.io](https://signpath.io/)** Foundation.
- [ ] Cross-platform desktop release (Linux AppImage/DEB and macOS DMG).

---

### 🔮 Future Horizons (Post-v1.0)
- [ ] **SNMP Support**: Query managed network switches, routers, and enterprise printers.
- [ ] **Physical Topology Mapping**: Infer switch port connections from real LLDP and ARP correlation.
- [ ] **Multi-Network Profiles**: Roaming between home, office, and lab subnets with distinct inventory profiles.
- [ ] **Engine CLI Tools**: Standalone CLI commands (`netwatch discover`, `netwatch scan`, `netwatch export`) powered directly by the core engine.
