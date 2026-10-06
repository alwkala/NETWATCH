# Technical Architecture

NETWATCH pairs a compiled Go network core with a high-performance React 19 desktop interface via Wails v2:

```
React 19 Desktop UI (frontend/)  ──HTTP + SSE──▶  Go Core Engine  ──▶  SQLite Database
   Wails / WebView2 Host                          127.0.0.1:<random>     %LOCALAPPDATA%\NetWatch\data\network.db
                                                  Bearer Token Auth
```

---

## Subsystem Boundary Map

| Subsystem Path | Responsibility | Architectural Boundary |
|---|---|---|
| `internal/model/` | Canonical Go wire contracts | JSON field names strictly mirror `frontend/src/types/*.ts`. Timestamps are RFC 3339 UTC. |
| `internal/netenv/` | All OS network access | Hidden behind `Env` interface. `env_windows.go` (`iphlpapi.dll`), `env_linux.go`. Cross-platform logic in `portable.go` (WoL, TCP probe, rDNS) and `wan.go` (STUN). |
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

## Concurrency & Thread Safety

- **Atomic Inventory Commit**: Serialized under `commitMu` in `Engine` to prevent SQLite lock contention.
- **Background Tasks**: Background tasks (`runAutoScan`, `runMonitor`) accept `context.Context` and shut down cleanly on process termination.
- **Port Scan Protection**: Concurrency guarded by `sync.Map` to prevent duplicate parallel port probe storms against the same device.
