# archived — Software Specification

> **Status:** Active development (MVP proven as of v0.0.1)
>
> **Language of this document:** English

## 1. Overview

**archived** is a cloud‑native service for storing **versioned data in a
space‑efficient manner**. It is designed for sharing **low‑cardinality data**
(relatively few, large, widely shared files) with many users or systems. The
canonical use case is an **APT/RPM (yum) package repository**, where many
versions of the same packages are served to many clients.

The core idea is inspired by `rsync --link-dest`, which allowed package
mirrors to be stored without duplicating data. archived removes the dependency
on a local file system by storing data in modern storage services
(S3‑compatible object storage) and metadata in a database (PostgreSQL).

The key win: **raw data usage is reduced by linking duplicates to a single
storage key instead of storing copies.**

### 1.1 Two backing stores

archived relies on two independent storage backends:

| Store | Role | Example |
| --- | --- | --- |
| **Metadata store** | Tracks the data model (namespaces, containers, versions, objects) | PostgreSQL |
| **CAS store** | Stores raw BLOB bytes behind objects under a content‑derived key (SHA256) | S3 |

Both backends are abstracted behind Go interfaces with mock implementations,
so the core `service` layer is storage‑agnostic.

### 1.2 Domain model

* **namespace** — a group of containers.
* **container** — a logical "directory".
* **version** — an immutable snapshot of the data in a container.
* **object** — a named data BLOB with additional metadata (size, MIME type,
  checksum).

A special version identifier `latest` resolves to the latest *published*
version and is supported by both the manager gRPC API and the HTML publisher.

## 2. Components (microservices)

Each component is a separate binary built from `cmd/<component>`.

| Binary | Entrypoint | Purpose | Backend access |
| --- | --- | --- | --- |
| `archived-manager` | `cmd/manager/main.go` | gRPC API (`ManageService`) to manage namespaces/containers/versions/objects | RW PostgreSQL |
| `archived-publisher` | `cmd/publisher/main.go` | HTTP server for data listing/fetching (Echo + HTML web UI) | RO PostgreSQL replica (scalable), optional memcache |
| `archived-exporter` | `cmd/exporter/main.go` | Prometheus metrics exporter for metadata entities (periodic observe loop) | RO PostgreSQL (single copy) |
| `archived-cli` | `cmd/cli/main.go` | CLI to operate the manager via gRPC | network access to manager |
| `archived-migrator` | `cmd/migrator/main.go` | Runs SQL migrations; must run before other components on upgrade | RW PostgreSQL |
| `archived-gc` | `cmd/gc/main.go` | Garbage collector; deletes expired unpublished versions and their objects; runs periodically | RW PostgreSQL |
| `archived-seeder` | `cmd/seeder/main.go` | Test/dev seeding tool that generates namespaces/containers/versions/objects | S3 + RW PostgreSQL |

### 2.1 Data flow

1. **CLI → Manager → S3:** `archived-cli` issues management commands over
   gRPC. For `version create --from-*`, a content `Source` adapter
   (local/apt/yum) walks content and calls `CreateObject`. The manager returns
   a **presigned upload URL** when the BLOB is absent; the CLI uploads the
   object bytes directly to S3 and then calls `AddObject`.
2. **Publisher → Postgres (+memcache) + S3:** serves HTTP listing/fetching;
   object fetches redirect to presigned S3 URLs.
3. **Exporter → Postgres:** periodic `CountStats` → Prometheus gauges.
4. **GC → Postgres + S3:** deletes expired unpublished versions and their
   unreferenced objects/BLOBs.

**Design note:** large object bytes are uploaded client‑side to S3 via
presigned URLs — the server never proxies large payloads.

## 3. Public interfaces

### 3.1 gRPC API (`ManageService`)

Protobuf: `manager/presenter/grpc/proto/v1/manager.proto`
(package `manager.presenter.grpc.proto.v1`).

