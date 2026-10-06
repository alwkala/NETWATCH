# Asset Ledger & Reconciliation Engine

The heart of NETWATCH is its embedded SQLite database (`%LOCALAPPDATA%\NetWatch\data\network.db`), running with Write-Ahead Logging (WAL) for high concurrency and zero lock contention.

---

## 1. Flap Resistance & State Transitions

Network devices—particularly smartphones, tablets, and smart home sensors—frequently sleep to conserve battery. Ephemeral scanners report them as disconnected, spamming the user with false departure alerts.

NETWATCH solves this via **Flap-Resistant State Reconciliation**:
- **Offline Threshold**: A device is marked offline only after **two consecutive missed sweeps** (`OfflineAfterMisses = 2`) and a minimum elapsed time window (`MinOfflineDuration = 5m`).
- **Immediate Recovery**: As soon as a single ARP or ICMP packet is observed, the device returns online instantly.

---

## 2. Randomized MAC Detection & Device Merge

Modern mobile operating systems (iOS, Android, Windows 10/11) rotate MAC addresses across Wi-Fi networks using Locally Administered Addresses (LAA).

- **LAA Detection**: NETWATCH inspects bit 1 of the first MAC octet (`mac[0] & 0x02 != 0`). If set, the UI renders an explanatory **"Private MAC"** badge.
- **Device Merge Workflow** (`POST /v1/devices/{targetId}/merge`):
  When a device rotates its private MAC address, users can merge the new identity into the canonical asset record. The engine updates the `device_mac_aliases` table and automatically merges future sightings under the unified device history.

---

## 3. Trust Tiers

Every inventoried device carries a trust status:
- **`known`**: Approved, permanent asset on the subnet.
- **`guest`**: Temporary or visiting client.
- **`unknown`**: Newly discovered or unverified device.

---

## 4. Custom Device Types & Icon Persistence

Users can override heuristic device classifications with custom types:
- Supported categories: `Computer`, `Phone`, `Tablet`, `Router`, `Server`, `Network Device`, `Printer`, `TV`, `Camera`, `Game Console`, `IoT`, `Unknown`.
- Stored permanently in SQLite (`custom_type`).
- Guaranteed to never be overwritten by automated background scans or heuristic reclassifications.
