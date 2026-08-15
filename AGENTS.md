# AGENTS.md — Guidance for AI coding agents working in `archived`

This file gives AI coding agents (and humans) the conventions, constraints and
gotchas needed to work safely and effectively in this repository. Read
`SPEC.md` for the full architecture and API reference.

> **Language:** All documentation, code comments and commit messages in this
> repository are written in **English**. Keep any new docs/comments in English.

## Repository at a glance

* **Purpose:** Cloud‑native service for space‑efficient storage of versioned,
  low‑cardinality data (canonical use case: APT/RPM package repositories).
* **Two storage backends:** PostgreSQL (metadata) + S3‑compatible object
  storage (CAS blobs). Both are abstracted behind Go interfaces.
* **Go module:** `github.com/teran/archived`, Go `1.26.0`.
* **Architecture:** microservices, one binary per component under `cmd/`
  (manager, publisher, exporter, cli, migrator, gc, seeder).

## Before you start

1. **Regenerate gRPC code.** `*.pb.go` files are **gitignored** and not
   committed. A fresh checkout (or CI) must run `go generate ./...` to produce
   them from `manager/presenter/grpc/proto/v1/manager.proto`. Requires the
   `protoc` toolchain. **Never hand‑edit generated `.pb.go` files.**
2. **Tests need Docker.** `go test ./...` spins up PostgreSQL/memcached via
   `github.com/teran/go-docker-testsuite`. Without Docker, repository tests
   will fail. Component/presenter tests may use `go-grpctest`.

## Common commands

```shell
go generate ./...            # regenerate gRPC stubs from .proto
goreleaser build --snapshot --clean   # build all binaries
go test ./...                # run tests (requires Docker)
go build ./...               # compile check without pb.go? Run go generate first
go vet ./...                 # static checks (golangci-lint in CI)
```

There is **no Makefile**. Use goreleaser + `go generate` + docker‑compose.

## How to structure changes

* **Interface‑first:** concrete types are returned as interfaces
  (`service.Manager`, `repositories/metadata.Repository`, etc.). Add or change
  methods at the interface level and update all implementations (postgresql,
  mock, memcache wrapper).
* **Error handling:** use `github.com/pkg/errors` (`errors.Wrap`,
  `errors.Errorf`) for wrapping. Use package‑level error sentinels
  (`metadata.ErrNotFound`, `metadata.ErrConflict`, `service.ErrNotFound`) and
  map them at presenter boundaries (e.g. `metadata.ErrNotFound` → gRPC
  `codes.NotFound`).
* **Config:** every binary gets its own `config` struct with `envconfig` tags,
  processed via `envconfig.MustProcess("", &cfg)`. Add new env vars there and
  update `docs/configuration.md`.
* **Logging:** `logrus`; use `log.WithFields(log.Fields{...})`. Text formatter
  with full timestamps.
* **Testing:** `stretchr/testify` (assert/require). Keep `testdata/` dirs
  alongside source. Repository tests are Docker‑backed; presenter tests use
  `go-grpctest`.
* **Frontend (publisher UI):** TypeScript with `--strict`; Bootstrap 5.
  Rebuilt via `npm install` in the multi‑stage Dockerfile.

## Linting & quality gates (run by CI)

CI (`verify.yml`) runs: `buf lint` (proto), `hadolint` (7 Dockerfiles),
`markdownlint`, `golangci-lint` (which itself runs `go generate`),
`go test ./...`, and goreleaser snapshot build + docker build.

* The repo currently has **no committed** `buf.yaml`, `.golangci.yml`, or
  `.editorconfig` — CI relies on defaults. If you add linter config, wire it
  up consistently with CI.
* Follow existing inline `//nolint:<linter>` conventions (errorlint, err113,
  sqlclosecheck, unparam, etc.). Prefer fixing the underlying issue over
  adding a new suppression.
* Keep `.proto` clean; `buf lint` must pass.
* Dockerfiles must pass `hadolint` and are built as unprivileged (`USER
  nobody`, `FROM scratch`, Alpine `ca-certificates`).

## Key architecture facts to respect

* **Presigned‑URL upload flow:** objects are uploaded client‑side
  (CLI/seeder) directly to S3 via presigned PUT URLs returned by the manager.
  Do not change this to proxy bytes through the server without strong
  justification.
* **Dedup via CAS:** duplicates are linked to the same content‑derived key
  (SHA256) instead of being stored twice. Preserve this property.
* **`latest` version alias:** `versionID == "latest"` resolves to the latest
  published version (both gRPC service and HTML publisher). Handle it
  consistently.
* **`AddObject` trims a leading `/` from object keys** — keep this behavior.
* **Page sizes:** `service.NewManager` fixes page sizes at 50/50/50;
  `service.NewPublisher` takes configurable page sizes. Don't hardcode
  pagination elsewhere.
* **Publisher can use an RO Postgres replica** and scale; manager needs RW.
  Keep read paths (publisher/exporter) read‑only.

## Security constraints

* **There is no authentication on any stage** (including CLI/manager). Do not
  silently add security that breaks the documented deployment model; flag
  security work explicitly and coordinate (run a `security` review before
  implementing auth‑related changes).
* Redirect scheme handling in the publisher is guarded by
  `BLOB_S3_PRESERVE_SCHEME_ON_REDIRECT` and an allow‑list of `http`/`https` —
  preserve the allow‑list (no arbitrary scheme injection).

## Known docs/code drift (verify before relying on docs)

* `docs/configuration.md` documents a `DRY_RUN` var for `archived-gc` that is
  **not implemented** in code.
* `docs/configuration.md` lists manager `ADDR` default `:8080`; code uses
  `:5555`.
* Docs do not yet cover `seeder`, exporter `OBSERVE_INTERVAL`, or gc
  `UNPUBLISHED_VERSION_MAX_AGE`.
* When you change behavior, update both code and `docs/configuration.md`.

## Git / workflow notes

* Default branch: `master`. Releases are tagged `v*` and built by the
  `release.yml` workflow (goreleaser → `ghcr.io/teran/archived/<component>`).
* Dependency bumps are handled by Dependabot (weekly gomod) — keep `go.mod` /
  `go.sum` tidy.
* Commit messages in the repo are concise and conventional (e.g. "Migrate to
  aws‑v2", "Debian source (#311)"). Match that style.
* Large exploratory feature branches are common (trunk‑based + feature
  branches). Keep generated artifacts out of commits (`*.pb.go`, `dist/` are
  ignored).

## Definition of done

* [ ] `go generate ./...` produces no diffs in generated code.
* [ ] `go build ./...` and `go vet ./...` pass.
* [ ] `go test ./...` passes (Docker available).
* [ ] Linters (golangci‑lint, buf lint, hadolint, markdownlint) pass.
* [ ] New env vars documented in `docs/configuration.md` and wired into the
      component's `config` struct.
* [ ] New public interfaces/behaviors reflected in `SPEC.md`.
* [ ] Docs/comments in English.
