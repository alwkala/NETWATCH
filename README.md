# NETWATCH

<div align="center">

### Know Every Device on Your Network. Without the Cloud Watching.

**A private, local-first network intelligence and asset ledger desktop app for Windows, Linux, and macOS.**

[![Release](https://img.shields.io/github/v/release/alwkala/NETWATCH?include_prereleases&color=blue&label=Latest%20Release)](https://github.com/alwkala/NETWATCH/releases/tag/v0.2.0-alpha.1)
[![CI](https://github.com/alwkala/NETWATCH/actions/workflows/ci.yml/badge.svg)](https://github.com/alwkala/NETWATCH/actions/workflows/ci.yml)
[![License: MIT & Apache 2.0](https://img.shields.io/badge/License-MIT%20%2F%20Apache%202.0-blue.svg)](LICENSE-MIT)
[![Privacy Guarantee](https://img.shields.io/badge/Privacy-100%25_Local--First-emerald.svg)](#privacy-guarantee)
[![Portable App](https://img.shields.io/badge/Windows-Portable_No_Install-blue.svg)](#quick-start)

<br>

<img src="docs/Snapshots/home-desktop-dark-theme.png" alt="NETWATCH Dashboard" width="100%" style="border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.3);" />

<br><br>

<p align="center">
  <a href="https://github.com/alwkala/NETWATCH/releases/tag/v0.2.0-alpha.1"><b>⬇️ Download Windows (.exe)</b></a> •
  <a href="#see-it-in-action"><b>📸 Screenshots & Tour</b></a> •
  <a href="#how-it-works"><b>⚡ How It Works</b></a> •
  <a href="docs/ARCHITECTURE.md"><b>🛠️ Technical Architecture</b></a> •
  <a href="https://github.com/alwkala/NETWATCH/wiki"><b>🌐 GitHub Wiki</b></a>
</p>

</div>

---

## Why NETWATCH?

> *"Scanning tells you what's there now.  
> Inventory tells you what changed."*

When you connect to Wi-Fi at home, in the office, or at a client site, you often want answers to simple, critical questions:
- *What devices are connected right now?*
- *What is that mystery IP address on my subnet?*
- *Who joined the network today? Did my server reboot?*
- *Did my phone get assigned a new IP address?*

Traditional tools (like *Advanced IP Scanner* or *Angry IP Scanner*) are momentary utilities: they perform a single sweep, show a plain table, and **discard everything the moment you close them**.

**NETWATCH is different.**  
Scanning is just the ingestion sensor. Behind the scenes, NETWATCH maintains a **persistent local SQLite asset ledger**. It remembers every host that ever touched your subnet, tracks joins and departures, alerts you to IP changes, and lets you organize your network into a clean, searchable inventory — **100% offline, with zero cloud and zero telemetry.**

---

## Key Capabilities

### ⚡ 1. Instant Subnet Discovery (No Admin Needed)
Discover every active computer, phone, printer, smart TV, and IoT sensor on your `/24` subnet in seconds.  
NETWATCH uses unprivileged user-mode operating system APIs — **no Administrator rights, no UAC popups, and no third-party packet capture drivers (WinPcap/Npcap) required**.

### 🏷️ 2. Air-Gapped Hardware Fingerprinting
Instantly identifies manufacturers (Apple, Intel, Samsung, Espressif, Raspberry Pi, ZTE) using an **embedded, offline IEEE OUI database**. Your hardware MAC addresses are never sent to external lookup APIs.

### 🎨 3. Personalized Device Inventory & Icons
Give your hardware friendly names (e.g., *"Living Room Apple TV"*, *"Proxmox Lab 01"*), select custom device categories, and choose matching icons (`Computer`, `Phone`, `Server`, `Router`, `IoT`, `Camera`, `Printer`, `Game Console`). Your customizations are saved permanently and are never overwritten by rescans.

### 🛡️ 4. Private MAC Detection & Device Merging
Modern smartphones (iOS, Android) and Windows 10/11 rotate their Wi-Fi MAC addresses for privacy. NETWATCH automatically flags **"Private MAC"** addresses and provides a **1-click Device Merge tool** to unify fragmented device histories under one record.

### 🤝 5. Three-Tier Trust Management
Organize your devices into trust levels:
- **`Known`**: Approved assets (your workstations, family phones, home servers).
- **`Guest`**: Temporary visitors.
- **`Unknown`**: Newly discovered or unrecognized hardware that needs inspection.

### 🌐 6. Visual Network Topology & Public IP
Inspect your local network hierarchy: WAN Internet gateway, default router, broadcast domain, and connected endpoints. Automatically resolves your network's external public IP over lightweight STUN (UDP) without cloud tracking.

### ⏱️ 7. Flap-Resistant State Tracking
Low-power Wi-Fi devices sleep frequently to save battery. NETWATCH only marks a device offline after consecutive missed sweeps, preventing annoying false disconnect alarms.

### 🔌 8. Built-In Network Diagnostics
- **Live ICMP Ping**: Real-time round-trip latency graph and packet loss measurement.
- **Wake-on-LAN (WoL)**: Send magic broadcast packets to wake sleeping PCs on your LAN.
- **Port Scanner**: Check open TCP service ports (HTTP, SSH, SMB, RDP, RTSP).
- **Desktop Toast Notifications**: Optional Windows notifications when unknown devices appear or gateways change.

---

## Who Is NETWATCH For?

| Role | How NETWATCH Helps You |
|---|---|
| 🏠 **Home Lab & Self-Hosters** | Track your Raspberry Pis, NAS drives, Proxmox clusters, and ESP32 home automation sensors without setting up heavy enterprise agents. |
| 💻 **DevOps & Remote Workers** | Instantly audit client networks or home office LANs. Verify IP allocations, test gateway latency, and check open service ports. |
| 🛡️ **Privacy Advocates** | Audit every device in your home without trusting third-party cloud scanners or sending your home network topology to remote servers. |
| 🏢 **Small Office & IT Techs** | Know immediately when an unrecognized laptop plugs into the office switch, spot IP conflicts, and maintain an up-to-date asset ledger. |

---

## See It in Action

<table width="100%">
  <tr>
    <td width="50%" valign="top">
      <h4 align="center">Real-Time Asset Ledger</h4>
      <img src="docs/Snapshots/home-desktop-dark-theme.png" alt="Device Ledger" width="100%" />
      <p align="center"><em>Searchable list of all network devices with latency sparklines, vendors, and status indicators.</em></p>
    </td>
    <td width="50%" valign="top">
      <h4 align="center">Subnet & Network Topology</h4>
      <img src="docs/Snapshots/nodes-desktop-dark-theme.png" alt="Topology View" width="100%" />
      <p align="center"><em>Visual physical and logical network hierarchy showing WAN Public IP, router, and connected endpoints.</em></p>
    </td>
  </tr>
  <tr>
    <td width="50%" valign="top">
      <h4 align="center">Device Dossier & Customization</h4>
      <img src="docs/Snapshots/settings-desktop-dark-theme.png" alt="Device Inspector" width="100%" />
      <p align="center"><em>Customize friendly names, device types, icons, trust status, and inspect multi-protocol evidence (mDNS, SSDP, NBNS).</em></p>
    </td>
    <td width="50%" valign="top">
      <h4 align="center">Activity Timeline & Audit Trail</h4>
      <img src="docs/Snapshots/logs-desktop-dark-theme.png" alt="Activity Logs" width="100%" />
      <p align="center"><em>Chronological timeline recording when devices join, depart, change IP, or merge records.</em></p>
    </td>
  </tr>
</table>

<div align="center">
  <p><strong>Mobile-First Responsive Design</strong>: Seamlessly inspect your network from phones, tablets, or narrow laptop viewports.</p>
  <img src="docs/Snapshots/home-mobile-full.png" alt="NETWATCH Mobile View" width="320px" style="border-radius: 8px; border: 1px solid #444;" />
</div>

---

## Quick Start

### Windows (Recommended)
1. Download **`netwatch.exe`** from the [Latest Release](https://github.com/alwkala/NETWATCH/releases/tag/v0.2.0-alpha.1).
2. *(Optional & Recommended)* Verify the SHA-256 cryptographic hash against the release notes:
   ```powershell
   Get-FileHash .\netwatch.exe -Algorithm SHA256
   ```
3. Double-click to run:
   - **No installation needed** (portable single-file executable).
   - **No administrator elevation required** (unprivileged user mode).
   - **Unsigned binary model**: On initial launch, Windows SmartScreen may display *"Unknown Publisher / Windows protected your PC"*. Click **"More info"** -> **"Run anyway"**.
   - Your data is stored locally in `%LOCALAPPDATA%\NetWatch\data\network.db`.

### Linux & Headless Servers
For headless machines, continuous background monitoring, or Linux servers:
```bash
# Clone the repository
git clone https://github.com/alwkala/NETWATCH.git && cd NETWATCH

# Run the standalone headless daemon
go run ./cmd/netwatchd
```

---

## Privacy Guarantee

NETWATCH was built around a non-negotiable philosophy: **"Your network data belongs to you."**

- **Zero Cloud Uploads**: Network mappings, IP addresses, and MAC addresses never leave your machine.
- **Zero Telemetry**: No analytics, no tracking beacons, no accounts.
- **Local SQLite Storage**: All history and settings are stored in an open SQLite database on your local disk.
- **Air-Gapped Typography**: All fonts (*Plus Jakarta Sans* & *JetBrains Mono*) are bundled inside the app. No calls to Google Fonts or remote CDNs.
- **Automated CI Security Gate**: An automated CI test inspects every commit to ensure no external HTTP requests can be introduced.

---

## Documentation & Deep Dives

Looking for deep technical architecture, threat models, or contributor runbooks?

- [Technical Architecture Guide](docs/ARCHITECTURE.md) — Comprehensive guide to the Go engine, React 19 UI, and IPC.
- [GitHub Wiki](https://github.com/alwkala/NETWATCH/wiki) — Full wiki documentation, subsystem deep dives, and operational runbooks.
- [Public Roadmap (M1–M7)](ROADMAP.md) — Milestone progress and future capabilities.
- [Threat Model (STRIDE)](THREAT_MODEL.md) — Security boundaries, trust domains, and threat analysis.
- [Release Changelog](CHANGELOG.md) — Detailed historical release notes following Keep a Changelog.
- [Security Policy](SECURITY.md) — Vulnerability reporting and responsible disclosure policy.
- [Contributing Guidelines](CONTRIBUTING.md) — How to contribute code, documentation, and design.

---

## License

NETWATCH is dual-licensed under both the **[MIT License](LICENSE-MIT)** and the **[Apache License 2.0](LICENSE-APACHE)**.  
You may choose either license at your option.
