## NETWATCH v0.2.0-alpha.1 — Multi-Protocol Discovery, User Customization & Asset Ledger

> **"Know Every Device on Your LAN. Without the Cloud Watching."**

NETWATCH **v0.2.0-alpha.1** marks the completion of **Milestone 3 (Settings & Database Management)** and **Milestone 4 (Multi-Protocol Discovery, Evidence & Identity)**. This release introduces user-customizable device classifications, RFC 5389 STUN-based external public IP discovery, multi-protocol identification (mDNS, SSDP, NetBIOS, ARP), and a responsive mobile-first UI layout.

---

### Key Capabilities & Highlights

1. **User-Defined Device Type & Icon Customization**:
   - Classify devices as Workstation, Laptop, Mobile, Tablet, Server, Gateway, Printer, Switch, Access Point, NAS, IoT, or Camera with real-time icon updates.
   - Schema migration v4 (`custom_type`) guarantees user classifications are permanently preserved across background rescans.
2. **WAN Public IP Discovery & Network Topology**:
   - Pure binary RFC 5389 STUN NAT traversal over UDP, displaying the public IP on the WAN/INTERNET topology node with zero HTTP egress.
3. **Multi-Protocol Discovery & Identification Evidence (M4)**:
   - Directed NetBIOS Name Service (NBNS) queries over UDP 137 with MAC cross-correlation.
   - SSDP / UPnP multicast discovery over `239.255.255.250:1900` with strict Zero-Fetch Location invariant.
   - Multi-pass mDNS / DNS-SD probing over `224.0.0.251:5353` with loop and packet bombing defenses.
   - Identification Evidence drawer displaying observable protocol signals and service records.
4. **Randomized MAC Detection & Merging**:
   - IEEE 802 LAA private MAC detection with "Private MAC" badges.
   - Interactive Device Merge & Aliasing engine (`POST /v1/devices/{id}/merge`).
5. **Mobile-First Responsive Layout (/tidyfactor-styler)**:
   - Slide-over drawer navigation and responsive top bar across all viewports while strictly preserving brand identity invariants.
6. **Desktop System Integration (M3)**:
   - Native Windows toast notifications via WinRT.
   - Autostart with `--minimized` background launch.
   - SQLite WAL compaction, real-time statistics, and retention pruning.
7. **Security Hardening & Fuzz-Tested Parsers**:
   - 600,000+ iterations executed across native Go fuzz tests (`FuzzParseMDNS`, `FuzzParseSSDP`, `FuzzParseNBNS`) with 0 panics.
   - Reverse DNS queries bounded strictly to local private resolvers preventing DNS information leakage to the internet.
   - Constant-time verification for SSE stream token authorization with zero token reflection.

---

### Visual Walkthrough & Screenshots

<div align="center">
  <img src="https://raw.githubusercontent.com/alwkala/NETWATCH/main/docs/Snapshots/home-desktop-dark-theme.png" alt="NETWATCH Device Ledger" width="800px" />
  <p><em>Real-Time Asset Ledger with latency trends, device classifications, and private MAC badges.</em></p>
  <br>
  <img src="https://raw.githubusercontent.com/alwkala/NETWATCH/main/docs/Snapshots/nodes-desktop-dark-theme.png" alt="Network Topology View" width="800px" />
  <p><em>Physical and logical network hierarchy with WAN Public IP and broadcast domain.</em></p>
  <br>
  <img src="https://raw.githubusercontent.com/alwkala/NETWATCH/main/docs/Snapshots/settings-desktop-dark-theme.png" alt="Device Inspector & Evidence" width="800px" />
  <p><em>Device Dossier: Custom icons, trust level management, and multi-protocol discovery evidence.</em></p>
</div>

---

### Assets & Integrity (SHA-256)

| Asset | Description | SHA-256 Checksum |
|---|---|---|
| `netwatch-v0.2.0-alpha.1-windows-amd64.zip` | Standalone Windows 64-bit Desktop App (.zip) | `266E0DA39BFEF52F99928CA5022872F304A07E45F66F0AAA0962B05554E84808` |
| `netwatch.exe` | Direct Windows 64-bit Executable (.exe) | `EE5060D1A547926E6B937D2CE798EDC326D724CB36D6B1D6F9032F3A0B5048CC` |
| `SHA256SUMS.txt` | Authoritative Cryptographic Hashes | *(Raw checksum file)* |

#### Hash Verification:
```powershell
Get-FileHash .\netwatch.exe -Algorithm SHA256
```

#### Windows SmartScreen Note:
Because NETWATCH is distributed unsigned to maintain an independent, zero-cloud open-source model, Windows Defender SmartScreen may display *"Unknown Publisher / Windows protected your PC"* on initial launch.
1. Click **"More info"**
2. Click **"Run anyway"**
*(NETWATCH requires no Administrator elevation and runs 100% locally)*

---

### Quality Assurance & Validation
- `go test -race ./...`: 100% PASS (Ubuntu & Windows CI)
- `tsc --noEmit && vite build`: Clean production build
- Multicast Parsers: Fuzz-tested with 600,000+ random mutations (0 panics)
- Zero-Egress Invariant: Verified by AST analysis (`TestZeroEgressInvariant`)

**Full Changelog**: https://github.com/alwkala/NETWATCH/compare/v0.1.0-alpha.1...v0.2.0-alpha.1
