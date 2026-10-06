# Multi-Protocol Discovery & Evidence Bus

NETWATCH uses an unprivileged, multi-layer sensor architecture to gather authentic network intelligence without requiring Administrator or root privileges.

---

## 1. Unprivileged Subnet Ingestion (Layer 2/3)

On Windows, NETWATCH avoids raw sockets and third-party capture drivers (Npcap/WinPcap) by querying native user-mode APIs:
- **`GetIpNetTable`**: Reads the OS kernel ARP cache populated by normal network activity.
- **`IcmpSendEcho`**: Unprivileged ICMP ping to actively solicit ARP cache entries from quiet hosts across RFC 1918 subnets.
- **`SendARP`**: Single-IP hardware address resolution for fallback verification.

---

## 2. Multi-Protocol Evidence Bus

NETWATCH correlates observable evidence across multiple local protocols:

### A. NetBIOS Name Service (NBNS — RFC 1002)
- Queries UDP port 137 on active hosts.
- Extracts Windows workstation/server names and the 6-byte **Unit ID** (hardware MAC address).
- Detects multi-homed machines and validates MAC addresses against the ARP table.

### B. SSDP / UPnP Discovery (RFC UPnP 1.0)
- Listens to multicast M-SEARCH on `239.255.255.250:1900`.
- **Zero-Fetch Invariant**: Records the `Location` header string as authentic diagnostic evidence only. The engine **never dials or fetches HTTP XML** from devices, preventing SSRF and network tampering.

### C. Multicast DNS (mDNS / DNS-SD — RFC 6762 / 6763)
- Listens on `224.0.0.251:5353`.
- Extracts advertised `.local` services (`_printer._tcp`, `_airplay._tcp`, `_googlecast._tcp`, `_workstation._tcp`).
- Hardened against DNS compression pointer loops and oversized malformed packets.

---

## 3. Destination Policy & Zero-Egress

All discovery probes pass through `IsAllowedDiscoveryDestination`:
- Strictly restricts probe destinations to **RFC 1918 private IPv4 subnets** and link-local multicast.
- Blocks public Internet egress at the socket layer.
- Prohibits probe loops against loopback interfaces.