| RPC | Request | Response |
| --- | --- | --- |
| `CreateNamespace` | `{name}` | `{}` |
| `RenameNamespace` | `{old_name, new_name}` | `{}` |
| `DeleteNamespace` | `{name}` | `{}` |
| `ListNamespaces` | `{}` | `{repeated name}` |
| `CreateContainer` | `{namespace, name, optional ttl_seconds}` | `{}` |
| `MoveContainer` | `{namespace, container_name, destination_namespace}` | `{}` |
| `RenameContainer` | `{namespace, old_name, new_name}` | `{}` |
| `DeleteContainer` | `{namespace, name}` | `{}` |
| `ListContainers` | `{namespace}` | `{repeated name}` |
| `SetContainerParameters` | `{namespace, name, optional ttl_seconds}` | `{}` |
| `CreateVersion` | `{namespace, container}` | `{version}` |
| `ListVersions` | `{namespace, container}` | `{repeated versions}` |
| `DeleteVersion` | `{namespace, container, version}` | `{}` |
| `PublishVersion` | `{namespace, container, version}` | `{}` |
| `CreateObject` | `{namespace, container, version, key, checksum, size, mime_type}` | `{optional upload_url}` |
| `ListObjects` | `{namespace, container, version}` | `{repeated objects}` |
| `GetObjectURL` | `{namespace, container, version, key}` | `{url}` |
| `DeleteObject` | `{namespace, container, version, key}` | `{}` |

Generated `*.pb.go` files are **gitignored** and regenerated by
`go generate ./...` (requires `protoc` toolchain). CI performs this step.

### 3.2 HTTP (publisher HTML presenter) — Echo routes

* `GET /` → namespace list
* `GET /:namespace/` → container list
* `GET /:namespace/:container/` → published version list
* `GET /:namespace/:container/:version/` → object list
* `GET /:namespace/:container/:version/:object` → **302** redirect to
  presigned S3 URL
* `?page=N` pagination query param on list routes.

### 3.3 CLI command surface (kingpin)

Groups: `namespace` (`create|rename|delete|list`), `container`
(`create|move|rename|delete|set|list`), `version` (`create|delete|list|publish`),
`object` (`list|url|delete`), `stat-cache show-path`.

`version create` flags include: `--publish`, `--from-dir`, `--from-yum-repo`,
`--from-yum-mirrorlist`, `--rpm-gpg-key-path`, `--rpm-gpg-key-checksum`,
`--from-apt-repo`, `--from-apt-repo-suite`, `--from-apt-repo-component`,
`--from-apt-repo-architecture`.

Global flags: `--debug/-d`, `--trace/-t`, `--endpoint/-s` (required),
`--insecure`, `--insecure-skip-verify`, `--cache-dir`, `--namespace/-n`
(default `default`).

### 3.4 Prometheus metrics

Exporter gauges:

`archived_namespaces_amount`, `archived_containers_amount`,
`archived_versions_amount`, `archived_objects_amount`, `archived_blobs_amount`,
`archived_blobs_raw_size_bytes`, `archived_blobs_effective_size_total_bytes`.

Repository layer:

`archived_repository_query_count_total`,
`archived_repository_query_time_seconds_total` (by `kind`).

Plus grpc‑middleware and Echo middleware metrics
(`manager_metrics`, `publisher_metrics`, `exporter_metrics` prefixes).

## 4. Core Go interfaces

* `service.Manager` — full mutation API (embeds `Publisher`).
* `service.Publisher` — read‑only listing/fetching API.
* `repositories/metadata.Repository` — all CRUD for
  namespaces/containers/versions/objects/BLOBs + `CountStats`. Sentinels:
  `ErrNotFound`, `ErrConflict`.
* `repositories/blob.Repository` — `PutBlobURL(key)`,
  `GetBlobURL(key, mimeType, filename)`.
* `cli/service/source.Source` — content ingestion adapters
  (apt, local, mock, yum).
* `cli/service/stat_cache.CacheRepository`, `cli/lazyblob.LazyBLOB`,
  `cli/router.Router`.
* `manager/presenter/grpc.ManageServerInterface`,
  `publisher/presenter/html.Handlers`.

The `service` layer wraps the metadata/BLOB repositories and maps
`metadata.ErrNotFound` → `service.ErrNotFound`. Fixed page sizes in
`service.NewManager` are 50/50/50; `service.NewPublisher` accepts configurable
page sizes.

## 5. Technology stack

* **Language:** Go (`go 1.26.0`); CI uses `1.26.x`.
* **Build/Release:** goreleaser v2, `protoc` + `protoc-gen-go`/
  `protoc-gen-go-grpc`.
* **Dependencies:** Go modules; Dependabot (weekly gomod).
* **Web:** `labstack/echo/v4` + `echo-contrib/echoprometheus`.
* **gRPC:** `google.golang.org/grpc`, grpc‑middleware (prometheus, logging,
  recovery).
* **Database:** PostgreSQL via `lib/pq`; query builder `Masterminds/squirrel`;
  migrations via `golang-migrate/migrate/v4`.
