# AGENTS.md — NETWATCH handoff for AI coding agents

Read this file fully before changing anything. Then read `README.md`, `internal/model/model.go`
(the wire contract) and `frontend/src/services/NetworkService.ts` (the UI contract).

## 1. What this product is
NETWATCH is a **local-first, privacy-first Windows desktop network utility** (Fing-like):
discover devices on the LAN, keep an inventory, run diagnostics, record history/events.

Non-negotiable principles:
- Network data never leaves the machine. No account, cloud, telemetry, analytics, or external API calls.
  The UI must make **zero external requests** (fonts are bundled; keep it that way — no CDN, no Google Fonts).
- OUI/vendor lookup is local (`internal/oui/ieee-oui.txt`, embedded). Never send MAC addresses anywhere.
- Never invent data: unknown stays "Unknown"; no fake health scores, threat detection or AI features.
  Unimplemented features are hidden or labelled "Coming later".
- Scanning is active (ARP/ICMP/TCP on the LAN). Only private/link-local/CGNAT IPv4 targets are allowed.

## 2. Architecture
```
React UI (frontend/)  ──HTTP + SSE──▶  Go engine  ──▶  SQLite (store)
  Wails/WebView2 host                  127.0.0.1:<random>, bearer token per session
```
| Path | Role |
|---|---|
| `internal/model` | Wire types. JSON names mirror `frontend/src/types/*.ts`; timestamps are RFC 3339 UTC |
| `internal/netenv` | **All OS access** behind the `Env` interface. `env_windows.go` (iphlpapi/ARP/ICMP/adapters), `env_linux.go` (dev), `env_other.go`. `Portable` holds cross-platform parts (WoL, TCP probe, rDNS) |
| `internal/engine` | Scan pipeline, reconciliation/diff, events, monitor, ping, WoL, port probe. Pure logic over `Env` + `Store` |
| `internal/store` | SQLite persistence (inventory, history, events, meta) |
| `internal/fingerprint` | Conservative device-type classification + naming from vendor/hostname/ports |
| `internal/oui` | Embedded IEEE OUI database |
| `internal/api` | REST + SSE. Security: loopback only, Host check, Origin allow-list, bearer token |
| `internal/appdata` | Data dir (`%LOCALAPPDATA%\NetWatch\data\network.db`) |
| `cmd/netwatchd` | Headless engine for development/CLI/service use |
| `main.go`, `app.go` | Wails host (build tag `windows`); `main_other.go` is a stub elsewhere |
| `frontend/` | Vite + React + TS + Tailwind. `HttpNetworkService` (real) / `MockNetworkService` (prototype) chosen in `services/createNetworkService.ts` |

Key behaviours already implemented (do not regress): per-network inventory keyed by subnet+gateway MAC;
a device goes offline only after `OfflineAfterMisses` (2) consecutive missed scans; new devices stay
`isNew` until acknowledged (`PATCH /v1/devices/{id}` `{isNew:false}`); scans stream progress over SSE;
only one scan runs at a time.

API (all under `/v1`, bearer token except `health`): `GET network, devices, devices/{id}, devices/{id}/history, events,
scans/{id}, scans/{id}/stream` · `PATCH devices/{id}` · `POST scans, ping, wol, devices/{id}/ports, data/open` · `DELETE data`.

Frontend conversion: the engine sends ISO times; `HttpNetworkService` formats them to the strings the
UI expects (`Now`, `5m ago`, `HH:mm`) and adds `lastSeenAt` for sorting. Keep the UI independent of the
engine: pages talk only to `NetworkService` via `NetworkContext`.

## 3. Rules for working in this repo
1. **Contract first.** Changing a wire shape = update `internal/model`, `frontend/src/types`, the adapter, and the tests together.
2. **OS code only in `internal/netenv`.** Engine/API must stay testable on Linux with the fake env (`internal/engine/fake_test.go`).
3. Every engine behaviour change needs a test in `internal/engine`. Run `go test -race ./...` before finishing.
4. Frontend: `cd frontend && npx tsc --noEmit && npx vite build` must pass. Keep the design system in `docs`-style
   spec: dense, quiet, desktop-first (min 1180×720), no emoji icons, no neon/glass/SaaS look.
