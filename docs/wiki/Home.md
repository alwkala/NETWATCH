# Welcome to the NETWATCH Wiki

**"Know Every Device on Your LAN. Without the Cloud Watching."**

NETWATCH is a local network intelligence and persistent asset ledger desktop application for Windows, Linux, and macOS.

---

## 📚 Table of Contents

- [[Architecture]] — Comprehensive overview of the Go engine, React 19 interface, and IPC model.
- [[Discovery and Evidence]] — Deep dive into unprivileged subnet sweeps, mDNS, SSDP, and NBNS probing.
- [[Asset Ledger and Reconciliation]] — How the persistent SQLite store tracks device lifecycles, flap resistance, and custom identities.
- [[Security and Privacy]] — Cryptographic loopback isolation, zero-egress CI gates, and air-gapped hardware fingerprinting.

---

## 🎯 Core Philosophy

Unlike traditional network scanners (*Advanced IP Scanner*, *Angry IP Scanner*) that discard all information once closed:
- **Scanning is only an ingestion sensor**.
- **The Core Product is the Local Asset Ledger**: An embedded SQLite database tracking the lifecycle of every network asset across days, weeks, and months.
- **State Reconciliation**: Answers not just *"What is online right now?"*, but *"What changed? Who joined? Who departed? When did an IP drift?"*
- **Flap Resistance**: Devices only transition offline after consecutive missed sweeps (`misses = 2`), preventing false alerts when low-power Wi-Fi devices sleep.

---

## 🚀 Quick Links
- [Main Repository](https://github.com/alwkala/NETWATCH)
- [Releases & Downloads](https://github.com/alwkala/NETWATCH/releases)
- [Threat Model](https://github.com/alwkala/NETWATCH/blob/main/THREAT_MODEL.md)
- [Roadmap](https://github.com/alwkala/NETWATCH/blob/main/ROADMAP.md)