* **Object storage:** AWS SDK v2 (`aws-sdk-go-v2`, S3).
* **Cache:** `bradfitz/gomemcache`.
* **CLI:** `alecthomas/kingpin/v2`; **Config:** `kelseyhightower/envconfig`;
  **Logging:** `sirupsen/logrus`; **Metrics:** `prometheus/client_golang`;
  **Tracing:** `go.opentelemetry.io/otel/trace`.
* **Package parsing:** `sassoftware/go-rpmutils`, `ProtonMail/go-crypto/openpgp`,
  `ulikunitz/xz`, `pault.ag/go/debian`, `pault.ag/go/topsort`, custom `yum_repo`
  models.
* **Validation:** `go-ozzo/ozzo-validation/v4`; **Templating:**
  `Masterminds/sprig/v3`.
* **Tests:** `stretchr/testify`, `teran/go-docker-testsuite` (Docker‑backed),
  `teran/go-grpctest`.
* **Frontend (publisher UI):** TypeScript (`tsc --strict`), Bootstrap 5.3.3,
  Bootstrap Icons (built via `npm install` in a multi‑stage Dockerfile).

## 6. Configuration

All components except the CLI are configured via **environment variables**
using `envconfig` (`envconfig.MustProcess("", &cfg)`). Full reference:
`docs/configuration.md`.

### Shared across components

| Variable | Type | Required | Default |
| --- | --- | --- | --- |
| `LOG_LEVEL` | logrus.Level | No | `info` |
| `METADATA_DSN` | string | **Yes** | — |

### `archived-manager`

`ADDR` (`:5555`), `METRICS_ADDR` (`:8081`), and the shared `BLOB_S3_*` set:
`BLOB_S3_ENDPOINT` (required), `BLOB_S3_BUCKET` (required),
`BLOB_S3_CREATE_BUCKET` (`false`), `BLOB_S3_PRESIGNED_LINK_TTL` (`5m`),
`BLOB_S3_ACCESS_KEY_ID` (required), `BLOB_S3_SECRET_KEY` (required),
`BLOB_S3_REGION` (`default`), `BLOB_S3_DISABLE_SSL` (`false`),
`BLOB_S3_FORCE_PATH_STYLE` (`true`).

### `archived-publisher`

`ADDR` (`:8080`), `METRICS_ADDR` (`:8081`), shared `BLOB_S3_*`, plus:
`MEMCACHE_SERVERS` (empty = cache disabled), `MEMCACHE_TTL` (`60m`),
`BLOB_S3_PRESERVE_SCHEME_ON_REDIRECT` (`true`), `HTML_TEMPLATE_DIR`
(required), `STATIC_DIR` (required),
`VERSIONS_PER_PAGE`/`OBJECTS_PER_PAGE`/`CONTAINERS_PER_PAGE` (`50`),
`MAX_PAGES_IN_PAGINATION` (`5`), `DEFAULT_THEME` (`dark`|`light`).

### `archived-exporter`

`METRICS_ADDR` (`:8081`), `OBSERVE_INTERVAL` (`60s`).

### `archived-gc`

`UNPUBLISHED_VERSION_MAX_AGE` (`168h`).

### `archived-migrator`

`METADATA_DSN` only.

### `archived-seeder`

Shared `BLOB_S3_*` + `CREATE_NAMESPACES` (10),
`CREATE_CONTAINERS_PER_NAMESPACE` (100), `CREATE_VERSIONS_PER_CONTAINER` (100),
`CREATE_OBJECTS_PER_VERSION` (100), `MAX_OBJECT_SIZE_BYTES` (4096).

### `archived-cli`

`ARCHIVED_CLI_DEBUG`, `ARCHIVED_CLI_TRACE`, `ARCHIVED_CLI_ENDPOINT`,
`ARCHIVED_CLI_STAT_CACHE_DIR`.

> **Known docs/code drift (to reconcile):**
>
> * `docs/configuration.md` documents a `DRY_RUN` var for gc that is **not
>   implemented**.
> * `docs/configuration.md` lists manager `ADDR` default `:8080`; code uses
>   `:5555`.
> * Docs do not yet cover `seeder`, exporter `OBSERVE_INTERVAL`, gc
>   `UNPUBLISHED_VERSION_MAX_AGE`.

## 7. Deployment

Distributed as prebuilt binaries; deployable via systemd, Kubernetes,
docker-compose, etc.

