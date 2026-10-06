# Changelog

All notable changes to the NETWATCH project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.0-alpha.1] - 2026-10-06

### Added
- **Device Type & Icon Customization**:
  - Added user-customizable device type selector with live icon preview in `DeviceDetails.tsx`.
  - Added `type` to `DevicePatch` wire contract and `PATCH /v1/devices/{id}` endpoint with server-side validation.
  - Added SQLite schema migration v4 with `custom_type` column, guaranteeing user-defined classifications and icons are permanently preserved against heuristic overwrites during background rescans.
- **WAN Public IP Discovery & Topology Integration**:
  - Implemented lightweight RFC 5389 STUN NAT traversal (`internal/netenv/wan.go`) discovering the external public IPv4 address over raw UDP while maintaining zero-egress HTTP compliance.
  - Added `publicIp` to `NetworkInfo` wire contract and displayed the external address directly on the `WAN / INTERNET` node in `Network.tsx`.
- **UI Ergonomics & Mobile-First Responsive Polish (/tidyfactor-styler)**:
  - Simplified global scan action button label from `Scan Network` to `Scan`.
  - Removed outdated section annotation `(Section 18)` from Network Topology Architecture title.
  - Implemented mobile-first responsive layout with slide-over drawer navigation and hamburger toggle in `Sidebar.tsx` and `TopBar.tsx`, while strictly preserving Sidebar brand identity invariants.
  - Enhanced responsive grid wrapping in `DeviceDetails.tsx` and horizontal scroll safety in `Network.tsx` topology diagram.
- **Multi-Protocol Discovery & Evidence Bus (M4)**:
  - Implemented `DiscoveryProbe` interface and unified `DiscoveryScope` contract in `internal/engine/probe.go` and `internal/model/evidence.go`.
  - Added NetBIOS Name Service (NBNS) directed UDP 137 query prober (`internal/netenv/nbns.go`) extracting RFC 1002 name table and 6-byte Unit ID (hardware MAC address).
  - Added NBNS Unit ID vs. ARP MAC correlation: flags MAC mismatches as diagnostic conflicts rather than performing blind merges.
  - Implemented SSDP / UPnP multicast discovery prober (`internal/netenv/ssdp.go`) over `239.255.255.250:1900` with strict Zero-Fetch Location invariant (records Location string only, never dials or fetches HTTP XML).
  - Implemented multi-pass mDNS / DNS-SD observer and prober (`internal/netenv/mdns.go`) over `224.0.0.251:5353` with defense against compression pointer loops, oversized packets, and memory poisoning.
  - Added protocol-aware Destination Policy (`internal/netenv/policy.go`): enforces `IsAllowedDiscoveryDestination` at the socket layer, strictly rejecting loopback for discovery probes and preventing public WAN egress.
  - Implemented Identity Resolver (`internal/engine/resolver.go`): normalizes raw evidence across all probes, verifies MAC integrity, correlates observed hostnames, and determines canonical display names.
  - Upgraded fingerprint engine (`internal/fingerprint/rules.go`): evaluates authentic observable evidence rules (e.g. `mDNS:_ipp._tcp` -> Printer, `SSDP:MediaRenderer` -> TV) with zero speculative confidence percentages.
  - Added SQLite schema migration v3 (`internal/store/store.go`): creates `evidence` table and `network_contexts` table for persistent intelligence and context tracking.
  - Added Identification Evidence panel in Device Details modal (`frontend/src/pages/DeviceDetails.tsx`) displaying protocol badges, structured mDNS/SSDP/NBNS/ARP evidence, and copy-only non-clickable SSDP Location.
- **Randomized MAC Detection & Identity Unification (M4)**:
  - Added IEEE 802 Locally Administered Address (LAA) bit detection (`mac[0] & 0x02 != 0`) and `isRandomizedMac` property in wire models and SQLite schema migration v2.
  - Added "Private MAC" badge across Device Table and Device Details drawer with tooltip explaining private Wi-Fi addresses.
  - Implemented Device Merge & Aliasing engine (`POST /v1/devices/{targetId}/merge`): unifies fragmented device records, transfers event history, registers MAC aliases in `device_mac_aliases`, and automatically reconciles future sweeps under the canonical identity.
