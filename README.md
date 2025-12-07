# Go Project Starter Template

Opinionated Go starter with Fiber v3, pre-commit hooks, linting, testing, proto tooling, Docker, and CI skeletons. Version pins in this template are intentional—do not change them during setup.

## Quick Start

1. Clone and set your module name

    ```bash
    git clone https://github.com/hyp3rd/starter.git my-new-project
    cd my-new-project
    ./setup-project.sh --module github.com/your/module
    ```

1. Install toolchain (core). Proto tools stay optional.

    ```bash
    make prepare-toolchain
    # If you need proto/gRPC/OpenAPI
    make prepare-proto-tools
    ```

1. Run quality gates and sample app

    ```bash
    make lint
    make test
    make run   # serves /health on HOSTNAME:PORT (defaults localhost:8000)
    ```

1. Optional: Docker and Compose

    ```bash
    docker build -t starter-app .
    docker compose up --build
    ```

## What’s Included

- Fiber v3 sample service with `/health`
- Pre-commit hooks (gci, gofumpt, golangci-lint, tests, spell/yaml/markdown lint)
- Make targets for lint/test/vet/security/proto
- Proto tooling via buf (configs promoted from examples)
- Dockerfile (multi-stage) and docker-compose.yml
- GitHub Actions templates for lint, test+coverage, proto, security, and pre-commit
- Dependabot for Go modules and GitHub Actions

## Configuration & Customization

- **Module path & imports**: `setup-project.sh` replaces `#PROJECT` in Makefile and hooks. If you change the module later, rerun the script with `--module`.
- **Go version**: Target Go 1.25.x (keep pins as provided).
- **GCI prefix**: Set by the setup script; defaults to `#PROJECT`.
- **Proto**: Copy/edit `api/core/v1/health.proto`. Generate stubs with `make proto`. Generated files land in `pkg/api/core/v1/`.
- **HTTP request examples**: `requests/*.http` read variables from `requests/.env` (copy from `.env-example`). `health_get.http` hits `/health`.
- **Docker**: Override `APP_VERSION`, `HOSTNAME`, `PORT` in `docker-compose.yml` or `docker run -e`.

## Make Targets (high level)

- `prepare-toolchain` — install core tools (gci, gofumpt, golangci-lint, staticcheck, govulncheck, gosec)
- `prepare-proto-tools` — install buf + protoc plugins (optional)
- `lint` — gci, gofumpt, staticcheck, golangci-lint
- `test` / `test-race` / `bench`
- `vet`, `sec`, `proto`, `run`, `update-deps`, `update-toolchain`

Run `make help` for the full list.

## Platform Prerequisites

- Go 1.25.x
- Docker
- Git
- Python 3 + `pre-commit`
- Optional proto toolchain (installed via `make prepare-proto-tools`)

## CI/CD (templates)

- `.github/workflows/lint.yml` — gofumpt, gci, staticcheck, golangci-lint
- `.github/workflows/test.yml` — unit tests (race + coverage artifact)
- `.github/workflows/proto.yml` — buf format/lint/generate
- `.github/workflows/security.yml` — govulncheck + gosec
- `.github/workflows/pre-commit.yml` — pre-commit hooks on all files
- `.github/dependabot.yml` — Go modules + GitHub Actions updates

## Contribution Notes

- Tests required for changes; run `make lint test` before PRs.
- Suggested branch naming: `feat/<scope>`, `fix/<scope>`, `chore/<scope>`.
- Update docs when altering tooling, Make targets, or setup steps.

## Troubleshooting

- **go.mod/go.sum changes after tidy**: run `go mod tidy`, commit the changes.
- **Tool missing errors**: rerun `make prepare-toolchain` (or `prepare-proto-tools` for proto).
- **Pre-commit slow**: run `pre-commit run --all-files` once to warm caches.

## License

GNU GPL v3. See `LICENSE`.
