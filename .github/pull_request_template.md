<!-- /* Pre-emit critique: P5 H5 E5 S5 R5 V5 D5 */ -->
## Description
<!-- Provide a concise summary of the changes and motivation. -->

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation / Refactoring / CI

## Architectural Invariants Checklist
- [ ] **Privacy-First**: Zero external network requests, zero telemetry, no cloud calls.
- [ ] **Contract-First**: If wire types changed, `internal/model`, `frontend/src/types`, and tests were updated together.
- [ ] **OS Isolation**: All OS interactions remain behind `internal/netenv.Env`.
- [ ] **Local OUI**: MAC/Vendor lookups use local embedded IEEE OUI data only.
- [ ] **Release Consistency**: `CHANGELOG.md` updated with SemVer classification, and `AGENTS.md` updated if milestone/status changed.

## Testing & Verification
<!-- Describe the tests you ran to verify your changes. -->
- [ ] Backend tests passed: `go test -race ./...` (or `go test ./...` on Windows)
- [ ] Frontend typecheck passed: `cd frontend && npm run lint`
- [ ] Frontend build passed: `cd frontend && npm run build`