- **Trust Status & Asset Allowlist (M4)**:
  - Added three-tier asset trust classification: `known` (Approved), `guest` (Temporary Visitor), and `unknown` (Unclassified).
  - Added trust status filtering and quick toggle controls directly in the Device Table and Details drawer.
  - Exposed `trustStatus` updates via `PATCH /v1/devices/{id}` with strict server-side validation.
- **Untrusted LAN String Sanitization (M4)**:
  - Added `SanitizeLANString` in `internal/fingerprint/sanitize.go`: strips control characters, ANSI escape codes, and Unicode Bidi overrides, clamping lengths to 64 runes.
  - Added `EscapeCSVField` to mitigate CSV formula injection (CWE-1236) by escaping execution trigger characters (`=`, `+`, `-`, `@`, `\t`, `\r`).
- **Automated Zero-Egress CI Gate (M4)**:
  - Added `internal/netenv/egress_test.go`: AST analysis verifying that no package outside `internal/api` imports `net/http`, no outbound `http.Client` is instantiated, and frontend assets never link to remote CDN domains.
- **Native Windows Toast Notifications (M3)**:
  - Implemented asynchronous, non-blocking WinRT desktop toast notifications using `ToastNotificationManager` via PowerShell 5.1 runtime in `internal/notifier/notifier_windows.go`.
  - Added clean cross-platform stubs in `internal/notifier/notifier_other.go` for Linux and macOS.
  - Wired notification triggers into `internal/engine/scan.go` respecting user preferences: `NotifyNewDevice`, `NotifyDeviceOffline`, and `NotifyNetworkChange`.
  - Added test suite in `internal/notifier/notifier_test.go`.
- **Windows Autostart via Registry (M3)**:
  - Created `internal/autostart` package managing `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` entries with `--minimized` launch arguments.
  - Wired `LaunchAtStartup` preference directly into the REST API (`PUT /v1/settings`).
  - Added test suite in `internal/autostart/autostart_test.go`.
- **Desktop Single-Instance Lock & Minimized Launch (M3)**:
  - Added `--minimized` / `-minimized` CLI argument handling via Wails `StartHidden: true` in `main.go`.
  - Configured `SingleInstanceLock` in `main.go` and `RestoreWindow` in `app.go` to bring existing instances to the foreground when launched again.
- **SQLite Event History Pruning & Retention (M3)**:
  - Added `PruneEvents(ctx, olderThanDays)` in `internal/store/store.go` and `internal/engine/engine.go` to purge audit logs and device history.
  - Exposed `POST /v1/data/prune` loopback endpoint in `internal/api/server.go`.
  - Added unit test cases in `internal/store/store_test.go` and `internal/api/server_test.go`.
  - Added "Prune Events (>30d)" maintenance action and removed preview badges in `frontend/src/pages/Settings.tsx`.
### Fixed
- **API CORS & Preflight Support (H1)**:
  - Added `PUT` to `Access-Control-Allow-Methods` in `internal/api/server.go`, resolving browser CORS preflight rejections when persisting settings via Vite/WebView2.
- **Auto-Scan Scheduler Resilience (H2)**:
  - Fixed background auto-scan loop hanging indefinitely upon transient SQLite store errors; added automatic 1-minute retry logic and downgraded offline scan log noise to debug.
- **Flap Resistance Dual-Condition (H3)**:
  - Enhanced offline state reconciliation in `internal/engine/reconcile.go` to require both `Missed >= 2` and `Now - LastSeen >= 5m`, preventing false offline flips on rapid scans or sleeping Wi-Fi radios.
- **Network Key & Identity Stability (H4, L3)**:
  - Added pre-emptive gateway ARP echo in `internal/engine/engine.go` to ensure gateway MAC address availability on first pass; guarded `netKey` persistence against write amplification on read paths.
  - Standardized MAC addresses in `deviceID` using `net.ParseMAC` to eliminate ID drift caused by delimiter variations.
- **Unicode, Arabic & Bidi Character Hardening (M1, M2)**:
  - Migrated device alias and notes length validation to `utf8.RuneCountInString`, enabling full-length Arabic names up to 80 characters.
  - Sanitized control characters and Unicode bidirectional spoofing override runes (`U+202A–202E`, `U+2066–2069`).
  - Added Unicode support to network interface `slug()` generator with SHA-256 fallback to prevent collision of `if-` IDs on non-English Windows editions.
- **Secure Data Erasure & Privacy Audit (M4)**:
  - Enhanced `ClearHistory` in `internal/store/store.go` to record a non-repudiation audit timestamp (`history_cleared_at`), execute `PRAGMA wal_checkpoint(TRUNCATE)`, and trigger `VACUUM` to physically reclaim and wipe SQLite disk pages.
