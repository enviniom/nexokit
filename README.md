# NexoKit

NexoKit is an opinionated modular Go starter for SaaS APIs. It ships with authentication, RBAC, multitenancy, migrations, a developer CLI, and a vertical-slice module structure so you can fork it and start building domain logic instead of boilerplate.

## Stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.26.3+ (see `go.mod`) |
| HTTP framework | Gin |
| ORM | GORM (runtime queries only) |
| Database | PostgreSQL |
| Migrations | Goose (`migrations/`) |
| Auth | PASETO v4.local access tokens + opaque refresh tokens |
| Passwords | argon2id |
| Cache | Redis/Valkey, optional with `CACHE_DRIVER=none` |
| Logging | `slog` + lumberjack rotation |
| Testing | Standard Go testing + httptest + testify |

## Architecture in one paragraph

The codebase is organized as a modular monolith: `cmd/` entrypoints wire `internal/app/`, which composes business modules under `internal/modules/`. Each module owns its vertical slices (`handler`/`service`/`repository`), shared `core/` types, and reusable `queries/`. Business modules do not import each other directly; collaboration is expressed through small interfaces owned by the module that needs or exposes the capability, then wired in `internal/app/container.go`. The database schema is the source of truth in `migrations/`, not Go models.

## Quick start

Prerequisites: installed Go 1.26.3+ (the current `go.mod` requirement) and project
dependencies available to Go. Run `make init` immediately after cloning/forking:
it prepares only local project files and does not require a running database.

```bash
make init
# Adapt the initial migrations, seeds and modules before applying any schema.
go build ./...
go test ./...
```

When you are ready to prepare the database, use an appropriate empty PostgreSQL
database with migration permissions. You can use the service already defined in
`docker-compose.yml`; a separate PostgreSQL installation is not required:

```bash
docker compose up -d postgres
docker compose ps postgres
# Wait until PostgreSQL is healthy, then execute these deliberate steps:
make migrate-up
make seed
make create-root
make dev
```

Redis is optional: use `CACHE_DRIVER=none` and the default in-memory rate limiter
unless you deliberately configure Redis. Neither `init` nor Make provisions services
automatically. Only edit initial migrations before their first application; use new
forward migrations for databases that already have applied migration history.

With the Compose defaults, enter `DB_HOST=localhost`, `DB_PORT=5432`,
`DB_NAME=nexokit`, `DB_USER=nexokit`, `DB_PASSWORD=nexokit` and
`DB_SSL_MODE=disable` in the initializer. These are local-development defaults,
not production credentials. If `.env`, exported `DB_*` variables or an existing
database volume use different settings, enter the actual connection values instead.
The initializer's proposed database name/user may differ from the Compose defaults.

`make init` is the once-only local initialization entrypoint. It collects project
identity and configuration (or an explicit keep/backup-and-rebuild choice for an
existing `.env`), previews the configured nonsecret database target without
contacting it, then asks for **one confirmation, default No**. Approval applies
only identity/environment files. It does not run migrations or seeds, create a
root user, or request root credentials. It prints the later manual commands instead.
Create the root after migrations and seeds with `make create-root`; prefer protected
`ROOT_USER_*` process inputs rather than password arguments or dotenv. The separate
create-root command's interactive password prompt is not hidden. Start the full
application only after schema, seeds and root are ready.

The API boots on `http://localhost:8080` (or the selected `APP_PORT`). Do not rerun
`init` on an adopted project; see [setup and recovery](docs/starter-template.md).
Keeping an existing environment preserves its bytes and mode and uses its database
configuration; it may require deliberate validation/configuration updates before setup.

## Everyday commands

| Command | Purpose |
|---------|---------|
| `make dev` | Run the API server in development mode |
| `make build` | Build `bin/api` and `bin/nexokit` |
| `make test` | Run all tests |
| `make init` | Adopt identity/configuration once, without contacting the database |
| `make migrate-up` | Apply reviewed migrations explicitly, after local initialization |
| `make migrate-down` | Rollback the last migration batch |
| `make migrate-create` | Create a new migration file |
| `make seed` | Run reviewed Go seeds explicitly, after migrations |
| `make create-root` | Create the root after seeds; skips an existing root without changing it |
| `make fmt` | Format all Go files |
| `make lint` | Run `go vet` + module boundary checks |

