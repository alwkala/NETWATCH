# AGENTS.md — NETWATCH Engineering & AI Agent Harness

> **Read this file fully before introducing changes.**  
> Then inspect `README.md`, `internal/model/model.go` (the Go wire contract), and `frontend/src/types/` (the TypeScript UI contract).

---

## 1. Product Mission & Core Invariants

NETWATCH is a **Local Network Intelligence & Device Inventory** desktop application for Windows, Linux, and macOS:  
**"Know Every Device on Your LAN. Without the Cloud Watching."**

### Beyond Ephemeral Scanning: The Asset Ledger
Traditional tools (*Advanced IP Scanner*, *Angry IP Scanner*) are ephemeral utilities that discard all data upon exit. In NETWATCH:
- **Scanning is merely the ingestion sensor**, not the end product.
- **The Core Product is the Local Asset Ledger & Reconciliation Engine**: An embedded SQLite database tracking the lifecycle of every network asset across hours, days, and months.
- **State Reconciliation**: Answers not just *"What is online now?"*, but *"What changed? Who joined? Who departed? When did an IP drift?"*
- **Flap Resistance**: Devices transition offline only after consecutive missed sweeps (`OfflineAfterMisses = 2`), preventing false alerts when low-power Wi-Fi devices sleep.

### Non-Negotiable Architectural Principles
1. **Zero External Egress**: Network topology, IP mappings, and MAC addresses never leave the host system. No accounts, no cloud relays, no telemetry, no analytics.
2. **Air-Gapped Hardware Fingerprinting**: MAC vendor lookups use an embedded, local IEEE OUI database (`internal/oui/ieee-oui.txt`). Hardware addresses are never sent to external APIs.
3. **Bundled Static Typography**: UI fonts (*Plus Jakarta Sans* and *JetBrains Mono*) are bundled locally via `@fontsource`. No Google Fonts or external CDNs.
4. **Authentic Data Only**: Unknown fields remain labeled `Unknown`. Never invent speculative threat scores, simulated health percentages, or hallucinated AI tags.
5. **Private IPv4 Scoping Only**: Scanning is restricted strictly to RFC 1918 private subnets, link-local, and CGNAT IPv4 targets.
6. **Sidebar Identity Constraint**: The Sidebar brand mark and header text (`NETWATCH / Desktop Edition`) in `frontend/src/components/layout/Sidebar.tsx` must remain untouched.

---

## 2. High-Level Architecture & Component Map

```
React 19 Desktop UI (frontend/)  ──HTTP + SSE──▶  Go Core Engine  ──▶  SQLite Database
   Wails / WebView2 Host                          127.0.0.1:<random>     %LOCALAPPDATA%\NetWatch\data\network.db
                                                  Bearer Token Auth
```

| Subsystem Path | Responsibility | Architectural Boundary |
|---|---|---|
| `internal/model/` | Canonical Go wire contracts | JSON field names strictly mirror `frontend/src/types/*.ts`. Timestamps are RFC 3339 UTC. |
| `internal/netenv/` | **All Operating System network access** | Hidden behind `Env` interface. `env_windows.go` (iphlpapi), `env_linux.go`, `env_other.go`. `Portable` holds cross-platform logic (WoL, TCP probe, rDNS). |
| `internal/engine/` | LAN discovery & intelligence | Pure Go logic over `Env` + `Store`. Subnet sweeps, state reconciliation/diff, event stream, ping, WoL, port probing, background auto-scan scheduler. |
| `internal/store/` | SQLite local ledger | WAL-mode persistence for network inventory, device history, event audit trail, and user settings. |
| `internal/fingerprint/`| Hardware classification | Conservative device-type heuristics from vendor OUI, hostname, and open port signatures. |
| `internal/oui/` | IEEE OUI database | Embedded raw vendor mappings. Air-gapped offline lookup. |
| `internal/api/` | REST + SSE Server | Binds strictly to `127.0.0.1:<random-port>`. Guarded by 192-bit Bearer token, Host validation (anti-DNS rebinding), and Origin allowlisting. |
| `internal/appdata/` | OS persistence paths | Manages `%LOCALAPPDATA%\NetWatch\data`, rotating log file (`netwatch.log`), and platform folder launchers. |
| `cmd/netwatchd/` | Headless daemon | Standalone CLI and service runner for development, headless environments, and Linux/macOS. |
| `main.go`, `app.go` | Wails Desktop Host | Native Windows window frame, menu bindings, system asset server, and lifecycle coordination. |
| `frontend/` | Desktop UI | Vite + React 19 + TypeScript + Tailwind CSS v4. Consumes engine exclusively via `NetworkService` interface. |

---

## 3. Go Core & Engine Best Practices

1. **Strict OS Isolation Boundary**:
   - Platform-dependent network calls (e.g., Windows `iphlpapi.dll` or Linux `/proc/net/arp`) MUST live exclusively within `internal/netenv/`.
   - The engine and API layers must remain 100% testable on any operating system using the simulated environment (`internal/engine/fake_test.go`).
