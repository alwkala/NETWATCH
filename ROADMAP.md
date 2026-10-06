# NETWATCH Project Roadmap

<!-- last-verified: 2026-10-06 -->

This roadmap provides a transparent overview of the development direction and milestones for **NETWATCH**.

---

## Strategic Vision

NETWATCH is transitioning from a validated Windows prototype into a battle-tested, privacy-first cross-platform desktop network intelligence suite (competing with commercial utilities like Fing, but 100% offline and telemetry-free).

---

## Milestone Progress

```
[M1: Windows Native]  ✅ Completed
        │
[M2: Tests & CI]      🔄 In Progress (CI & Governance Done, Unit Tests Active)
        │
[M3: Real Settings]   📅 Planned (Scheduled Auto-Scan, SQLite Persistence, Toast Notifications)
        │
[M4: Better Identity] 📅 Planned (mDNS, NetBIOS, SSDP, Confidence Scores)
        │
[M5: Diagnostics]     📅 Planned (Traceroute, Reverse DNS, Latency History Charts)
        │
[M6: Data Lifecycle]  📅 Planned (JSON/CSV Export, DB Migrations, Retention)
        │
[M7: Signed Release]  📅 Planned (MSI/NSIS Packaging, SignPath.io Authenticode Signing)
```

---

### ✅ Milestone 1: Windows Native Validation (Completed)
- [x] Verified `GetAdaptersAddresses`, `GetIpNetTable`, and unprivileged `IcmpSendEcho` on live LAN (`192.168.1.0/24`).
- [x] Resolved 4-byte ARP table structure offset drift (`sizeof(MIB_IPNETROW) = 24`).
- [x] Wired local data folder handler to Windows Explorer.
- [x] Built standalone native Windows binary `build/bin/netwatch.exe` with GUI subsystem.
- [x] Thread-safe rotating logger in `%LOCALAPPDATA%\NetWatch\data\netwatch.log`.

---

### 🔄 Milestone 2: Tests, CI & Governance (In Progress)
- [x] GitHub Actions CI with SHA-pinned actions and multi-OS matrix (Ubuntu & Windows).
- [x] Dependabot automated dependency monitoring (`gomod`, `npm`, `github-actions`).
- [x] Dual licensing (MIT & Apache 2.0) and community health files.
- [x] Formal STRIDE Threat Model (`THREAT_MODEL.md`) and Incident Response Plan (`INCIDENT_RESPONSE.md`).
- [ ] Unit test suite for `internal/store` (SQLite transactions and event pruning).
- [ ] Unit test suite for `internal/api` (Bearer auth, Origin/Host blocking, SSE stream).
- [ ] Unit tests for `internal/netenv` parsing helpers.
- [ ] Playwright end-to-end smoke tests against `netwatchd` with `-demo` environment.

---

### 📅 Milestone 3: Real Settings & Automation
- [ ] Persist settings in SQLite (`GET /v1/settings`, `PUT /v1/settings`).
- [ ] Scheduled background scan execution at configurable intervals.
- [ ] Native Windows toast notifications for new device discovery, device offline, and network change.
- [ ] System Tray integration: minimize-to-tray and start minimized on login.

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

### 📅 Milestone 6: Data Portability & Retention
- [ ] Export device inventory and event history to JSON and CSV formats.
- [ ] Configurable data retention policies (auto-purge events older than X days).
- [ ] Versioned database schema migrations with automated backup/restore.

---

### 📅 Milestone 7: Production Release & Code Signing
- [ ] Windows Installer (MSI / NSIS) and Winget package integration.
- [ ] Free Authenticode code signing integration via **[SignPath.io](https://signpath.io/)** Foundation.
- [ ] Cross-platform desktop release (Linux AppImage/DEB and macOS DMG).
