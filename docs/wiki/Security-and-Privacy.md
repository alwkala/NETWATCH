# Security & Privacy Principles

NETWATCH is engineered with strict, sovereign local-first boundaries.

---

## 1. Zero External Egress

- **Network Topology Sovereign**: IP addresses, MAC mappings, hostnames, and Wi-Fi SSIDs never leave the host system.
- **Air-Gapped OUI Database**: Embedded IEEE OUI database (`internal/oui/ieee-oui.txt`) provides offline hardware vendor identification without cloud queries.
- **Automated CI Enforcement**: A dedicated CI test (`TestZeroEgressInvariant`) parses Go ASTs and frontend bundles on every commit, failing immediately if any `net/http` client dialer or external CDN stylesheet/script is introduced.

---

## 2. Loopback-Only API Isolation

The Go engine exposes a local REST and SSE API bound strictly to `127.0.0.1`:
- **Ephemeral Port Binding**: Binds to a randomly chosen local port at startup.
- **192-Bit Session Bearer Token**: Generates a cryptographically random Bearer token on launch (`crypto/rand`). Requests missing or having invalid tokens receive `401 Unauthorized`.
- **Anti-DNS Rebinding**: Rejects any request where the `Host` header is not `127.0.0.1:<port>` or `localhost:<port>`.
- **Origin Validation**: Strict Origin header allow-listing ensures local browser tabs cannot bridge to the engine.

---

## 3. Untrusted Input Hardening

- **LAN String Sanitization**: Hostnames, mDNS records, and NetBIOS labels pass through `SanitizeLANString` to strip control characters, ANSI escape codes, and Unicode Bidi override runes.
- **CSV Formula Injection Defense**: Any exported data fields beginning with `=`, `+`, `-`, or `@` are safely escaped to prevent spreadsheet execution vulnerabilities (CWE-1236).