For direct `nexokit` CLI usage, see [`docs/references/runtime/cli.md`](docs/references/runtime/cli.md).

## Project map

```
cmd/api/              API server entrypoint
cmd/nexokit/          Internal developer CLI
internal/app/         Bootstrap and dependency graph
internal/config/      Typed configuration from .env
internal/infra/       DB, cache, logger adapters
internal/server/      HTTP server and router
internal/middleware/  Auth, tenant, rate limit, logging
internal/platform/    Cross-cutting contracts and helpers
internal/modules/     Business modules (companies, iam, onboarding)
internal/shared/      BaseModel types
migrations/           Goose SQL migrations
seeds/                Go seed files
scripts/              Hooks and helper scripts
tests/                Integration tests and helpers
```

IAM owns authentication under `internal/modules/iam/slices/auth`, including credentials, token lifecycle and session handling. Public authentication routes remain `/api/v1/auth/{login,refresh,logout,me}`.

## Using NexoKit as a starter

Use the interactive `make init` quick start above for local project initialization. For a safe,
stdin-free and write-free identity preview:

```bash
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api --dry-run
```

Bare `init` derives the module from display name and GitHub username; it creates
no GitHub repository and requests no GitHub credentials. Flag calls require both
`--name` and `--module`. Their default is preview; `--dry-run` is explicit preview
and cannot be combined with `--apply`. Flag `--apply` applies identity/environment
without stdin or a database connection: provide the local configuration process
inputs for a new/rebuilt environment, not root credentials. For an existing
environment, applying requires `--env-action keep` or `--env-action rebuild` explicitly.

Backup/rebuild retains a protected `.env.backup-<32 lowercase hex characters>`
(mode `0600`, ignored by a narrow Git pattern) and creates a new `0600` environment
from the example, preserving unknown content/comments with a fresh development
PASETO key and empty `DATABASE_URL`. Keep does not silently replace configuration.
Symlinks/nonregular destinations are refused.

Identity/module imports change; `cmd/nexokit` and original NexoKit tool provenance
remain. Your application is independently owned, with no downloader/upstream
update contract. Adopted or partially renamed projects are refused. Local defaults
are not production safeguards; no database provisioning or production-readiness
promise is made. Keep or evolve modules, migrations and vertical slices deliberately.

See [`docs/starter-template.md`](docs/starter-template.md) for the full adoption guide.

## Production path

```bash
make build
# copy binaries, .env, migrations/, seeds/ to the host
./bin/nexokit migrate up
./bin/nexokit seed
./bin/nexokit create-root
./bin/api
```

> **Seeding requires Go on the host today.** `nexokit seed` discovers Go files in `seeds/` and runs them with a temporary `go run` runner, so the host that performs seeding must have the Go toolchain available. The safe alternatives are to run `nexokit seed` from a build/admin environment that has Go and database access, or to keep Go installed on the production host that performs the one-time seed step. See [`docs/deployment.md`](docs/deployment.md) for details.

For environment variables, TLS, reverse proxy, logging, and operational checklists, see [`docs/deployment.md`](docs/deployment.md).

## Documentation

- [`docs/README.md`](docs/README.md) — documentation index
- [`docs/architecture_and_vsa.md`](docs/architecture_and_vsa.md) — canonical architecture guide
- [`docs/deployment.md`](docs/deployment.md) — production guide
- [`docs/references/runtime/cli.md`](docs/references/runtime/cli.md) — direct CLI reference
- [`docs/starter-template.md`](docs/starter-template.md) — adopt as a starter
- [`docs/references/runtime/request-flow.md`](docs/references/runtime/request-flow.md) — request/auth/tenant flow
- [`docs/system-map.md`](docs/system-map.md) — mounted modules and capabilities