- **Engine Shutdown & Concurrency Safeguards (M5, M6, L4)**:
  - Added `sync.WaitGroup` and `Engine.Stop()` for orderly shutdown of background goroutines on window exit.
  - Added in-flight concurrency lock to `ScanDevicePorts` preventing duplicate port scan storms on target devices.
  - Replaced blocking sleep with context-aware timer in SSE stream handler.
  - Sanitized 500 error responses and enforced strict `Bearer ` token parsing.

### Changed
- **SQLite Engine Driver Upgrade**:
  - Upgraded `github.com/ncruces/go-sqlite3` from `v0.30.0` to `v0.35.6` (resolves critical Windows WAL mode data corruption under heavy concurrency, Issue #404).
- **Core Dependencies**:
  - Upgraded `golang.org/x/sys` from `v0.37.0` to `v0.48.0`.
  - Upgraded `lucide-react` from `0.546.0` to `0.577.0`.
- **Automated Dependency Governance & Hardening**:
  - Codified permanent ignore policies in `.github/dependabot.yml` preventing breaking major version updates for `github.com/wailsapp/wails/v2`, `motion`, and `@types/node`.
  - Created repository label `type:dependency`.
- **Version Manifest Synchronization**:
  - Synchronized `var version` in `main.go` and `cmd/netwatchd/main.go` to match the official `0.1.0-alpha.1` release.

## [0.1.0-alpha.1] - 2026-10-06
 
### Added
- **GitHub Repository Rulesets & Branch Governance**:
  - Configured active ruleset on default branch `main` blocking branch deletions and force pushes (`non_fast_forward`).
  - Enforced required CI status checks (`Go Engine Tests (ubuntu-latest)`, `Go Engine Tests (windows-latest)`, `Frontend Typecheck & Build`) under strict base-branch synchronization policy.
  - Enforced linear history and conversation thread resolution before merging.
  - Configured tag immutability ruleset protecting `refs/tags/v*` releases against deletion and modifications.
- **Official GitHub Release Publication**:
  - Published release `v0.1.0-alpha.1` with attached standalone portable executable `netwatch.exe` (19.9 MB).
- **Windows Native Verification (M1)**:
  - Verified `GetAdaptersAddresses`, `GetIpNetTable`, and unprivileged `IcmpSendEcho` on real Windows LAN (`192.168.1.0/24`).
  - Successfully discovered network adapters, default gateway, DNS servers, and live devices.
- **Rotating File Logger**:
  - Implemented thread-safe rotating file logger in `internal/appdata/logger.go` with 5 MB threshold and 3 backup rotations.
  - Logs are persisted to `%LOCALAPPDATA%\NetWatch\data\netwatch.log` and mirrored to stderr.
  - Added unit test suite in `internal/appdata/logger_test.go`.
- **Desktop Executable Build**:
  - Built standalone native Windows binary `build/bin/netwatch.exe` with GUI subsystem flag (`-H windowsgui`) and required Wails build tags (`-tags desktop,production`), embedding the complete Vite React UI and local bundled fonts.
- **Data Folder Integration**:
  - Wired `OpenDataFolder` handler to open `%LOCALAPPDATA%\NetWatch\data` via Windows Explorer in both Wails host and `netwatchd`.
- **Open Source Governance & CI/CD Infrastructure (M2 baseline)**:
  - Established dual licensing under MIT (`LICENSE-MIT`) and Apache 2.0 (`LICENSE-APACHE`).
  - Formalized STRIDE threat model in `THREAT_MODEL.md` (unprivileged API isolation, anti-DNS rebinding, loopback security).
  - Added structured incident response plan in `INCIDENT_RESPONSE.md`.
  - Created public roadmap document `ROADMAP.md` tracking milestones M1 through M7.
  - Implemented GitHub Actions CI workflow in `.github/workflows/ci.yml` with SHA-pinned actions, concurrency cancellation, and cross-platform matrix testing on Ubuntu (`-race`) and Windows.
  - Added automated dependency management in `.github/dependabot.yml` covering `gomod`, `npm`, and `github-actions`.
  - Added `.github/CODEOWNERS` mapping ownership to `@alwkala`, and `.github/FUNDING.yml` for sponsorships.
  - Standardized repo hygiene with `.editorconfig`, `.gitattributes`, and `CONTRIBUTORS.md`.
  - Created structured issue forms in `.github/ISSUE_TEMPLATE/` (`bug_report.yml`, `feature_request.yml`, `config.yml`) and `.github/pull_request_template.md`.
  - Added `SECURITY.md` (vulnerability disclosure SLA & loopback privacy invariants) and `CONTRIBUTING.md` (local-first design rules).
  - Adopted Contributor Covenant v2.1 in `CODE_OF_CONDUCT.md`.
  - Redesigned `README.md` with elevated positioning: *Local Network Intelligence & Device Inventory*, hero value propositions, architectural comparison matrix, visual architecture, and collapsible diagnostics.
  - Updated repository topics and description on GitHub via `gh` CLI: *"Local Network Intelligence & Device Inventory. Instant LAN discovery, historical reconciliation, 100% offline, zero cloud."*

### Fixed
- **Dark Mode Variant Resolution (Tailwind CSS v4)**:
  - Added explicit `@custom-variant dark (&:where(.dark, .dark *));` to `frontend/src/index.css` to enable class-based dark mode switching in Tailwind CSS v4.
  - Removed hardcoded dark body background in `frontend/index.html` and enabled adaptive transitions (`bg-neutral-100 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-100`).
  - Added dynamic fallback to `prefers-color-scheme` in `ThemeContext.tsx` when no localStorage preference is set.
- **Windows ARP Table Desynchronization**:
  - Corrected `ipNetRowSize` in `internal/netenv/env_windows.go` from `28` to `24` bytes (`sizeof(MIB_IPNETROW)` on Windows). Previously, the 4-byte offset drift per row corrupted row parsing, causing the Default Gateway (`192.168.1.1`) and most neighboring devices to be skipped.
- **Network Topology Router Inspection**:
  - Updated `frontend/src/pages/Network.tsx` to match the actual gateway device by IP and type dynamically, and eliminated the static mock TP-Link fallback that caused the "No device selected" screen upon clicking.
- Fixed missing imports in `cmd/netwatchd/main.go` (`os/exec`, `syscall`) and `main.go` (`netwatch/internal/appdata`).
- Cleaned up proxy mirror `replace` directives in `go.mod` for direct, clean module resolution on standard Windows environments.

### Changed
- **Strategic Product Re-Positioning & Narrative Elevation**:
  - Re-positioned product from ephemeral scanner to *Local Network Intelligence & Device Inventory* across README, ROADMAP, AGENTS handoff, and Settings About card.
  - Added dedicated architectural matrix contrasting ephemeral "fire-and-forget" scanners with persistent SQLite asset ledgers, historical timelines, and reconciliation state tracking.
  - Synchronized repository metadata on GitHub with `network-intelligence` topic and elevated tagline.
- **Desktop Window & Header Customization**:
  - Eliminated redundant simulated `<TitleBar />` in production desktop host, reclaiming vertical space and aligning with Windows native frame controls.
  - Added embedded multi-resolution Windows application icon resource (`rsrc_windows_amd64.syso`) with Icon ID 3 mapping (`winc.AppIconID = 3`) generated via `go-winres` for native Windows TitleBar, Taskbar, Alt+Tab, and Explorer integration.
  - Integrated official vector `favicon.svg` directly into the Sidebar logo lockup and Settings About card.
  - Set native window title to `NETWATCH — Local Network Intelligence`.
  - Added window background color in Wails options to prevent white flashes upon window creation.
- **Search & Scan Performance Optimization**:
  - Integrated React 19 `useDeferredValue` into device and event search filters for zero-latency, 60 FPS input response.
  - Slashed Quick Scan duration with tighter LAN ICMP ping timeouts (`200ms`), optimized TCP fallback discovery on primary ports (`150ms`), and accelerated reverse DNS lookup (`200ms`).
- **Developer & Application Identity**:
  - Added Alwkala developer identity capsule to the Sidebar footer with pulsing engine status, version tag, and GitHub link.
  - Added dedicated **About NETWATCH & Alwkala Engineering** panel in Settings displaying architecture details, privacy invariants, and open-source licenses.
- **Scan & Search UI Consolidation**:
  - Resolved visual and semantic duplication between network probing (`Scan`) and device text filtering (`Search`).
  - Unified the global scan action in `TopBar` with the dedicated `Radar` icon, automatically hiding it when navigating to the dedicated `Scanner` page to prevent stacked button redundancy.
  - Eliminated redundant `Scan Subnet` and `Scan Network` buttons from the page headers in `Devices` and `Dashboard`.
  - Restricted the `Search` (magnifying glass) icon strictly to text query inputs (`DeviceTable`).
  - Standardized the repeat scan button in the `Scanner` page to `Rescan` with a sync icon.
- **Database Management & Settings Persistence (Milestone 3)**:
  - Implemented persistent user preferences in SQLite (`GET /v1/settings`, `PUT /v1/settings`) stored in `meta` table, seamlessly bound to UI toggles.
  - Added live Database Inspection & Health panel in Settings with real-time stats: database file size, stored devices, events, scans, and WAL mode indicators (`GET /v1/data/stats`).
  - Fixed Windows Explorer path resolution in `OpenDataFolder` using absolute `%WINDIR%\explorer.exe` or `C:\Windows\explorer.exe` fallback across `app.go` and `cmd/netwatchd/main.go`.
  - Added direct database maintenance actions: SQLite VACUUM compaction (`POST /v1/data/vacuum`) and integrity validation (`POST /v1/data/integrity`).
  - Resolved UI feedback error display: differentiated green success alerts (`Check`) from red error alerts (`AlertCircle`).
- **Cross-Platform CI & Daemon Build Stability**:
  - Split `cmd/netwatchd` data folder opener into `openfolder_windows.go` (Windows `syscall.SysProcAttr{HideWindow: true}`) and `openfolder_other.go` (Linux `xdg-open` / macOS `open`), resolving compilation failure on Linux/macOS runners.
  - Added `frontend/dist/.gitkeep` ensuring clean checkouts compile successfully without missing embed directory errors.
  - Removed residual generator critique markers from `.github/workflows/ci.yml`, `SECURITY.md`, and `.github/CODEOWNERS`.
- **API & Store Unit Test Coverage**:
  - Added test suite in `internal/api/server_test.go` covering triple-layer loopback security invariants (192-bit Bearer token, anti-DNS rebinding Host check, Origin allow-lists), unauthenticated `/health`, and REST endpoints.
  - Added test suite in `internal/store/store_test.go` covering SQLite migrations, writeset transactions, device patch updates, settings persistence, database metrics, VACUUM compaction, and integrity validation.
- **Documentation Refinements**:
  - Refined Linux netenv documentation to accurately cite `/proc/net/arp`.
  - Clarified sweep performance description to focus on concurrent ARP/ICMP/TCP parallel probing.
  - Clarified packaging status: standalone portable `.exe` active, installer scheduled for M7.
  - Added security advisory on `-token` flag visibility in local process tables for `netwatchd`.
  - Updated `THREAT_MODEL.md` Repudiation mitigation to explicitly describe user-controlled local ledger clearing.
- **Engine Auto-Scan Scheduling & Settings Validation**:
  - Implemented background auto-scan scheduler worker in `internal/engine/engine.go` with dynamic interval reconfiguration and non-blocking trigger execution.
  - Added strict validation in `PUT /v1/settings` enforcing allowed scan intervals (`1m`, `5m`, `15m`, `1h`, `manual`) with 400 Bad Request rejection for invalid payloads.
  - Enhanced error handling in `Store.Stats` by validating and propagating query scan errors instead of discarding them.
  - Tuned Quick scan timeouts (ICMP 450ms, TCP 250ms) to prevent sleeping Wi-Fi devices from prematurely flapping offline.
  - Consolidated duplicate folder opener logic into `internal/appdata` with platform-separated implementations (`openfolder_windows.go` and `openfolder_other.go`).
  - Added SSE stream test suite in `internal/api/server_test.go` verifying real-time progress events.
  - Added transparent "Coming in M3" badges to staged system integration toggles in `Settings.tsx`.
  - Replaced competitor matrix in `README.md` with verifiable Design Principles & Architectural Guarantees.
  - Formatted all Go sources with `gofmt -w`.

[Unreleased]: https://github.com/alwkala/NETWATCH/compare/v0.2.0-alpha.1...HEAD
[v0.2.0-alpha.1]: https://github.com/alwkala/NETWATCH/compare/v0.1.0-alpha.1...v0.2.0-alpha.1
[0.1.0-alpha.1]: https://github.com/alwkala/NETWATCH/releases/tag/v0.1.0-alpha.1

