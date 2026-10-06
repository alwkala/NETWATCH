# Security Policy


## Security Invariants

NETWATCH is engineered with strict local-first security boundaries:
1. **Loopback Only**: The backend HTTP and SSE API binds strictly to `127.0.0.1` on an ephemeral port.
2. **Per-Session Bearer Token**: Every API call requires an unguessable 192-bit cryptographic token generated per application startup.
3. **Anti DNS-Rebinding & Origin Isolation**: All incoming requests are validated against strict `Host` and `Origin` allow-lists, isolating the daemon from malicious web pages open in local browsers.
4. **No Telemetry / No Cloud**: Zero sensitive data (MAC addresses, IP configurations, network topologies) ever leaves the local machine.

## Supported Versions

| Version | Supported |
|---|---|
| `0.1.x` | :white_check_mark: |
| `< 0.1.0` | :x: |

## Reporting a Vulnerability

If you discover a security vulnerability or bypass in NETWATCH's loopback protection, token verification, or packet handling:

1. **Do not open a public issue.**
2. Report the vulnerability privately via [GitHub Security Advisories](https://github.com/alwkala/NETWATCH/security/advisories/new).
3. Provide detailed steps to reproduce, including environment details (OS, network configuration) and proof of concept if available.

### Response SLA
- **Initial Acknowledgement**: Within 48 hours.
- **Triage & Assessment**: Within 5 business days.
- **Remediation & Patch Release**: Disclosed responsibly once a patch is tested and published.
