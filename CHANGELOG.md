# Changelog

All notable changes to the NETWATCH project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/alwkala/NETWATCH/compare/v0.1.0-alpha.1...HEAD
[0.1.0-alpha.1]: https://github.com/alwkala/NETWATCH/releases/tag/v0.1.0-alpha.1

