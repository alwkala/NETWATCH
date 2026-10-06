# NETWATCH Technical Architecture & Engineering Guide

This document describes the internal engineering design, subsystem boundaries, data flow, and runtime security model of **NETWATCH**.

---

## 1. High-Level Architecture & Component Map

```
React 19 Desktop UI (frontend/)  ──HTTP + SSE──▶  Go Core Engine  ──▶  SQLite Database
   Wails / WebView2 Host                          127.0.0.1:<random>     %LOCALAPPDATA%\NetWatch\data\network.db
                                                  Bearer Token Auth
```

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

## 2. Ingestion & Sensing (Layer 2/3)

NETWATCH is engineered to operate strictly in **unprivileged user-mode**, eliminating the need for Administrator rights, UAC elevation, or third-party packet drivers (such as Npcap or WinPcap).

### Windows Implementation (`internal/netenv/env_windows.go`)
- **`GetAdaptersAddresses`**: Queries network interfaces, operational status, IPv4 netmasks, DNS servers, and default gateways.
- **`GetIpNetTable`**: Reads the OS kernel ARP cache populated by normal network activity.
- **`IcmpSendEcho`**: Sends unprivileged ICMP echo requests to sweep RFC 1918 subnets, proactively populating the OS ARP cache for quiet devices.
- **`SendARP`**: Directly resolves single-IP hardware addresses when a host does not respond to ICMP.

### Multi-Protocol Evidence Bus
- **NetBIOS Name Service (NBNS — UDP 137)**: Directed queries extracting machine names and the 6-byte Unit ID (hardware MAC).
- **SSDP / UPnP (Multicast UDP 1900)**: Observes UPnP device advertisements (`239.255.255.250`). Adheres to the **Zero-Fetch Invariant**: records `Location` headers as diagnostic text without fetching HTTP XML.
- **mDNS / DNS-SD (Multicast UDP 5353)**: Passively listens to `224.0.0.251` for `.local` services with loop protection against DNS compression pointer exploits.
- **RFC 5389 STUN (`internal/netenv/wan.go`)**: Queries Cloudflare/Google STUN over raw UDP to discover the external public IPv4 address without external HTTP egress.

---

## 3. The Local Asset Ledger & Reconciliation Engine

Unlike ephemeral scanners that wipe data on exit, NETWATCH persists all discoveries in an embedded SQLite database using Write-Ahead Logging (WAL mode).

### Flap-Resistant Reconciliation (`internal/engine/reconcile.go`)
- **Threshold Rule**: Devices transition offline only after **consecutive missed sweeps** (`OfflineAfterMisses = 2`) and a minimum elapsed duration (`MinOfflineDuration = 5m`).
- **Rapid Wake**: Any observed packet immediately restores the device to `online`.

### Identity Unification & Random MACs
- **LAA Detection**: Identifies Locally Administered Addresses (`mac[0] & 0x02 != 0`) used by iOS, Android, and Windows 10/11 Private Wi-Fi.
- **Device Merge Workflow**: `POST /v1/devices/{targetId}/merge` combines rotating MACs into a canonical asset history, storing aliases in `device_mac_aliases`.
- **Customization Permanence (Schema v4)**: User-defined names, device types, and icons are saved in `custom_type` and protected from automated scanner overwrites.

---

## 4. Triple-Layer Security Model

NETWATCH follows an air-gapped, zero-egress security architecture:

1. **Zero External Egress**:
   - Hardware addresses and subnets never leave the host.
   - Enforced by automated CI gate `TestZeroEgressInvariant` verifying no `net/http` client dialers exist in the engine.
2. **Loopback-Only Binding**:
   - The internal HTTP/SSE daemon binds strictly to `127.0.0.1:<random-port>`.
   - Each startup generates an unguessable 192-bit cryptographic Bearer token (`crypto/rand`).
3. **Anti-DNS Rebinding & Origin Protection**:
   - Incoming requests validate the `Host` header against `127.0.0.1` and `localhost`.
   - Browser tabs on the local machine cannot cross-origin interact with the daemon.

---

## 5. Development & Build Toolchain

### Prerequisites
- Go 1.24+
- Node.js 20+ & npm

### Development Mode
```powershell
# 1. Start headless Go daemon
go run ./cmd/netwatchd -origin http://localhost:3000

# 2. Start React frontend
cd frontend
npm install
npm run dev
# Open: http://localhost:3000/?api=<baseUrl>&token=<token>
```

### Production Build (Windows GUI)
```powershell
# 1. Build frontend bundle
cd frontend && npm run build && cd ..

# 2. Compile standalone GUI executable (no console window)
go build -tags desktop,production -ldflags "-w -s -H windowsgui" -o build/bin/netwatch.exe .
```
