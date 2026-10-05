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

### Fixed
- **Windows ARP Table Desynchronization**:
  - Corrected `ipNetRowSize` in `internal/netenv/env_windows.go` from `28` to `24` bytes (`sizeof(MIB_IPNETROW)` on Windows). Previously, the 4-byte offset drift per row corrupted row parsing, causing the Default Gateway (`192.168.1.1`) and most neighboring devices to be skipped.
- **Network Topology Router Inspection**:
  - Updated `frontend/src/pages/Network.tsx` to match the actual gateway device by IP and type dynamically, and eliminated the static mock TP-Link fallback that caused the "No device selected" screen upon clicking.
- Fixed missing imports in `cmd/netwatchd/main.go` (`os/exec`, `syscall`) and `main.go` (`netwatch/internal/appdata`).
- Cleaned up proxy mirror `replace` directives in `go.mod` for direct, clean module resolution on standard Windows environments.