2. **Concurrency & Thread Safety**:
   - Every background task (such as `runAutoScan(ctx)`) must accept a `context.Context` and terminate cleanly on cancellation.
   - Dynamic configuration triggers must use non-blocking channel notifications (`select { case ch <- struct{}{}: default: }`).
   - All state modifications must pass `go test -race ./...` without data races.
3. **SQLite Concurrency & WAL Stability**:
   - NETWATCH utilizes `github.com/ncruces/go-sqlite3` (v0.35.6+). Ensure driver versions avoid concurrency bugs in Windows shared-memory WAL mode.
   - Always verify query scans and propagate database errors explicitly; never silently swallow scan failures in `Store.Stats`.
4. **Triple-Layer Loopback API Security**:
   - Never bind outside `127.0.0.1`.
   - Never log the ephemeral Bearer token to stdout or log files.
   - Never relax Host or Origin verification headers in `internal/api/server.go`.

---

## 4. TypeScript & React 19 UI Best Practices

1. **Wire Contract Parity**:
   - Changing any model property requires simultaneous, synchronized updates across `internal/model/model.go`, `frontend/src/types/`, `frontend/src/services/HttpNetworkService.ts`, and corresponding unit tests.
2. **React 19 Performance**:
   - Utilize `useDeferredValue` for interactive text search and filtering across device and event tables to ensure 60 FPS input fluidity.
   - Keep page components decoupled from HTTP details: pages consume only `NetworkService` through `NetworkContext`.
3. **Quiet, Desktop-First Aesthetic**:
   - Follow desktop-first UI design (minimum viewport 1180×720).
   - Dense, high-information typography with zero decorative fluff. Avoid flashy SaaS neon, heavy glassmorphism, or oversized mobile padding.
   - Dark/Light mode is powered by Tailwind v4 CSS variables with explicit `@custom-variant dark (&:where(.dark, .dark *));`.
4. **Action Consolidation**:
   - Maintain clear separation between network sweeps (`Radar` icon / `Scan` / `Rescan`) and client-side text filtering (`Search` magnifying glass).

---

## 5. Wails v2 Desktop Bridge & Toolchain Guidelines

1. **Toolchain CLI Alignment**:
   - Keep the Go module `github.com/wailsapp/wails/v2` pinned to `v2.10.x` to match the locally installed Wails CLI (`wails v2.10.2`). Do not allow automated tools to bump it to untested major versions.
2. **Clean Checkout Compilation (Vite Embed Stub)**:
   - When Go embeds `frontend/dist` (`//go:embed all:frontend/dist`), a clean git clone lacks `dist/`.
   - Maintain `frontend/public/.gitkeep` (copied into `dist/` on build) AND ensure build scripts execute `mkdir -p frontend/dist && touch frontend/dist/.gitkeep` before Go toolchain invocation.
3. **Desktop Subsystem Build**:
   - Windows desktop production binaries must be compiled with GUI subsystem suppression to prevent background console windows:
     ```powershell
     go build -tags desktop,production -ldflags "-w -s -H windowsgui" -o build/bin/netwatch.exe .
     ```
4. **Platform Shell Launchers**:
   - System folder actions (`OpenDataFolder`) must use OS-abstracted implementations:
     - Windows: absolute `%WINDIR%\explorer.exe` execution with `syscall.SysProcAttr{HideWindow: true}`.
     - Linux/macOS: `xdg-open` or `open`.

---

## 6. Git & CI/CD Supply Chain Invariants

1. **Binary Resource Preservation (`.gitattributes`)**:
   - Whenever `* text eol=lf` is active, binary assets MUST be explicitly marked as `binary` in `.gitattributes`:
     ```gitattributes
     *.syso binary
     *.dll binary
     *.exe binary
     *.ico binary
     *.wasm binary
     ```
   - Failing to do so causes Git to alter CRLF line endings, corrupting COFF object headers (`fail to read string table length: unexpected EOF`).
2. **Supply Chain Security & SHA Pinning (`sec-01`)**:
   - All GitHub Actions in `.github/workflows/` must be pinned to full 40-character commit SHAs with minimum `permissions: contents: read`.
3. **Dependabot Governance (`.github/dependabot.yml`)**:
   - Enforce grouped updates (`go-dependencies` and `frontend-dependencies`).
   - Lock down breaking dependencies under `ignore`:
     - `github.com/wailsapp/wails/v2`
     - `@types/node` (major version)
     - `motion` (major version)
     - `lucide-react` (major version)
4. **GitHub Rulesets Governance**:
   - Branch `main`: Active ruleset blocking force-push, deletion, requiring linear history and strict green CI status checks (`Go Engine Tests (ubuntu-latest)`, `Go Engine Tests (windows-latest)`, `Frontend Typecheck & Build`).
   - Release tags `refs/tags/v*`: Protected against deletion and mutation.

