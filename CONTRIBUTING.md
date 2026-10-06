# Contributing to NETWATCH

<!-- /* Pre-emit critique: P5 H5 E5 S5 R5 V5 D5 */ -->

Thank you for your interest in contributing to NETWATCH! We welcome contributions that align with our core design invariants.

## Core Non-Negotiable Invariants

Before writing code, please review the foundational rules documented in [AGENTS.md](AGENTS.md):

1. **Privacy & Local-First**: Zero external network requests. No telemetry, analytics, cloud accounts, or third-party web fonts.
2. **Local Identity**: MAC/vendor lookup must use the local embedded database (`internal/oui/ieee-oui.txt`). Never transmit MAC addresses off-machine.
3. **No Synthetic Data**: Unknown values stay `Unknown`. Do not add speculative "health scores" or simulated threat alarms.
4. **Contract First**: Any change to shared wire shapes requires simultaneous updates to `internal/model`, `frontend/src/types`, service adapters, and test suites.
5. **OS Isolation**: All OS-specific calls belong exclusively behind `internal/netenv.Env`.

## Development Setup

### Prerequisites
- Go 1.24+ (or Go 1.27)
- Node.js 20+ & npm
- Windows (for native Wails desktop testing) or Linux/macOS (for headless daemon development)

### Running Locally
```bash
# Terminal 1: Start the headless daemon (prints baseUrl and session token)
go run ./cmd/netwatchd -origin http://localhost:3000

# Terminal 2: Start the frontend development server
cd frontend
npm install
npm run dev
# Open the URL printed with ?api=<baseUrl>&token=<token>
```

## Running Tests

All pull requests must pass local verification:

```bash
# Go tests and race detector
go test -v ./...

# Frontend typecheck & lint
cd frontend
npm run lint

# Frontend production build
npm run build
```

## Pull Request Guidelines

1. **Branch Naming**: Use descriptive prefixes: `feat/`, `fix/`, `docs/`, `refactor/`.
2. **Small & Focused**: Keep PRs scoped to one issue or milestone capability.
3. **Keep Changelog Updated**: Add a bullet point to `CHANGELOG.md` under `[Unreleased]` following [Keep a Changelog](https://keepachangelog.com/) format.
4. **Semantic Versioning**: Adhere strictly to SemVer (`MAJOR.MINOR.PATCH`).
