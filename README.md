# NexoKit

NexoKit is an opinionated modular Go starter for SaaS APIs. It ships with authentication, RBAC, multitenancy, migrations, a developer CLI, and a vertical-slice module structure so you can fork it and start building domain logic instead of boilerplate.

## Stack

| Layer | Choice |
|-------|--------|
| Language | Go 1.26+ |
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

```bash
cp .env.example .env
docker compose up -d
make migrate-up
make seed
make create-root
make dev
```

The API boots on `http://localhost:8080` (or `APP_PORT` from `.env`).

## Everyday commands

| Command | Purpose |
|---------|---------|
| `make dev` | Run the API server in development mode |
| `make build` | Build `bin/api` and `bin/nexokit` |
| `make test` | Run all tests |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Rollback the last migration batch |
| `make migrate-create` | Create a new migration file |
| `make seed` | Run seed files from `seeds/` |
| `make create-root` | Create the initial root user |
| `make fmt` | Format all Go files |
| `make lint` | Run `go vet` + module boundary checks |

For direct `nexokit` CLI usage, see [`docs/cli.md`](docs/cli.md).

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
internal/modules/     Business modules (auth, companies, iam, onboarding)
internal/shared/      BaseModel types
migrations/           Goose SQL migrations
seeds/                Go seed files
scripts/              Hooks and helper scripts
tests/                Integration tests and helpers
```

## Using NexoKit as a starter

NexoKit is designed to be cloned or forked as a project starting point:

1. Fork the repo and rename the Go module.
2. Keep the modules you need; delete or evolve the rest.
3. Add your own migrations, seeds, and vertical slices.
4. Deploy the binary with your environment variables.

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
- [`docs/architecture.md`](docs/architecture.md) — canonical architecture guide
- [`docs/deployment.md`](docs/deployment.md) — production guide
- [`docs/cli.md`](docs/cli.md) — direct CLI reference
- [`docs/starter-template.md`](docs/starter-template.md) — adopt as a starter
- [`docs/request-flow.md`](docs/request-flow.md) — request/auth/tenant flow
- [`docs/modules.md`](docs/modules.md) — module and vertical-slice guide