---

## 7. SemVer & Release Lifecycle Management

### The Release Consistency Principle
A change is never complete if the project's versioning documentation is inconsistent.
$$\text{Code Implementation} + \text{Version Identifiers} + \text{CHANGELOG.md} + \text{ROADMAP.md} = \text{One Change Set}$$

1. **Semantic Versioning (SemVer 2.0.0)**:
   - **PATCH**: Backward-compatible bug fixes and internal maintenance.
   - **MINOR**: Backward-compatible new capabilities or endpoints.
   - **MAJOR**: Incompatible public API, contract, or architecture changes.
2. **Keep a Changelog (v1.1.0)**:
   - Stage active work under `## [Unreleased]`.
   - Use standard subheadings: `### Added`, `### Changed`, `### Fixed`.
   - Maintain release compare links at the bottom of `CHANGELOG.md`.
3. **Version Identifier Synchronization**:
   - The authoritative version must remain strictly identical across:
     - `main.go` (`var version`)
     - `cmd/netwatchd/main.go` (`var version`)
     - `frontend/package.json` (`"version"`)
     - `winres/winres.json` (`"ProductVersion"`)
     - Active GitHub Release tag (`v*`)

---

## 8. Development Loop & Current Milestone Status

### Local Development Loop
```powershell
# 1. Run headless daemon (prints local loopback URL & token)
go run ./cmd/netwatchd -origin http://localhost:3000

# 2. Run frontend Vite dev server (in frontend/)
cd frontend && npm install && npm run dev
# Open: http://localhost:3000/?api=<baseUrl>&token=<token>

# 3. Validation test suite
go test -race ./...
cd frontend && npx tsc --noEmit && npx vite build
```

### Current Roadmap Status
*For the complete milestone tracker, see `ROADMAP.md`.*

- [x] **M1 – Windows Native Validation**: Verified on real Windows LAN (`192.168.1.0/24`), unprivileged ICMP & ARP, rotating logger (`netwatch.log`), and standalone GUI build (`build/bin/netwatch.exe`).
- [x] **M2 – CI, Security & Governance Baseline**: Multi-OS CI with SHA-pinned actions, GitHub Rulesets on `main` and release tags, Dependabot grouped updates, STRIDE threat model (`THREAT_MODEL.md`), incident response playbook (`INCIDENT_RESPONSE.md`), and unit test suites for API loopback security and Store persistence.
- [x] **M3 – Real Settings & Database Management (Completed)**:
  - [x] SQLite settings persistence (`GET/PUT /v1/settings`).
  - [x] Background auto-scan engine scheduler (`1m`, `5m`, `15m`, `1h`, `manual`).
  - [x] Real-time SQLite statistics card, VACUUM compaction, and integrity checks.
  - [x] Native Windows toast notifications for new devices, offline events, and network changes (`internal/notifier`).
  - [x] System startup integration: launch at startup via HKCU Run key (`internal/autostart`), `--minimized` launch flag, and single-instance restoration.
  - [x] Event history retention and pruning (`POST /v1/data/prune`).
- [x] **M4 – Multi-Protocol Discovery, Evidence & Trust (Completed)**: Randomized MAC detection (`isLocallyAdministered`), "Private MAC" badge, device merge & aliasing (`POST /v1/devices/{id}/merge`), trust allowlist (`known`/`guest`/`unknown`), unprivileged mDNS, SSDP & NetBIOS probers, untrusted hostname sanitization, and automated Zero-Egress CI gate.
- [x] **M5 – Watchdog, Enhanced Discovery & Device Actions (In Progress)**: IPv6 NDP Neighbor Table discovery (`GetIpNetTable2`), Windows Public Network firewall alert, WS-Discovery protocol (`UDP:3702`), "Open Device" quick actions (Web UI/SSH/RDP). Remaining: ARP conflict alerts, and latency ledger.
- [ ] **M6 – Packaging, Cryptographic Integrity & Distribution**: Unsigned binary distribution, automated SHA-256 checksums (`SHA256SUMS.txt`), Windows Defender SmartScreen guidance, reproducible build verification, MSI/NSIS installer, and SBOM.
- [ ] **M7 – Internationalization (i18n), Diff Reports & Public v1.0 (In Progress)**: Air-gapped i18n engine with full Arabic RTL translation completed. Remaining: additional locales, historical snapshot diffs, safe CSV export, and Playwright E2E smoke suite.

---

## 9. Handoff Checklist for Every PR / Task
Before declaring any task complete, verify:
1. `go test -race ./...` passes without errors or data races.
2. `cd frontend && npm run build` compiles with zero TypeScript or Vite errors.
3. Any version-affecting change updates `CHANGELOG.md`, `ROADMAP.md`, and version variables atomically.
4. Working tree is clean (`git status`).