5. Windows code cannot run on Linux. Verify with `GOOS=windows go build ./... && GOOS=windows go vet ./...`, and state clearly
   what was **not** run on real Windows. Never claim hardware behaviour you did not observe.
6. Don't commit secrets/tokens; don't log the session token; never widen the API beyond 127.0.0.1 or relax Host/Origin checks.
7. Prefer small, reviewable commits. Don't rewrite working modules; extend them.

Build notes: `golang.org/x/*` may need `replace` to GitHub mirrors in restricted networks (see `go.mod`);
on a normal machine run `go mod tidy`. Desktop build: `wails build` (Wails v2.10.x).

## 4. Dev loop
```
go run ./cmd/netwatchd -origin http://localhost:3000      # prints {baseUrl, token}
cd frontend && npm install && npm run dev                  # open /?api=<baseUrl>&token=<token>
go test -race ./...    &&   (cd frontend && npx tsc --noEmit)
```

## 5. Status and honest gaps
Done: engine v0.1 (discovery, scan quick/full, diff/events, SQLite, ping, WoL, port probe), API, Wails host,
UI wired to the engine, security checks, simulated-network tests.
**Verified on real Windows**: iphlpapi ARP (`GetIpNetTable`), unprivileged ICMP (`IcmpSendEcho`), adapter and DNS discovery (`GetAdaptersAddresses`), live LAN scan (`192.168.1.0/24`), rotating file logging (`%LOCALAPPDATA%\NetWatch\data\netwatch.log`), and standalone GUI build (`build/bin/netwatch.exe`).
Known gaps: SSID (UI falls back to interface name), mDNS/NetBIOS/SSDP names, DHCP info, settings are UI-only
(scan interval, notifications, tray, start-with-Windows are not persisted or applied), unit tests for api/store/netenv helpers,
installer/signing, icon/branding.

## 6. Roadmap (do in this order; one milestone per PR)
- [x] **M1 – Validate on Windows**: verified Go engine on real LAN, iphlpapi ARP & ICMP echo, adapter discovery, Open Data Folder wired, built standalone Windows executable (`build/bin/netwatch.exe`), and added rotating file logging (`netwatch.log`).
- [ ] **M2 – Tests & CI**: GitHub Actions established for `go test -race` (Ubuntu), `go test` + `GOOS=windows go build` (Windows), and frontend typecheck/build; Dependabot and OSS governance files configured; pending unit tests for `store`, `api` (auth, Host/Origin, SSE), `netenv` parsing helpers, and Playwright smoke against `netwatchd` with `-demo`.
**M3 – Real settings**: persist settings in SQLite (`GET/PUT /v1/settings`); scheduled auto-scan with the chosen interval;
Windows toast notifications for new device / offline / network change; tray + start minimized + launch at startup (HKCU Run key).
**M4 – Better identity**: mDNS, NetBIOS, SSDP/UPnP, DHCP lease info; SSID via WLAN API; improved classification with confidence
and an explicit "Unknown" fallback; user-editable device type; full OUI refresh script (offline file, no runtime download).
**M5 – Diagnostics**: traceroute, DNS lookup/reverse, latency history charts per device, gateway/DNS/internet health from real probes only.
**M6 – Data**: export JSON/CSV, retention policy, DB migrations framework with versioned schema, backup/restore.
**M7 – Release**: MSI/NSIS installer, code signing, auto-update **opt-in only**, privacy statement matching actual behaviour.
Later: SNMP, topology from real LLDP/ARP data, multi-network profiles, CLI (`netwatch discover|scan|export`) on top of the engine.

## 7. When you finish a task
Report: what changed, what you ran (exact commands + results), what you could **not** verify (especially Windows),
and the next milestone item. Update this file's section 5/6 if status changed.