Deployment prerequisites / constraints:

* `archived-publisher` can use an **RO PostgreSQL replica** and can scale.
* `archived-manager` requires **RW PostgreSQL** (writes); can also scale.
* `archived-exporter` is sufficient in a single copy; RO replica access is
  enough.
* `archived-migrator` must run **on every upgrade, right before other
  components**.
* `archived-cli` runs anywhere with network access to the manager.
* `archived-gc` requires RW PostgreSQL and runs periodically as a job.
* **Security:** there is **no authentication on any stage** at the moment
  (including CLI/manager).

Kubernetes example manifests: `docs/examples/deploy/k8s/`. Prerequisites per
README: ingress‑nginx, OpenEBS hostpath, VictoriaMetrics/Prometheus operator
(PodMonitors), CloudNativePG, external S3.

Docker images run **unprivileged** (`USER nobody`, copied `/etc/passwd`/
`/etc/group`, `FROM scratch` final stages with Alpine `ca-certificates`).

## 8. Build, test, run

Prerequisites: Go v1.22+ (currently 1.26), goreleaser v2.0+,
protoc‑gen‑go v1.34+, protoc‑gen‑go‑grpc v1.4, docker.

```shell
# Regenerate gRPC code from .proto, then build
go generate ./...
goreleaser build --snapshot --clean

# Container images
docker-compose build
# or per-component:
docker build -f dockerfiles/Dockerfile.<component> .

# Full local stack
docker-compose up
# or from source:
go generate -v ./... && \
goreleaser build --snapshot --clean && \
docker-compose build && \
docker-compose up || docker-compose down

# Tests (NOTE: require Docker for PostgreSQL/memcached via go-docker-testsuite)
go test ./...
```

There is **no Makefile**; orchestration is via goreleaser, `go generate`, and
docker-compose.

### CI/CD (GitHub Actions)

* `verify.yml` (push to `master`, PRs): `buf lint`, `hadolint`
  (7 Dockerfiles), `markdownlint`, `golangci` (golangci-lint, installs protoc,
  runs `go generate`), `unittests` (`go test ./...`), `build` (goreleaser
  snapshot + docker build‑push, `push: false`).
* `release.yml` (`v*` tags): same lint/test gates, then goreleaser
  `release --clean` and push of all 7 images to
  `ghcr.io/teran/archived/<component>` with tags `latest`, `<ref_name>`,
  `<ref_name>-<timestamp>`.

> **Note:** CI runs `buf lint .` and golangci‑lint but the repo has **no**
> committed `buf.yaml`, `.golangci.yml`, or `.editorconfig` — defaults are
> used.

## 9. Coding conventions

* Standard Go; interface‑first design (concrete types returned as interfaces).
* Error wrapping via `pkg/errors` (`errors.Wrap`, `errors.Errorf`); error
  sentinel package vars (`metadata.ErrNotFound`, `service.ErrNotFound`).
* Inline `//nolint:<linter>` directives (errorlint, sqlclosecheck, err113,
  unparam, …).
* Interface assertions: `var _ metadata.Repository = (*memcache)(nil)`.
* `logrus` logging with `log.WithFields(log.Fields{...})`.
* `envconfig` struct tags on every binary's config struct.
* Testing with `testify`; repository tests Docker‑backed via
  `go-docker-testsuite`; presenter tests via `go-grpctest`; `testdata/` dirs
  alongside source.
* Frontend: TypeScript `--strict`, Bootstrap 5.

## 10. Current state & roadmap

* Active development; **not archived** ("archived" is the project name).
* MVP shipped as v0.0.1. "Almost everything is a subject to change."
* Full feature list tracked in GitHub issues.
* Recent functional work: migration to `aws-v2`, Debian source support (#311),
  seeder tooling (still commented out of docker-compose).
* Only `context.TODO()` calls exist in production/test code; no outstanding
  TODO/FIXME in production code.

## 11. Data model evolution (migrations)

Notable schema evolution (10 migrations, `0000`–`0009`):

* `0002` — object keys moved to a dedicated `object_keys` table.
* `0004` — namespaces introduced.
* `0005` — blob sizes became BIGINT.
* `0006` — container version TTL added.
* `0003`, `0008`, `0009` — index‑only migrations.

## 12. Out of scope / non‑goals

* Authentication and authorization (explicitly out of scope at present).
* Object‑storage filesystem semantics (CAS + presigned URLs, not a POSIX
  store).
* Non‑Go language support.
