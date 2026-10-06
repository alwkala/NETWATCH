# NETWATCH — Super Fast Network Scanner & Device Inventory

<div align="center">

### Privacy-First Local Network Intelligence & Discovery Utility

[![CI](https://github.com/alwkala/NETWATCH/actions/workflows/ci.yml/badge.svg)](https://github.com/alwkala/NETWATCH/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE-MIT)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-green.svg)](LICENSE-APACHE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8.svg?logo=go)](go.mod)
[![React](https://img.shields.io/badge/React-19-61DAFB.svg?logo=react)](frontend/package.json)
[![Privacy](https://img.shields.io/badge/Privacy-100%25_Local--First-success.svg)](#privacy-invariants)

<a href="#downloads--platform-support"><b>Download</b></a> •
<a href="#why-netwatch-comparison-matrix"><b>Why NETWATCH?</b></a> •
<a href="ROADMAP.md"><b>Roadmap</b></a> •
<a href="THREAT_MODEL.md"><b>Threat Model</b></a> •
<a href="CONTRIBUTING.md"><b>Contributing</b></a> •
<a href="CHANGELOG.md"><b>Changelog</b></a>

<br>

> ## Know Every Device on Your LAN. Without the Cloud Watching.
> **NETWATCH sweeps your subnet in milliseconds, fingerprints hardware locally, and maintains an unshakeable device inventory. 100% offline, zero telemetry, no account required.**
>
> *Instant LAN discovery and persistent device tracking — 100% offline, zero cloud, zero telemetry.*

</div>

---

## Why NETWATCH? (Comparison Matrix)

| Capability | NETWATCH | Fing Desktop | Advanced IP Scanner | Sniffnet |
|:---|:---:|:---:|:---:|:---:|
| **Local-First (No Account / No Cloud)** | ✅ **100% Offline** | ❌ Requires Cloud Account | ⚠️ Local but closed-source | ✅ 100% Offline |
| **Persistent Device Inventory & Timeline** | ✅ **Yes (SQLite)** | ✅ Yes (Cloud sync) | ❌ Transient list only | ❌ Session traffic only |
| **Subnet Sweep Speed** | ⚡ **Milliseconds** | ⚠️ Moderate | ⚡ Fast | N/A (Traffic capture) |
| **Flap-Resistant Reconciliation** | ✅ **Yes (`misses=2`)** | ⚠️ Partial | ❌ No | N/A |
| **Modern Desktop UI** | ✅ **React 19 + Wails** | ⚠️ Electron / Paywalled | ❌ 2010 Win32 Legacy UI | ✅ Modern Iced GUI |
| **Unprivileged Execution (No Root/Admin)** | ✅ **Yes (`iphlpapi.dll`)** | ⚠️ Requires Npcap/Admin | ⚠️ Requires UAC / Npcap | ⚠️ Requires WinPcap/Npcap |
| **Open Source & Dual Licensed** | ✅ **MIT / Apache 2.0** | ❌ Proprietary | ❌ Proprietary | ✅ MIT / Apache 2.0 |

---

## Architecture Overview

```
React 19 UI  ──HTTP + SSE──▶  Go Core Engine (127.0.0.1:random, Bearer Token)  ──▶  SQLite Store
 (Wails/WebView2)              discovery · scan · diff/events · fingerprint           (%LOCALAPPDATA%\NetWatch\data)
```

- **`internal/netenv`**: Operating system network access abstracted behind an `Env` interface:
  - **Windows**: Unprivileged native `iphlpapi.dll` (`IcmpSendEcho`, `GetIpNetTable`, `SendARP`, `GetAdaptersAddresses`).
  - **Linux**: Netlink and `/proc/net/arp` socket implementation for development and daemon mode.
  - **Portable**: RFC-compliant Wake-on-LAN (WoL), TCP port probing, reverse DNS.
- **`internal/engine`**: Active discovery engine with concurrent subnet sweeps, reconciliation diff, and offline detection (`OfflineAfterMisses = 2` consecutive missed scans to prevent sleep-mode flapping).
- **`internal/store`**: Embedded SQLite database persisting device inventory, state changes, and historical event logs.
- **`internal/api`**: REST and Server-Sent Events (SSE) server bound strictly to loopback (`127.0.0.1`) with per-session token authorization and Host/Origin validation.
- **`frontend/`**: Modern React 19 + TypeScript + Vite + Tailwind CSS desktop interface with bundled air-gapped typography (*Plus Jakarta Sans* & *JetBrains Mono*).

---

## Key Features

- ⚡ **Super Fast Subnet Discovery**: Active, unprivileged ARP and ICMP sweeps across private IPv4 LANs in milliseconds.
- 🏷️ **Local Hardware Fingerprinting**: Local MAC OUI lookup via embedded IEEE database (zero external network requests).
- ⏱️ **Flap-Resistant Status Tracking**: Smart reconciliation marks devices offline only after two consecutive missed sweeps.
- 🔌 **Integrated Network Diagnostics**: Unprivileged Ping with latency measurement, TCP port probing, and Wake-on-LAN (WoL) transmission.
- 📜 **Historical Network Timeline**: Persistent SQLite log tracking when devices join, leave, or change IP addresses.
- 🛡️ **100% Offline & Air-Gapped**: Zero cloud accounts, zero telemetry, and bundled fonts with zero external CDN dependencies.

---

## Downloads & Platform Support

| Operating System | Architecture | Package Format | Status |
|:---|:---|:---|:---|
| **Windows 10 / 11** | `x64` | Portable `.exe` / Installer | **Validated & Active** ([Build Instructions](#build-the-desktop-app-windows)) |
| **Linux** | `amd64`, `arm64` | Daemon / CLI (`netwatchd`) | **Development / Headless** |
| **macOS** | `Apple Silicon`, `Intel` | Wails `.app` / `.dmg` | **Roadmap (M7)** |

---

## Privacy & Security Invariants

- **No Remote Telemetry**: Network configurations, IP mappings, and MAC addresses never leave the host system.
- **Loopback Isolation**: The local API binds exclusively to `127.0.0.1` on an ephemeral port. Web pages open in local browsers cannot access the engine.
- **Authentic Data Only**: Unknown fields remain labeled `Unknown`. NETWATCH never invents speculative health scores or simulated threat indicators.
- **Formal Threat Model**: Full STRIDE analysis and mitigations documented in [THREAT_MODEL.md](THREAT_MODEL.md).

---

## Quick Start & Development

### Prerequisites
- Go 1.24+
- Node.js 20+ & npm

```bash
# Terminal 1: Run the Go headless daemon (prints local URL & bearer token)
go run ./cmd/netwatchd -origin http://localhost:3000

# Terminal 2: Run the frontend development server
cd frontend
npm install
npm run dev
# Open the printed URL with ?api=<baseUrl>&token=<token>
```

### Build the Desktop App (Windows)

```bash
go build -tags desktop,production -ldflags "-w -s -H windowsgui" -o build/bin/netwatch.exe .
```

* Data folder: `%LOCALAPPDATA%\NetWatch\data\network.db`
* Rotating log: `%LOCALAPPDATA%\NetWatch\data\netwatch.log` (5 MB cap with 3 backup rotations)

---

## Troubleshooting

<details>
<summary><b>Click to expand common questions & diagnostics</b></summary>

### 1. Gateway or devices missing during scan
Ensure your network profile is set to **Private Network** in Windows Settings. Certain third-party firewalls block ICMP Echo replies. NETWATCH relies on both ARP and ICMP for discovery.

### 2. Viewing local application logs
You can inspect live logs generated by the engine at:
```powershell
Get-Content -Wait $env:LOCALAPPDATA\NetWatch\data\netwatch.log
```
Or click **Open Data Folder** directly inside the Settings page in NETWATCH.

### 3. Running tests locally
```bash
# Go unit tests & race detector
go test -race ./...

# Frontend typecheck & build
cd frontend && npm run lint && npm run build
```

</details>

---

## Roadmap & Governance

- [ROADMAP.md](ROADMAP.md) — Public milestones and technical progress (M1–M7).
- [THREAT_MODEL.md](THREAT_MODEL.md) — STRIDE security model and boundary definitions.
- [INCIDENT_RESPONSE.md](INCIDENT_RESPONSE.md) — Security incident response protocols.
- [CONTRIBUTING.md](CONTRIBUTING.md) — Guidelines for code, documentation, and design contributions.
- [SECURITY.md](SECURITY.md) — Vulnerability disclosure policy and SLAs.
- [CONTRIBUTORS.md](CONTRIBUTORS.md) — Core team and contributor acknowledgments.

---

## License

NETWATCH is dual-licensed under both:
- **[MIT License](LICENSE-MIT)**
- **[Apache License 2.0](LICENSE-APACHE)**

You may choose either license at your option.
