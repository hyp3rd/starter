# RFC: Standardized Go Project Toolchain and Structure

## Status

Proposed

## Context

We maintain multiple Go services/libraries with divergent tooling, layouts, and CI practices. This increases onboarding time, runtime drift, and compliance risk. A single, opinionated starter and toolchain will cut variance while keeping proto/gRPC optional.

## Goals

- Establish a repeatable, production-ready baseline for Go apps and packages.
- Codify tooling, versioning, CI, and security checks so teams don’t reinvent them.
- Keep customization points clear (module name, proto usage, Docker image, GCI prefix).

## Non-Goals

- Mandating specific product features or frameworks beyond the chosen baseline (Fiber v3 for the sample app; proto/gRPC is opt-in).
- Replacing product-specific infra (observability, secrets, deploy pipelines).

## Proposal

Adopt the `starter` layout and toolchain as the default for new Go projects and as a migration target for existing repos during major refactors.

### Tooling Baseline

- **Formatting/Linting**: `gci`, `gofumpt`, `golangci-lint`, `staticcheck` driven by `make lint` and pre-commit hooks.
- **Security**: `govulncheck`, `gosec` via `make sec` and CI.
- **Testing**: `go test -race -cover` in CI; local `make test`/`test-race`.
- **Proto (optional)**: `buf` with managed `buf.yaml`/`buf.gen.yaml`, optional workflow gated by `PROTO_ENABLED`.
- **Package versions**: Centralized in `.project-settings.env`; consumed by Makefile, hooks, Dockerfile, and CI workflows to avoid drift.

### Project Structure

- `cmd/app` (sample Fiber v3 app with healthcheck) + `internal/` for services + `pkg/` for shared library code.
- `api/` for protos (optional); generated stubs under `pkg/api/...`.
- `requests/` for HTTP examples wired to `.env`.

### Environment & Config

- `.project-settings.env` as the single source for Go/tool versions, GCI prefix, proto toggle.
- `.env.example` for runtime defaults consumed by the app, Docker Compose, and HTTP requests.

### CI/CD

- GitHub Actions for lint, test (race+coverage artifact), proto (skipped when disabled), security, and pre-commit; all source `.project-settings.env`, cache Go modules/build, and enforce `go mod tidy` cleanliness.
- Dependabot for Go modules and Actions.

### Containers

- Multi-stage Dockerfile (non-root) with built-in healthcheck (`/app -healthcheck`).
- Compose wired to `.env` with healthcheck.

### Developer UX

- `make init` to run setup + tool installs (honors proto toggle).
- `setup-project.sh --module <path>` updates module/import prefixes and settings.
- App flags `-addr` and `-healthcheck` for easy probes and overrides.

## Rationale

- **Consistency**: Shared make targets, hooks, and CI remove per-team divergence and lower onboarding time.
- **Safety**: Security scanners and tidy checks run by default; non-root images with healthchecks.
- **Velocity**: Cached CI, pre-commit hooks, and generated stubs reduce review churn.
- **Flexibility**: Proto and gRPC remain optional; module/GCI prefixes are template-driven; library-only consumers can disable proto.
- **Governance**: Centralized versions and CI enforce minimum standards without blocking custom additions.

## Migration Plan (high-level)

1) New projects: start from the starter template, run `./setup-project.sh --module <module>`, `make init`.
2) Existing projects (opt-in): align layout (`cmd/`, `internal/`, `pkg/`), add `.project-settings.env`, adopt Makefile targets, wire CI workflows, and optionally adopt proto toolchain; gate proto via `PROTO_ENABLED=false` if not needed.
3) Verify with `make lint test` and ensure CI green; add Docker healthcheck if containerized.

## Risks & Mitigations

- **Tool/version drift**: Mitigated by sourcing from `.project-settings.env` and Dependabot updates.
- **Proto overhead for libraries**: Mitigated by `PROTO_ENABLED=false` and conditional workflows.
- **CI time**: Cached modules/build and tidy-only diffs reduce cost; can further scope CI matrices if needed.

## Open Questions

- Do we enforce a minimum coverage threshold in CI?
- Should we standardize on additional observability/logging defaults (e.g., OTEL) in the starter?
- Should we add a reusable workflow composite to reduce duplication across repos?
