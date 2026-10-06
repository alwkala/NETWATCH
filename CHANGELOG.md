# Changelog

All notable changes to the NETWATCH project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0-dev] - 2026-10-06

### Added
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
  - Redesigned `README.md` with visual architecture badges, download matrix, and collapsible diagnostics.
  - Updated repository topics and description on GitHub via `gh` CLI to reflect cross-platform architecture.

### Fixed
- **Windows ARP Table Desynchronization**:
  - Corrected `ipNetRowSize` in `internal/netenv/env_windows.go` from `28` to `24` bytes (`sizeof(MIB_IPNETROW)` on Windows). Previously, the 4-byte offset drift per row corrupted row parsing, causing the Default Gateway (`192.168.1.1`) and most neighboring devices to be skipped.
- **Network Topology Router Inspection**:
  - Updated `frontend/src/pages/Network.tsx` to match the actual gateway device by IP and type dynamically, and eliminated the static mock TP-Link fallback that caused the "No device selected" screen upon clicking.
- Fixed missing imports in `cmd/netwatchd/main.go` (`os/exec`, `syscall`) and `main.go` (`netwatch/internal/appdata`).
- Cleaned up proxy mirror `replace` directives in `go.mod` for direct, clean module resolution on standard Windows environments.
